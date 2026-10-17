package files

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/data"
)

// requestEditToken includes moderation and publication state as well as content,
// so a stale edit cannot silently undo another editor's work or decision.
func requestEditToken(ctx context.Context, db data.DBTX, requestID string) (string, error) {
	rows, err := db.QueryContext(ctx, `SELECT r.status,COALESCE(CAST(r.revision_of AS TEXT),''),COALESCE(r.reviewed_at,''),COALESCE(CAST(r.reviewed_by AS TEXT),''),a.id,a.group_id,a.edit_version,a.is_latest,a.module_id,a.version,a.exam_date,COALESCE(a.comment,''),a.checksum,a.nbytes,a.updated_at,p.module_id,p.version
		FROM archive_requests r JOIN archive a ON a.request_id=r.id AND a.deleted_at IS NULL JOIN archive_placements p ON p.archive_id=a.id WHERE r.id=? ORDER BY a.id,p.module_id,p.version`, requestID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	encoder := json.NewEncoder(hash)
	count := 0
	for rows.Next() {
		values := make([]string, len(columns))
		targets := make([]any, len(columns))
		for i := range values {
			targets[i] = &values[i]
		}
		if err := rows.Scan(targets...); err != nil {
			return "", err
		}
		if err := encoder.Encode(values); err != nil {
			return "", err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if count == 0 {
		return "", ErrNotFound
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// EditRequest updates a reviewed submission without transferring ownership,
// replacing its stable download IDs or changing its moderation state.
func (s *Service) EditRequest(ctx context.Context, userID int64, submission Submission, expected string) (result UploadResult, err error) {
	role, err := accountRole(ctx, s.db, userID)
	if err != nil {
		return result, err
	}
	if !canReview(role) {
		return result, ErrForbidden
	}
	if submission.RequestID == "" {
		return result, ErrNotFound
	}
	submission.RevisionOf = ""
	if fields := submission.Validate(); len(fields) > 0 {
		return result, fields
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "UPDATE archive_requests SET id=id WHERE id=?", submission.RequestID); err != nil {
		return result, err
	}
	role, err = accountRole(ctx, tx, userID)
	if err != nil {
		return result, err
	}
	if !canReview(role) {
		return result, ErrForbidden
	}
	var status Status
	if err := tx.QueryRowContext(ctx, "SELECT status FROM archive_requests WHERE id=?", submission.RequestID).Scan(&status); errors.Is(err, sql.ErrNoRows) {
		return result, ErrNotFound
	} else if err != nil {
		return result, err
	}
	if status != Pending && status != Approved && status != Rejected {
		return result, ErrConflict
	}
	current, err := requestEditToken(ctx, tx, submission.RequestID)
	if err != nil {
		return result, err
	}
	if expected == "" || current != expected {
		return result, ErrConflict
	}
	metadata, err := prepare(ctx, tx, userID, submission)
	if err != nil {
		return result, err
	}
	rows, err := tx.QueryContext(ctx, "SELECT id,is_latest FROM archive WHERE request_id=? AND deleted_at IS NULL ORDER BY id", submission.RequestID)
	if err != nil {
		return result, err
	}
	var ids []int64
	for rows.Next() {
		var id, latest int64
		if err := rows.Scan(&id, &latest); err != nil {
			rows.Close()
			return result, err
		}
		if status == Approved && latest == 0 {
			rows.Close()
			return result, ErrConflict
		}
		ids = append(ids, id)
	}
	err = errors.Join(rows.Err(), rows.Close())
	if err != nil {
		return result, err
	}
	if len(ids) != 1 {
		return result, ErrConflict
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
	var comment *string
	if value := strings.TrimSpace(submission.Comment); value != "" {
		comment = &value
	}
	if len(submission.Documents) > 0 {
		value := submission.Documents[0]
		checksum, err := storeObject(ctx, tx, storage, value, &attempts)
		if err != nil {
			return result, err
		}
		if _, err := tx.ExecContext(ctx, "UPDATE archive SET checksum=?,nbytes=?,mime_type='application/pdf' WHERE id=?", checksum, len(value), ids[0]); err != nil {
			return result, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE archive SET module_id=?,version=?,exam_date=?,comment=?,updated_at=strftime('%Y-%m-%dT%H:%M:%SZ','now') WHERE id=?`, metadata.moduleID, submission.placements()[0].Version, submission.Date, comment, ids[0]); err != nil {
		return result, err
	}
	if err := writeAssignments(ctx, tx, ids[0], submission.placements()); err != nil {
		return result, err
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
		return result, err
	}
	committed = true
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if err := s.CleanupBlobs(cleanup); err != nil {
		log.Printf("Remove replaced archive object: %v", err)
	}
	return UploadResult{RequestID: submission.RequestID, ModuleID: submission.placements()[0].ModuleID, IDs: []string{strconv.FormatInt(ids[0], 10)}}, nil
}
