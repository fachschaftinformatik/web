package users

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/data"
	"github.com/fachschaftinformatik/web/internal/id"
)

// Invite uses the caller's transaction so the membership and queued mail commit
// together. Existing passwords, verification and activation are preserved.
func (s *Service) Invite(ctx context.Context, tx *sql.Tx, name, email string) (data.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if EmailError(email) != "" {
		return data.User{}, ErrCredentials
	}
	q := data.New(tx)
	user, err := q.GetUserByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		var existing int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE lower(email)=?`, email).Scan(&existing); err != nil {
			return data.User{}, err
		}
		if existing != 0 {
			return data.User{}, ErrEmailExists
		}
		user, err = q.CreateUser(ctx, data.CreateUserParams{ID: int64(id.New()), Email: email, Name: strings.TrimSpace(name), Password: "!", Role: "editor", Active: 1})
	} else if err == nil && user.Role == "user" {
		user, err = q.SetUserRole(ctx, data.SetUserRoleParams{ID: user.ID, Role: "editor"})
	}
	if err != nil {
		return data.User{}, err
	}
	if user.Verified != 1 {
		if err := s.queueLink(ctx, tx, user.ID, email, name, "verify", "/invite"); err != nil {
			return data.User{}, err
		}
	}
	return user, nil
}

type Invitation struct{ Name, Email string }

func (s *Service) Invitation(ctx context.Context, value string) (Invitation, error) {
	if len(value) != 64 {
		return Invitation{}, ErrToken
	}
	var result Invitation
	err := s.db.QueryRowContext(ctx, `SELECT u.name,u.email FROM user_tokens t JOIN users u ON u.id=t.user_id JOIN people p ON p.user_id=u.id WHERE t.id=? AND t.kind='verify' AND t.expires_at>? AND u.verified=0 AND u.deleted_at IS NULL`, digest(value), time.Now().UTC().Format(time.RFC3339)).Scan(&result.Name, &result.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return Invitation{}, ErrToken
	}
	result.Email = strings.ToLower(result.Email)
	return result, err
}

func (s *Service) AcceptInvitation(ctx context.Context, value, password string) error {
	if _, err := s.Invitation(ctx, value); err != nil {
		return err
	}
	message := PasswordError(password)
	if message != "" {
		return errors.New(message)
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var userID int64
	err = tx.QueryRowContext(ctx, `DELETE FROM user_tokens WHERE id=? AND kind='verify' AND expires_at>? AND user_id IN (SELECT u.id FROM users u JOIN people p ON p.user_id=u.id WHERE u.verified=0 AND u.deleted_at IS NULL) RETURNING user_id`, digest(value), time.Now().UTC().Format(time.RFC3339)).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrToken
	}
	if err != nil {
		return err
	}
	if _, err := data.New(tx).UpdateUserPassword(ctx, data.UpdateUserPasswordParams{ID: userID, Password: hash}); err != nil {
		return err
	}
	if _, err := data.New(tx).VerifyUser(ctx, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM web_sessions WHERE user_id=?`, userID); err != nil {
		return err
	}
	if err := data.New(tx).DeleteUserSessions(ctx, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM mail_outbox WHERE user_id=? AND kind='verify'`, userID); err != nil {
		return err
	}
	return tx.Commit()
}
