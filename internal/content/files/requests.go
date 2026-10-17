package files

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/data"
)

type Status string

const (
	Draft    Status = "draft"
	Pending  Status = "pending"
	Approved Status = "approved"
	Rejected Status = "rejected"
)

func (s Status) Label() string {
	switch s {
	case Draft:
		return "Entwurf"
	case Pending:
		return "Eingereicht"
	case Approved:
		return "Bestätigt"
	case Rejected:
		return "Abgelehnt"
	}
	return ""
}

var ErrConflict = errors.New("Der Antrag wurde bereits geändert. Lade die Seite neu und prüfe den aktuellen Stand.")

type Request struct {
	ID, Name, Title, Program, ModuleID, Version, Date, Comment, RevisionOf string
	UserID                                                                 int64
	Status                                                                 Status
	CreatedAt                                                              string
	SubmittedAt, ReviewedAt                                                *string
	Documents                                                              []Exam
	Count                                                                  int
	Assignments                                                            []Assignment
	SharedAssignments                                                      bool
	EditToken                                                              string
}

func (r Request) DateLabel() string {
	if date, err := parseDate(r.Date); err == nil {
		return date.Format("02.01.2006")
	}
	return r.Date
}

func (r Request) SubmittedLabel() string {
	value := r.CreatedAt
	if r.SubmittedAt != nil {
		value = *r.SubmittedAt
	}
	if date, err := time.Parse(time.RFC3339, value); err == nil {
		return date.Format("02.01.2006")
	}
	return value
}

func accountRole(ctx context.Context, db data.DBTX, userID int64) (string, error) {
	var role string
	err := db.QueryRowContext(ctx, `SELECT role FROM users WHERE id=? AND active=1 AND verified=1 AND deleted_at IS NULL`, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrForbidden
	}
	return role, err
}

func canReview(role string) bool { return role == "admin" || role == "editor" }

