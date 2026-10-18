package exams

import (
	"context"
	"database/sql"
	"errors"

	"github.com/fachschaftinformatik/web/internal/database"
)

func (s *Service) Revision(ctx context.Context, userID int64, value string) (SubmissionDraft, error) {
	entryID, err := resolveID(ctx, s.db, "documents", value)
	if err != nil {
		return SubmissionDraft{}, ErrNotFound
	}
	entry, err := canonicalEntry(ctx, s.db, entryID)
	if errors.Is(err, sql.ErrNoRows) {
		return SubmissionDraft{}, ErrNotFound
	}
	if err != nil {
		return SubmissionDraft{}, err
	}
	latest, err := database.New(s.db).GetLatestByGroupId(ctx, entry.GroupID)
	if errors.Is(err, sql.ErrNoRows) {
		return SubmissionDraft{}, ErrNotFound
	}
	if err != nil {
		return SubmissionDraft{}, err
	}
	moduleID, err := lookupID(ctx, s.db, "modules", latest.ModuleID)
	if err != nil {
		return SubmissionDraft{}, err
	}
	documentID, err := lookupID(ctx, s.db, "documents", latest.ID)
	if err != nil {
		return SubmissionDraft{}, err
	}
	submission := SubmissionDraft{ModuleID: moduleID, Version: latest.Version, Date: latest.ExamDate, RevisionOf: documentID}
	submission.Assignments, err = assignments(ctx, s.db, latest.ID)
	if err != nil {
		return SubmissionDraft{}, err
	}
	if latest.Comment != nil {
		submission.Comment = *latest.Comment
	}
	if _, err := prepare(ctx, s.db, userID, submission); err != nil {
		return SubmissionDraft{}, err
	}
	return submission, nil
}
