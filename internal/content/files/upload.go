package files

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/data"
	"github.com/fachschaftinformatik/web/internal/id"
)

const MaxUploadBytes = 32 << 20

var ErrForbidden = errors.New("Du hast keine Berechtigung, diese Reko zu ändern.")

type Submission struct {
	ModuleID, Version, Date, Comment, RevisionOf string
	RequestID                                    string
	Documents                                    [][]byte
	Assignments                                  []Assignment
}

type ValidationErrors map[string]string

func (e ValidationErrors) Error() string { return "Prüfe die Angaben im Formular." }

func (s Submission) Validate() ValidationErrors {
	result := ValidationErrors{}
	if len(s.Assignments) > 1024 {
		result["module_id"] = "Wähle höchstens 1024 Zuordnungen aus."
	}
	if len(s.Assignments) == 0 && s.ModuleID == "" {
		result["module_id"] = "Wähle ein Modul aus."
	}
	if len(s.Assignments) == 0 && s.Version == "" {
		result["version"] = "Wähle eine Prüfungsordnung aus."
	}
	date, err := time.Parse("2006-01-02", s.Date)
	if err != nil || date.Year() < 1900 || date.Year() > 2100 {
		result["date"] = "Gib ein gültiges Prüfungsdatum zwischen 1900 und 2100 ein."
	}
	if len(s.Documents) == 0 && s.RequestID == "" {
		result["files"] = "Wähle eine Datei aus."
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
		result["files"] = "Wähle genau eine Datei aus."
	}
	if len(s.Documents) > 1 {
		result["files"] = "Wähle genau eine Datei aus."
	}
	if s.RevisionOf != "" && len(s.Documents) != 1 && !(s.RequestID != "" && len(s.Documents) == 0) {
		result["files"] = "Wähle genau eine Datei für die neue Revision aus."
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
	selected := submission.placements()
	for _, a := range selected {
		var valid int
		err := db.QueryRowContext(ctx, `SELECT 1 FROM modules m JOIN programs p ON p.id=m.program_id AND p.deleted_at IS NULL JOIN program_versions v ON v.program_id=p.id AND v.name=? WHERE m.id=? AND m.deleted_at IS NULL`, a.Version, a.ModuleID).Scan(&valid)
		if errors.Is(err, sql.ErrNoRows) {
			return prepared{}, ValidationErrors{"module_id": "Wähle gültige Module und dazugehörige Prüfungsordnungen aus."}
		}
		if err != nil {
			return prepared{}, err
		}
	}
	moduleID, err := strconv.ParseInt(selected[0].ModuleID, 10, 64)
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
	err = db.QueryRowContext(ctx, `SELECT 1 FROM program_versions WHERE program_id=? AND name=?`, programID, selected[0].Version).Scan(&exists)
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
	parent, err := canonicalEntry(ctx, db, parentID)
	if errors.Is(err, sql.ErrNoRows) {
		return prepared{}, ErrNotFound
	}
	if err != nil {
		return prepared{}, err
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "UPDATE archive_requests SET id=id WHERE id=?", submission.RequestID); err != nil {
		return result, err
	}
	metadata, err := prepare(ctx, tx, userID, submission)
	if err != nil {
		return result, err
	}
	var attempts []string
	committed := false
	defer func() {
		if committed {
			return
		}
		tx.Rollback()
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		for _, key := range attempts {
			if cleanupErr := storage.Delete(cleanup, key); cleanupErr != nil {
				err = errors.Join(err, fmt.Errorf("remove uncommitted object: %w", cleanupErr))
				_, queueErr := s.db.ExecContext(cleanup, "INSERT OR IGNORE INTO archive_blob_cleanup(blob_key) VALUES(?)", key)
				err = errors.Join(err, queueErr)
			}
		}
	}()
	requestID := submission.RequestID
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
		if _, err := tx.ExecContext(ctx, `UPDATE archive SET module_id=?,version=?,exam_date=?,comment=? WHERE request_id=?`, metadata.moduleID, submission.placements()[0].Version, submission.Date, comment, requestID); err != nil {
			return result, err
		}
	}
	seen := map[string]bool{}
	for _, document := range submission.Documents {
		checksum, err := storeObject(ctx, tx, storage, document, &attempts)
		if err != nil {
			return result, err
		}
		if seen[checksum] {
			continue
		}
		seen[checksum] = true
		entryID := int64(id.New())
		groupID := metadata.groupID
		if groupID == 0 {
			groupID = entryID
		}
		_, err = data.New(tx).CreateArchiveEntry(ctx, data.CreateArchiveEntryParams{
			ID: entryID, GroupID: groupID, UserID: userID, ModuleID: metadata.moduleID,
			Version: submission.placements()[0].Version, ExamDate: submission.Date, Comment: comment,
			EditVersion: metadata.editVersion, IsLatest: 0, AccessKey: rand.Text(),
			MimeType: "application/pdf", Nbytes: int64(len(document)), Checksum: checksum,
		})
		if err != nil {
			return result, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE archive SET request_id=? WHERE id=?`, requestID, entryID); err != nil {
			return result, err
		}
		result.IDs = append(result.IDs, strconv.FormatInt(entryID, 10))
	}
	rows, err := tx.QueryContext(ctx, "SELECT id FROM archive WHERE request_id=? ORDER BY id", requestID)
	if err != nil {
		return result, err
	}
	var entries []int64
	for rows.Next() {
		var entryID int64
		if err := rows.Scan(&entryID); err != nil {
			rows.Close()
			return result, err
		}
		entries = append(entries, entryID)
	}
	err = errors.Join(rows.Err(), rows.Close())
	if err != nil {
		return result, err
	}
	for _, entryID := range entries {
		if err := writeAssignments(ctx, tx, entryID, submission.placements()); err != nil {
			return result, err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO archive_blob_cleanup(blob_key) SELECT blob_key FROM archive_objects o WHERE NOT EXISTS(SELECT 1 FROM archive a WHERE a.checksum=o.checksum)`); err != nil {
		return result, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO archive_blob_cleanup(blob_key) SELECT c.blob_key FROM archive_legacy_copies c WHERE NOT EXISTS(SELECT 1 FROM archive a WHERE a.checksum=c.checksum)`); err != nil {
		return result, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM archive_objects WHERE NOT EXISTS(SELECT 1 FROM archive a WHERE a.checksum=archive_objects.checksum)`); err != nil {
		return result, err
	}
	if err := tx.Commit(); err != nil {
		return UploadResult{}, err
	}
	committed = true
	result.ModuleID = submission.placements()[0].ModuleID
	result.RequestID = requestID
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if err := s.CleanupBlobs(cleanup); err != nil {
		log.Printf("Remove replaced draft object: %v", err)
	}
	return result, nil
}
