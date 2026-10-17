package archive

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/database"
)

func (s *Service) Revision(ctx context.Context, userID int64, value string) (Submission, error) {
	entryID, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return Submission{}, ErrNotFound
	}
	entry, err := canonicalEntry(ctx, s.db, entryID)
	if errors.Is(err, sql.ErrNoRows) {
		return Submission{}, ErrNotFound
	}
	if err != nil {
		return Submission{}, err
	}
	latest, err := database.New(s.db).GetLatestByGroupId(ctx, entry.GroupID)
	if errors.Is(err, sql.ErrNoRows) {
		return Submission{}, ErrNotFound
	}
	if err != nil {
		return Submission{}, err
	}
	submission := Submission{ModuleID: strconv.FormatInt(latest.ModuleID, 10), Version: latest.Version, Date: latest.ExamDate, RevisionOf: strconv.FormatInt(latest.ID, 10)}
	submission.Assignments, err = assignments(ctx, s.db, latest.ID)
	if err != nil {
		return Submission{}, err
	}
	if latest.Comment != nil {
		submission.Comment = *latest.Comment
	}
	if _, err := prepare(ctx, s.db, userID, submission); err != nil {
		return Submission{}, err
	}
	return submission, nil
}
