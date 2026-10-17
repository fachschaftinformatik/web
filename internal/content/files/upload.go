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
	"log"
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
	RequestID                                    string
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
	if len(s.Documents) == 0 && s.RequestID == "" {
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
	if s.RevisionOf != "" && len(s.Documents) != 1 && !(s.RequestID != "" && len(s.Documents) == 0) {
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
	ModuleID  string
	RequestID string
	IDs       []string
}

type prepared struct {
	moduleID, groupID, editVersion int64
	baseID                         int64
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
	var status string
	err = db.QueryRowContext(ctx, `SELECT COALESCE(r.status,'approved') FROM archive a LEFT JOIN archive_requests r ON r.id=a.request_id WHERE a.id=?`, parent.ID).Scan(&status)
	if err != nil {
		return prepared{}, err
	}
	if status != "approved" {
		return prepared{}, ErrNotFound
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
	latest, err := data.New(db).GetLatestByGroupId(ctx, parent.GroupID)
	if err != nil {
		return prepared{}, err
	}
	result.baseID = latest.ID
	return result, nil
}

// SaveDraft stores the uploads without changing the public archive.
func (s *Service) SaveDraft(ctx context.Context, userID int64, submission Submission) (result UploadResult, err error) {
	if fields := submission.Validate(); len(fields) != 0 {
		return result, fields
	}
	if _, err := prepare(ctx, s.db, userID, submission); err != nil {
		return result, err
	}
	if submission.RequestID != "" {
		request, err := s.Request(ctx, userID, submission.RequestID, false)
		if err != nil {
			return result, err
		}
		if request.Status != Draft {
			return result, ErrConflict
		}
		if request.RevisionOf != submission.RevisionOf {
			return result, ErrConflict
		}
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
	requestID := submission.RequestID
	var obsolete []string
	if requestID == "" {
		requestID = rand.Text()
		var base any
		if metadata.baseID != 0 {
			base = metadata.baseID
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO archive_requests (id,user_id,revision_of) VALUES (?,?,?)`, requestID, userID, base); err != nil {
			return result, err
		}
	} else {
		var owner int64
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT user_id,status FROM archive_requests WHERE id=?`, requestID).Scan(&owner, &status); err != nil {
			return result, err
		}
		if owner != userID || status != "draft" {
			return result, ErrConflict
		}
		if len(submission.Documents) != 0 {
			rows, err := tx.QueryContext(ctx, `SELECT id FROM archive WHERE request_id=?`, requestID)
			if err != nil {
				return result, err
			}
			for rows.Next() {
				var value int64
				if err := rows.Scan(&value); err != nil {
					rows.Close()
					return result, err
				}
				obsolete = append(obsolete, blob.ArchiveSourceKey(strconv.FormatInt(value, 10)))
			}
			err = rows.Err()
			rows.Close()
			if err != nil {
				return result, err
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM archive WHERE request_id=?`, requestID); err != nil {
				return result, err
			}
		}
	}
	var comment *string
	if value := strings.TrimSpace(submission.Comment); value != "" {
		comment = &value
	}
	if len(submission.Documents) == 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE archive SET module_id=?,version=?,exam_date=?,comment=? WHERE request_id=?`, metadata.moduleID, submission.Version, submission.Date, comment, requestID); err != nil {
			return result, err
		}
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
			EditVersion: metadata.editVersion, IsLatest: 0, AccessKey: rand.Text(),
			MimeType: "application/pdf", Nbytes: int64(len(document)), Checksum: hex.EncodeToString(checksum[:]),
		})
		if err != nil {
			return result, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE archive SET request_id=? WHERE id=?`, requestID, ids[i]); err != nil {
			return result, err
		}
		result.IDs = append(result.IDs, strconv.FormatInt(ids[i], 10))
	}
	if err := tx.Commit(); err != nil {
		return UploadResult{}, err
	}
	committed = true
	result.ModuleID = submission.ModuleID
	result.RequestID = requestID
	for _, key := range obsolete {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		cleanupErr := storage.Delete(cleanup, key)
		cancel()
		if cleanupErr != nil {
			log.Printf("Remove replaced draft object: %v", cleanupErr)
		}
	}
	return result, nil
}