func (s *Service) Request(ctx context.Context, userID int64, value string, reviewer bool) (Request, error) {
	role, err := accountRole(ctx, s.db, userID)
	if err != nil {
		return Request{}, err
	}
	if reviewer && !canReview(role) {
		return Request{}, ErrForbidden
	}
	var result Request
	var base *int64
	err = s.db.QueryRowContext(ctx, `SELECT r.id,r.user_id,r.status,r.revision_of,r.created_at,r.submitted_at,r.reviewed_at,u.name FROM archive_requests r JOIN users u ON u.id=r.user_id WHERE r.id=?`, value).Scan(&result.ID, &result.UserID, &result.Status, &base, &result.CreatedAt, &result.SubmittedAt, &result.ReviewedAt, &result.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return Request{}, ErrNotFound
	}
	if err != nil {
		return Request{}, err
	}
	if (reviewer && result.Status == Draft) || (!reviewer && result.UserID != userID) {
		return Request{}, ErrNotFound
	}
	if base != nil {
		result.RevisionOf = strconv.FormatInt(*base, 10)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT a.id,a.module_id,m.name,p.name,a.version,a.exam_date,COALESCE(a.comment,''),a.nbytes FROM archive a JOIN modules m ON m.id=a.module_id JOIN programs p ON p.id=m.program_id WHERE a.request_id=? AND a.deleted_at IS NULL ORDER BY a.id`, value)
	if err != nil {
		return Request{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, moduleID int64
		var size int64
		if err := rows.Scan(&id, &moduleID, &result.Title, &result.Program, &result.Version, &result.Date, &result.Comment, &size); err != nil {
			return Request{}, err
		}
		result.ModuleID = strconv.FormatInt(moduleID, 10)
		result.Documents = append(result.Documents, Exam{ID: strconv.FormatInt(id, 10), ModuleID: result.ModuleID, Title: result.Title, Date: result.DateLabel(), Bytes: size, Number: len(result.Documents) + 1})
	}
	if err := rows.Err(); err != nil {
		return Request{}, err
	}
	rows.Close()
	if len(result.Documents) == 0 {
		return Request{}, ErrNotFound
	}
	result.Count = len(result.Documents)
	result.SharedAssignments = true
	for i := range result.Documents {
		entryID, _ := strconv.ParseInt(result.Documents[i].ID, 10, 64)
		selected, err := assignments(ctx, s.db, entryID)
		if err != nil {
			return Request{}, err
		}
		result.Documents[i].Assignments = selected
		if i == 0 {
			result.Assignments = selected
			continue
		}
		if !slices.EqualFunc(result.Assignments, selected, func(a, b Assignment) bool { return a.ModuleID == b.ModuleID && a.Version == b.Version }) {
			result.SharedAssignments = false
		}
	}
	var titles []string
	for _, a := range result.Assignments {
		if !slices.Contains(titles, a.Title) {
			titles = append(titles, a.Title)
		}
	}
	result.Title = strings.Join(titles, ", ")
	if reviewer {
		result.EditToken, err = requestEditToken(ctx, s.db, value)
		if err != nil {
			return Request{}, err
		}
	}
	return result, nil
}

func (s *Service) Requests(ctx context.Context, userID int64, reviewer bool) ([]Request, error) {
	role, err := accountRole(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}
	if reviewer && !canReview(role) {
		return nil, ErrForbidden
	}
	rows, err := s.db.QueryContext(ctx, `SELECT r.id,r.user_id,r.status,r.created_at,r.submitted_at,u.name,GROUP_CONCAT(DISTINCT m.name),COUNT(DISTINCT a.id) FROM archive_requests r JOIN users u ON u.id=r.user_id JOIN archive a ON a.request_id=r.id AND a.deleted_at IS NULL JOIN archive_placements b ON b.archive_id=a.id JOIN modules m ON m.id=b.module_id WHERE (? AND r.status!='draft') OR (NOT ? AND r.user_id=?) GROUP BY r.id ORDER BY CASE r.status WHEN 'pending' THEN 0 ELSE 1 END,r.created_at DESC,r.id DESC`, reviewer, reviewer, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Request
	for rows.Next() {
		var request Request
		if err := rows.Scan(&request.ID, &request.UserID, &request.Status, &request.CreatedAt, &request.SubmittedAt, &request.Name, &request.Title, &request.Count); err != nil {
			return nil, err
		}
		result = append(result, request)
		result[len(result)-1].Title = strings.ReplaceAll(request.Title, ",", ", ")
	}
	return result, rows.Err()
}

func (s *Service) SendRequest(ctx context.Context, userID int64, value string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := accountRole(ctx, tx, userID); err != nil {
		return err
	}
	var status Status
	err = tx.QueryRowContext(ctx, `SELECT status FROM archive_requests WHERE id=? AND user_id=?`, value, userID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status == Pending {
		return nil
	}
	if status != Draft {
		return ErrConflict
	}
	if _, err := tx.ExecContext(ctx, `UPDATE archive_requests SET status='pending',submitted_at=strftime('%Y-%m-%dT%H:%M:%SZ','now') WHERE id=?`, value); err != nil {
		return err
	}
	return tx.Commit()
}

// Review updates the request and public revision flags in one transaction.
func (s *Service) Review(ctx context.Context, userID int64, value string, approve bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	role, err := accountRole(ctx, tx, userID)
	if err != nil {
		return err
	}
	if !canReview(role) {
		return ErrForbidden
	}
	var status Status
	var base *int64
	err = tx.QueryRowContext(ctx, `SELECT status,revision_of FROM archive_requests WHERE id=?`, value).Scan(&status, &base)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	target := Rejected
	if approve {
		target = Approved
	}
	if status == target {
		return nil
	}
	if status != Pending && status != Approved && status != Rejected {
		return ErrConflict
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM archive WHERE request_id=? AND deleted_at IS NULL ORDER BY id`, value)
	if err != nil {
		return err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return ErrNotFound
	}
	if approve {
		for _, id := range ids {
			entry, err := data.New(tx).GetArchiveEntry(ctx, id)
			if err != nil {
				return err
			}
			var newerHead bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM archive WHERE group_id=? AND is_latest=1 AND deleted_at IS NULL AND edit_version>?)`, entry.GroupID, entry.EditVersion).Scan(&newerHead); err != nil {
				return err
			}
			if newerHead {
				return ErrConflict
			}
			// A repeated upload can add applicability to an existing document,
			// but only after review. Record this as a new applicability revision
			// so pending revisions against the old head become stale normally.
			if base == nil && entry.GroupID == entry.ID {
				var existingID, groupID, editVersion int64
				err := tx.QueryRowContext(ctx, `SELECT a.id,a.group_id,(SELECT MAX(edit_version)+1 FROM archive h WHERE h.group_id=a.group_id) FROM archive a LEFT JOIN archive_requests r ON r.id=a.request_id
				WHERE a.id!=? AND a.checksum=? AND a.nbytes=? AND (SELECT original.user_id FROM archive original WHERE original.group_id=a.group_id ORDER BY original.edit_version,original.id LIMIT 1)=? AND a.exam_date=? AND COALESCE(a.comment,'')=COALESCE(?,'') AND a.is_latest=1 AND a.deleted_at IS NULL AND (r.id IS NULL OR r.status='approved')
				AND NOT EXISTS(SELECT 1 FROM archive_entry_aliases x WHERE x.legacy_id=a.id AND x.canonical_id!=a.id) ORDER BY a.id LIMIT 1`, entry.ID, entry.Checksum, entry.Nbytes, entry.UserID, entry.ExamDate, entry.Comment).Scan(&existingID, &groupID, &editVersion)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					return err
				}
				if err == nil {
					if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO archive_assignments(archive_id,module_id,version) SELECT ?,module_id,version FROM archive_placements WHERE archive_id=?`, id, existingID); err != nil {
						return err
					}
					if _, err := tx.ExecContext(ctx, "INSERT OR REPLACE INTO archive_group_aliases(group_id,canonical_group_id) VALUES(?,?)", entry.GroupID, groupID); err != nil {
						return err
					}
					if _, err := tx.ExecContext(ctx, "UPDATE archive SET group_id=?,edit_version=? WHERE id=?", groupID, editVersion, id); err != nil {
						return err
					}
					entry.GroupID = groupID
				}
			}
			if base != nil {
				latest, err := data.New(tx).GetLatestByGroupId(ctx, entry.GroupID)
				if errors.Is(err, sql.ErrNoRows) || err == nil && latest.ID != *base {
					return ErrConflict
				}
				if err != nil {
					return err
				}
			}
			var exists int
			err = tx.QueryRowContext(ctx, `SELECT 1 FROM modules m JOIN programs p ON p.id=m.program_id WHERE m.id=? AND m.deleted_at IS NULL AND p.deleted_at IS NULL`, entry.ModuleID).Scan(&exists)
			if errors.Is(err, sql.ErrNoRows) {
				return ErrConflict
			}
			if err != nil {
				return err
			}
			selected, err := assignments(ctx, tx, id)
			if err != nil {
				return err
			}
			if len(selected) == 0 {
				return ErrConflict
			}
			for _, a := range selected {
				var valid int
				err := tx.QueryRowContext(ctx, `SELECT 1 FROM modules m JOIN programs p ON p.id=m.program_id AND p.deleted_at IS NULL JOIN program_versions v ON v.program_id=p.id AND v.name=? WHERE m.id=? AND m.deleted_at IS NULL`, a.Version, a.ModuleID).Scan(&valid)
				if errors.Is(err, sql.ErrNoRows) {
					return ErrConflict
				}
				if err != nil {
					return err
				}
			}
			if err := data.New(tx).ClearLatestFlagByGroupId(ctx, entry.GroupID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE archive SET is_latest=0 WHERE group_id IN (SELECT group_id FROM archive_group_aliases WHERE canonical_group_id=?)`, entry.GroupID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE archive SET is_latest=1 WHERE id=?`, id); err != nil {
				return err
			}
		}
	} else if status == Approved {
		for _, id := range ids {
			entry, err := data.New(tx).GetArchiveEntry(ctx, id)
			if err != nil {
				return err
			}
			// A decision about an older revision must not displace a newer head.
			if entry.IsLatest == 0 {
				continue
			}
			if _, err := tx.ExecContext(ctx, `UPDATE archive SET is_latest=0 WHERE id=?`, id); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE archive SET is_latest=1 WHERE id=(
				SELECT a.id FROM archive a LEFT JOIN archive_requests r ON r.id=a.request_id
				WHERE a.group_id=? AND a.deleted_at IS NULL AND a.edit_version<?
				AND (r.id IS NULL OR r.status='approved') AND (a.request_id IS NULL OR a.request_id!=?)
				AND NOT EXISTS(SELECT 1 FROM archive_entry_aliases x WHERE x.legacy_id=a.id AND x.canonical_id!=a.id)
				ORDER BY a.edit_version DESC,a.id DESC LIMIT 1)`, entry.GroupID, entry.EditVersion, value); err != nil {
				return err
			}
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE archive_requests SET status=?,reviewed_at=strftime('%Y-%m-%dT%H:%M:%SZ','now'),reviewed_by=? WHERE id=?`, target, userID, value); err != nil {
		return err
	}
	return tx.Commit()
}
