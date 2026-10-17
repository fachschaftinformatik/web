package users

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/data"
	"github.com/fachschaftinformatik/web/internal/id"
)

var (
	ErrCredentials = errors.New("E-Mail-Adresse oder Passwort ist falsch.")
	ErrInactive    = errors.New("Dein Konto ist deaktiviert.")
	ErrUnverified  = errors.New("Bestätige zuerst deine E-Mail-Adresse.")
	ErrToken       = errors.New("Der Link ist ungültig oder abgelaufen. Fordere einen neuen Link an.")
	ErrEmailExists = errors.New("Für diese E-Mail-Adresse besteht bereits ein Konto.")
)

type Service struct {
	db      *sql.DB
	queries *data.Queries
	baseURL string
}

func NewService(db *sql.DB, baseURL string) *Service {
	return &Service{db: db, queries: data.New(db), baseURL: strings.TrimRight(baseURL, "/")}
}

type Session struct {
	ID, CSRF      string
	UserID        int64
	Role          string
	Authenticated bool
	Expires       time.Time
}

func (s Session) CanAdmin() bool { return s.Authenticated && s.Role == "admin" }

func token() (string, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (s *Service) CreateSession(ctx context.Context, userID int64) (*Session, error) {
	value, err := token()
	if err != nil {
		return nil, err
	}
	expires := time.Now().UTC().Add(8 * time.Hour)
	var user any
	if userID != 0 {
		user = userID
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO web_sessions (id,user_id,expires_at) VALUES (?,?,?)`, digest(value), user, expires.Format(time.RFC3339)); err != nil {
		return nil, err
	}
	return s.Session(ctx, value)
}

func (s *Service) Session(ctx context.Context, value string) (*Session, error) {
	if len(value) != 64 {
		return nil, nil
	}
	var userID sql.NullInt64
	var role, expiry string
	var active, verified int
	err := s.db.QueryRowContext(ctx, `SELECT ws.user_id, ws.expires_at, COALESCE(u.role,''), COALESCE(u.active,0), COALESCE(u.verified,0) FROM web_sessions ws LEFT JOIN users u ON u.id=ws.user_id AND u.deleted_at IS NULL WHERE ws.id=?`, digest(value)).Scan(&userID, &expiry, &role, &active, &verified)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	expires, err := time.Parse(time.RFC3339, expiry)
	if err != nil || !time.Now().Before(expires) || userID.Valid && (active != 1 || verified != 1 || role == "") {
		if err := s.DeleteSession(ctx, value); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return &Session{ID: value, CSRF: digest("csrf:" + value), UserID: userID.Int64, Role: role, Authenticated: userID.Valid, Expires: expires}, nil
}

func (s *Service) DeleteSession(ctx context.Context, value string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM web_sessions WHERE id=?`, digest(value))
	return err
}

func (s *Service) Login(ctx context.Context, email, password string) (int64, error) {
	if EmailError(email) != "" {
		return 0, ErrCredentials
	}
	user, err := s.queries.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, sql.ErrNoRows) {
		// Match the password-work factor without exposing whether the user exists.
		VerifyPassword("$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", password)
		return 0, ErrCredentials
	}
	if err != nil {
		return 0, err
	}
	if !VerifyPassword(user.Password, password) {
		return 0, ErrCredentials
	}
	if user.Active != 1 {
		return 0, ErrInactive
	}
	if user.Verified != 1 {
		return 0, ErrUnverified
	}
	return user.ID, nil
}

func (s *Service) Directory(ctx context.Context) (*Directory, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT CAST(u.id AS TEXT),u.name,lower(u.email),u.role,COALESCE(p.name,''),CASE WHEN u.active=1 THEN 'Aktiv' ELSE 'Inaktiv' END,u.verified=1 FROM users u LEFT JOIN programs p ON p.id=u.program_id WHERE u.deleted_at IS NULL ORDER BY u.name COLLATE NOCASE,u.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Permission, &user.Program, &user.Status, &user.Verified); err != nil {
			return nil, err
		}
		switch user.Permission {
		case "admin":
			user.Permission = "Administration"
		case "editor":
			user.Permission = "Redaktion"
		case "user":
			user.Permission = "Mitglied"
		}
		records = append(records, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return NewDirectory(records), nil
}

func (s *Service) Programs(ctx context.Context) ([]Program, error) {
	rows, err := s.queries.ListPrograms(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Program, 0, len(rows))
	for _, p := range rows {
		result = append(result, Program{ID: strconv.FormatInt(p.ID, 10), Name: p.Name})
	}
	return result, nil
}

