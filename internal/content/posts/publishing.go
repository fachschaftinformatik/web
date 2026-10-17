package posts

import (
	"context"
	"database/sql"
	"errors"
)

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
