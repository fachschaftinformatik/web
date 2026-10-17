package users

import (
	"context"

	"github.com/fachschaftinformatik/web/internal/data"
)

// ResendConfirmation replaces the outstanding link without changing the account.
// Members complete their invitation; other accounts verify their email address.
func (s *Service) ResendConfirmation(ctx context.Context, userID int64) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	user, err := data.New(tx).GetUser(ctx, userID)
	if err != nil {
		return false, err
	}
	if user.Verified == 1 {
		return false, nil
	}
	var member bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM people WHERE user_id=?)`, userID).Scan(&member); err != nil {
		return false, err
	}
	path := "/verify"
	if member {
		path = "/invite"
	}
	if err := s.queueLink(ctx, tx, user.ID, user.Email, user.Name, "verify", path); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