func (s *Service) Register(ctx context.Context, name, email, password, programID string) error {
	a, b, c := RegistrationErrors(name, email, password)
	if a != "" || b != "" || c != "" {
		return errors.New("Prüfe die Angaben im Formular.")
	}
	var program *int64
	if programID != "" {
		value, err := strconv.ParseInt(programID, 10, 64)
		if err != nil {
			return errors.New("Wähle einen gültigen Studiengang aus.")
		}
		if _, err = s.queries.GetProgram(ctx, value); err != nil {
			return errors.New("Wähle einen gültigen Studiengang aus.")
		}
		program = &value
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
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE lower(email)=lower(?)`, strings.TrimSpace(email)).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return ErrEmailExists
	}
	userID := int64(id.New())
	_, err = data.New(tx).CreateUser(ctx, data.CreateUserParams{ID: userID, Email: strings.ToLower(strings.TrimSpace(email)), Name: strings.TrimSpace(name), Password: hash, Role: "user", Active: 1, ProgramID: program})
	if err != nil {
		return err
	}
	if err := s.queueToken(ctx, tx, userID, strings.ToLower(strings.TrimSpace(email)), strings.TrimSpace(name), "verify"); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) queueToken(ctx context.Context, tx *sql.Tx, userID int64, email, name, kind string) error {
	value, err := token()
	if err != nil {
		return err
	}
	duration := time.Hour
	if kind == "verify" {
		duration = 24 * time.Hour
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO user_tokens (id,user_id,kind,expires_at) VALUES (?,?,?,?) ON CONFLICT(user_id,kind) DO UPDATE SET id=excluded.id,expires_at=excluded.expires_at`, digest(value), userID, kind, time.Now().UTC().Add(duration).Format(time.RFC3339))
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM mail_outbox WHERE user_id=? AND kind=?`, userID, kind); err != nil {
		return err
	}
	path := "/reset"
	if kind == "verify" {
		path = "/verify"
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO mail_outbox (id,user_id,kind,recipient,name,link) VALUES (?,?,?,?,?,?)`, int64(id.New()), userID, kind, email, name, s.baseURL+path+"?token="+url.QueryEscape(value))
	return err
}

func (s *Service) RequestReset(ctx context.Context, email string) error {
	if EmailError(email) != "" {
		return ErrCredentials
	}
	user, err := s.queries.GetUserByEmail(ctx, strings.TrimSpace(email))
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := s.queueToken(ctx, tx, user.ID, strings.ToLower(user.Email), user.Name, "reset"); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) ConsumeToken(ctx context.Context, value, kind, password string) error {
	if len(value) != 64 || kind != "verify" && kind != "reset" {
		return ErrToken
	}
	var hash string
	if kind == "reset" {
		_, _, message := RegistrationErrors("Valid Name", "valid@studmail.w-hs.de", password)
		if message != "" {
			return errors.New(message)
		}
		var err error
		hash, err = HashPassword(password)
		if err != nil {
			return err
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var userID int64
	err = tx.QueryRowContext(ctx, `DELETE FROM user_tokens WHERE id=? AND kind=? AND expires_at>? AND user_id IN (SELECT id FROM users WHERE deleted_at IS NULL) RETURNING user_id`, digest(value), kind, time.Now().UTC().Format(time.RFC3339)).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrToken
	}
	if err != nil {
		return err
	}
	if kind == "verify" {
		_, err = data.New(tx).VerifyUser(ctx, userID)
	} else {
		_, err = data.New(tx).UpdateUserPassword(ctx, data.UpdateUserPasswordParams{ID: userID, Password: hash})
		if err == nil {
			_, err = tx.ExecContext(ctx, `DELETE FROM web_sessions WHERE user_id=?`, userID)
		}
		if err == nil {
			err = data.New(tx).DeleteUserSessions(ctx, userID)
		}
	}
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM mail_outbox WHERE user_id=? AND kind=?`, userID, kind); err != nil {
		return err
	}
	return tx.Commit()
}

type Mailer interface {
	SendVerificationEmail(to, name, link string) error
	SendPasswordResetEmail(to, name, link string) error
}

// DeliverMail retains failed messages for a later retry.
func (s *Service) DeliverMail(ctx context.Context, mailer Mailer) error {
	if err := s.PurgeExpired(ctx); err != nil {
		return err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,kind,recipient,name,link FROM mail_outbox ORDER BY id LIMIT 20`)
	if err != nil {
		return err
	}
	type message struct {
		id                   int64
		kind, to, name, link string
	}
	var pending []message
	for rows.Next() {
		var m message
		if err := rows.Scan(&m.id, &m.kind, &m.to, &m.name, &m.link); err != nil {
			rows.Close()
			return err
		}
		pending = append(pending, m)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	var failures []error
	for _, m := range pending {
		if err := ctx.Err(); err != nil {
			return err
		}
		if m.kind == "verify" {
			err = mailer.SendVerificationEmail(m.to, m.name, m.link)
		} else {
			err = mailer.SendPasswordResetEmail(m.to, m.name, m.link)
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("deliver queued mail: %w", err))
			continue
		}
		if _, err = s.db.ExecContext(ctx, `DELETE FROM mail_outbox WHERE id=?`, m.id); err != nil {
			return err
		}
	}
	return errors.Join(failures...)
}

func (s *Service) PurgeExpired(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.ExecContext(ctx, `DELETE FROM web_sessions WHERE expires_at<=?`, now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_tokens WHERE expires_at<=?`, now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM mail_outbox WHERE NOT EXISTS (SELECT 1 FROM user_tokens t JOIN users u ON u.id=t.user_id WHERE t.user_id=mail_outbox.user_id AND t.kind=mail_outbox.kind AND u.deleted_at IS NULL)`); err != nil {
		return err
	}
	return tx.Commit()
}
