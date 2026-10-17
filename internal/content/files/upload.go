package files

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/data"
	"github.com/fachschaftinformatik/web/internal/data/blob"
	"github.com/fachschaftinformatik/web/internal/id"
)

const MaxUploadBytes = 32 << 20

var ErrForbidden = errors.New("Du hast keine Berechtigung, diese Reko zu ändern.")

type Submission struct {
	ModuleID, Version, Date, Comment, RevisionOf string
	Documents                                    [][]byte
}

type ValidationErrors map[string]string

func (e ValidationErrors) Error() string { return "Prüfe die Angaben im Formular." }

func (s Submission) Validate() ValidationErrors {
	result := ValidationErrors{}
	if s.ModuleID == "" {
		result["module_id"] = "Wähle ein Modul aus."
	}
	if s.Version == "" {
		result["version"] = "Wähle eine Prüfungsordnung aus."
	}
	date, err := time.Parse("2006-01-02", s.Date)
	if err != nil || date.Year() < 1900 || date.Year() > 2100 {
		result["date"] = "Gib ein gültiges Prüfungsdatum zwischen 1900 und 2100 ein."
	}
	if len(s.Documents) == 0 {
		result["files"] = "Wähle mindestens eine PDF-Datei aus."
	}
	total := 0
	for _, document := range s.Documents {
		if message := content.PDFError(document); message != "" {
			result["files"] = message
			break
		}
		total += len(document)
	}
	if total > MaxUploadBytes {
		result["files"] = "Die PDF-Dateien dürfen zusammen höchstens 32 MB groß sein."
	}
	if s.RevisionOf != "" && len(s.Documents) != 1 {
		result["files"] = "Wähle genau eine PDF-Datei für die neue Revision aus."
	}
	return result
}

type writableStorage interface {
	Storage
	Upload(context.Context, string, io.Reader, int64, string) error
	Delete(context.Context, string) error
}

type UploadResult struct {
	ModuleID string
	IDs      []string
}

type prepared struct {
	moduleID, groupID, editVersion int64
}

// prepare rechecks account policy and revision ownership inside the write
// transaction as well as before contacting object storage.
func prepare(ctx context.Context, db data.DBTX, userID int64, submission Submission) (prepared, error) {
	var role string
	err := db.QueryRowContext(ctx, `SELECT role FROM users WHERE id=? AND active=1 AND verified=1 AND deleted_at IS NULL`, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return prepared{}, ErrForbidden
	}
	if err != nil {
		return prepared{}, err
	}
	moduleID, err := strconv.ParseInt(submission.ModuleID, 10, 64)
	if err != nil {
		return prepared{}, ValidationErrors{"module_id": "Wähle ein gültiges Modul aus."}
	}
	var programID int64
	err = db.QueryRowContext(ctx, `SELECT m.program_id FROM modules m JOIN programs p ON p.id=m.program_id AND p.deleted_at IS NULL WHERE m.id=? AND m.deleted_at IS NULL`, moduleID).Scan(&programID)
	if errors.Is(err, sql.ErrNoRows) {
		return prepared{}, ValidationErrors{"module_id": "Wähle ein gültiges Modul aus."}
	}
	if err != nil {
		return prepared{}, err
	}
	var exists int
	err = db.QueryRowContext(ctx, `SELECT 1 FROM program_versions WHERE program_id=? AND name=?`, programID, submission.Version).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return prepared{}, ValidationErrors{"version": "Wähle eine Prüfungsordnung aus, die zu diesem Modul gehört."}
	}
	if err != nil {
		return prepared{}, err
	}
	result := prepared{moduleID: moduleID, editVersion: 1}
	if submission.RevisionOf == "" {
		return result, nil
	}
	parentID, err := strconv.ParseInt(submission.RevisionOf, 10, 64)
	if err != nil {
		return prepared{}, ErrNotFound
	}
	parent, err := data.New(db).GetArchiveEntry(ctx, parentID)
	if errors.Is(err, sql.ErrNoRows) {
		return prepared{}, ErrNotFound
	}
	if err != nil {
		return prepared{}, err
	}
	if parent.ModuleID != moduleID {
		return prepared{}, ValidationErrors{"module_id": "Eine Revision muss zum bisherigen Modul gehören."}
	}
	var ownerID, latestVersion int64
	err = db.QueryRowContext(ctx, `SELECT user_id FROM archive WHERE group_id=? ORDER BY edit_version,id LIMIT 1`, parent.GroupID).Scan(&ownerID)
	if err != nil {
		return prepared{}, err
	}
	if role != "admin" && role != "editor" && userID != ownerID {
		return prepared{}, ErrForbidden
	}
	err = db.QueryRowContext(ctx, `SELECT MAX(edit_version) FROM archive WHERE group_id=?`, parent.GroupID).Scan(&latestVersion)
	if err != nil {
		return prepared{}, err
	}
	result.groupID, result.editVersion = parent.GroupID, latestVersion+1
	return result, nil
}

// Submit publishes independent uploads or one explicit revision. Old revision
// flags are changed only after every object has been stored successfully.
func (s *Service) Submit(ctx context.Context, userID int64, submission Submission) (result UploadResult, err error) {
	if fields := submission.Validate(); len(fields) != 0 {
		return result, fields
	}
	if _, err := prepare(ctx, s.db, userID, submission); err != nil {
		return result, err
	}
	storage, ok := s.storage.(writableStorage)
	if !ok {
		return result, ErrStorage
	}
	var attempts []string
	committed := false
	defer func() {
		if committed {
			return
		}
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		for _, key := range attempts {
			if cleanupErr := storage.Delete(cleanup, key); cleanupErr != nil {
				err = errors.Join(err, fmt.Errorf("remove uncommitted object: %w", cleanupErr))
			}
		}
	}()
	ids := make([]int64, len(submission.Documents))
	for i, document := range submission.Documents {
		ids[i] = int64(id.New())
		key := blob.ArchiveSourceKey(strconv.FormatInt(ids[i], 10))
		attempts = append(attempts, key)
		if err := storage.Upload(ctx, key, bytes.NewReader(document), int64(len(document)), "application/pdf"); err != nil {
			return result, fmt.Errorf("%w: %v", ErrStorage, err)
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	metadata, err := prepare(ctx, tx, userID, submission)
	if err != nil {
		return result, err
	}
	if metadata.groupID != 0 {
		if err := data.New(tx).ClearLatestFlagByGroupId(ctx, metadata.groupID); err != nil {
			return result, err
		}
	}
	var comment *string
	if value := strings.TrimSpace(submission.Comment); value != "" {
		comment = &value
	}
	for i, document := range submission.Documents {
		groupID := metadata.groupID
		if groupID == 0 {
			groupID = ids[i]
		}
		checksum := sha256.Sum256(document)
		_, err := data.New(tx).CreateArchiveEntry(ctx, data.CreateArchiveEntryParams{
			ID: ids[i], GroupID: groupID, UserID: userID, ModuleID: metadata.moduleID,
			Version: submission.Version, ExamDate: submission.Date, Comment: comment,
			EditVersion: metadata.editVersion, IsLatest: 1, AccessKey: rand.Text(),
			MimeType: "application/pdf", Nbytes: int64(len(document)), Checksum: hex.EncodeToString(checksum[:]),
		})
		if err != nil {
			return result, err
		}
		result.IDs = append(result.IDs, strconv.FormatInt(ids[i], 10))
	}
	if err := tx.Commit(); err != nil {
		return UploadResult{}, err
	}
	committed = true
	result.ModuleID = submission.ModuleID
	return result, nil
}
