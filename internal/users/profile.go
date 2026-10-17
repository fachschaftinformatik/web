package users

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
)

var ErrProfileNotFound = errors.New("public profile not found")

// PublicProfile contains only information a user has chosen to make public.
type PublicProfile struct {
	ID   int64
	Name string
}

func (s *Service) PublicProfile(ctx context.Context, value string) (PublicProfile, error) {
	var profile PublicProfile
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 || strconv.FormatInt(id, 10) != value {
		return profile, ErrProfileNotFound
	}
	err = s.db.QueryRowContext(ctx, `SELECT id,name FROM users WHERE id=? AND private=0 AND deleted_at IS NULL`, id).Scan(&profile.ID, &profile.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return profile, ErrProfileNotFound
	}
	return profile, err
}
