package users

import (
	"context"
	"database/sql"
	"errors"
	"github.com/fachschaftinformatik/web/internal/datetime"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/publicid"
	"github.com/fachschaftinformatik/web/internal/urls"
)

// Invite uses the caller's transaction so the membership and queued mail commit
// together. Existing passwords, verification and activation are preserved.
func (s *Service) Invite(ctx context.Context, tx *sql.Tx, name, email string) (database.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if s.EmailError(email) != "" {
		return database.User{}, ErrCredentials
	}
	q := database.New(tx)
	user, err := q.GetUserByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		var existing int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE lower(email)=?`, email).Scan(&existing); err != nil {
			return database.User{}, err
		}
		if existing != 0 {
			return database.User{}, ErrEmailExists
		}
		user, err = CreateUserTx(ctx, tx, database.CreateUserParams{Email: email, Name: strings.TrimSpace(name), Password: "!", Role: "editor", Active: 1})
	} else if err == nil && user.Role == "user" {
		user, err = q.SetUserRole(ctx, database.SetUserRoleParams{ID: user.ID, Role: "editor"})
		if err == nil {
			_, err = tx.ExecContext(ctx, `DELETE FROM web_sessions WHERE user_id=?`, user.ID)
		}
		if err == nil {
			err = q.DeleteUserSessions(ctx, user.ID)
		}
	}
	if err != nil {
		return database.User{}, err
	}
	if _, err := publicid.Ensure(ctx, tx, "users", strconv.FormatInt(user.ID, 10)); err != nil {
		return database.User{}, err
	}
	if user.Verified != 1 {
		if err := s.queueLink(ctx, tx, user.ID, email, name, "verify", urls.AcceptInvitationPath); err != nil {
			return database.User{}, err
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
	err := s.db.QueryRowContext(ctx, `SELECT u.name,u.email FROM user_tokens t JOIN users u ON u.id=t.user_id JOIN people p ON p.user_id=u.id WHERE t.id=? AND t.kind='verify' AND t.expires_at>? AND u.verified=0 AND u.deleted_at IS NULL`, digest(value), datetime.FormatUTC(time.Now())).Scan(&result.Name, &result.Email)
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
	err = tx.QueryRowContext(ctx, `DELETE FROM user_tokens WHERE id=? AND kind='verify' AND expires_at>? AND user_id IN (SELECT u.id FROM users u JOIN people p ON p.user_id=u.id WHERE u.verified=0 AND u.deleted_at IS NULL) RETURNING user_id`, digest(value), datetime.FormatUTC(time.Now())).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrToken
	}
	if err != nil {
		return err
	}
	if _, err := database.New(tx).UpdateUserPassword(ctx, database.UpdateUserPasswordParams{ID: userID, Password: hash}); err != nil {
		return err
	}
	if _, err := database.New(tx).VerifyUser(ctx, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM web_sessions WHERE user_id=?`, userID); err != nil {
		return err
	}
	if err := database.New(tx).DeleteUserSessions(ctx, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM mail_outbox WHERE user_id=? AND kind='verify'`, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// ResendConfirmation replaces the outstanding link without changing the account.
// Members complete their invitation; other accounts verify their email address.
func (s *Service) ResendConfirmation(ctx context.Context, userID int64) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	user, err := database.New(tx).GetUser(ctx, userID)
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
	path := urls.ConfirmEmailPath
	if member {
		path = urls.AcceptInvitationPath
	}
	if err := s.queueLink(ctx, tx, user.ID, user.Email, user.Name, "verify", path); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
