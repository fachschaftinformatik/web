package forum

import (
	"context"
	"database/sql"
	"errors"

	"github.com/fachschaftinformatik/web/internal/content"
)

type Draft struct {
	Title, Body, Type, Tags, Date, Time, Location string
}

type FieldError = content.FieldError

// Validate uses the existing application's title and body length policy.
func (d Draft) Validate() []FieldError {
	errors := content.ValidateText(d.Title, d.Body)
	switch d.Type {
	case "discussion", "news", "event":
	default:
		errors = append(errors, FieldError{"type", "Wähle eine Beitragsart aus."})
	}
	if d.Type == "event" {
		errors = append(errors, content.ValidateSchedule(d.Date, d.Time)...)
	}
	return errors
}

func editorialPermission(ctx context.Context, db querier, user int64) error {
	var role string
	err := db.QueryRowContext(ctx, `SELECT role FROM users WHERE id=? AND active=1 AND verified=1 AND deleted_at IS NULL AND role IN ('admin','editor')`, user).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	return err
}

func (s *Service) CanPublish(ctx context.Context, user int64) error {
	return editorialPermission(ctx, s.db, user)
}
