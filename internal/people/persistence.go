package people

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/fachschaftinformatik/web/internal/id"
	"github.com/fachschaftinformatik/web/internal/users"
)

func Open(ctx context.Context, db *sql.DB) (*Store, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	insert, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO people_state(id) VALUES (1)`)
	if err != nil {
		return nil, err
	}
	count, err := insert.RowsAffected()
	if err != nil {
		return nil, err
	}
	if count != 0 {
		seed := NewStore()
		if err := writeState(ctx, tx, seed.people, seed.responsibilities); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.Refresh(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func writeState(ctx context.Context, tx *sql.Tx, people []Person, tasks []Responsibility) error {
	for position, p := range people {
		var user any
		if p.UserID != 0 {
			user = p.UserID
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO people(id,name,email,user_id,position) VALUES (?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,email=excluded.email,user_id=excluded.user_id,position=excluded.position`, p.ID, p.Name, p.Email, user, position); err != nil {
			return err
		}
	}
	for position, task := range tasks {
		if _, err := tx.ExecContext(ctx, `INSERT INTO responsibilities(id,group_name,title,email,position) VALUES (?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET group_name=excluded.group_name,title=excluded.title,email=excluded.email,position=excluded.position`, task.ID, task.Group, task.Title, task.Email, position); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM responsibility_members`); err != nil {
		return err
	}
	for _, task := range tasks {
		for position, member := range task.MemberIDs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO responsibility_members(responsibility_id,person_id,position) VALUES (?,?,?)`, task.ID, member, position); err != nil {
				return err
			}
		}
	}
	return nil
}

// Refresh reads account confirmation without copying private login data into profiles.
func (s *Store) Refresh(ctx context.Context) error {
	if s.db == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.refresh(ctx)
}

func (s *Store) refresh(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT p.id,p.name,p.email,COALESCE(p.user_id,0),COALESCE(u.email,''),p.user_id IS NULL OR (u.verified=1 AND u.deleted_at IS NULL) FROM people p LEFT JOIN users u ON u.id=p.user_id ORDER BY p.position,p.id`)
	if err != nil {
		return err
	}
	var people []Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Email, &p.UserID, &p.InviteEmail, &p.Confirmed); err != nil {
			rows.Close()
			return err
		}
		p.InviteEmail = strings.ToLower(p.InviteEmail)
		people = append(people, p)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	rows, err = tx.QueryContext(ctx, `SELECT id,group_name,title,email FROM responsibilities ORDER BY position,id`)
	if err != nil {
		return err
	}
	var tasks []Responsibility
	for rows.Next() {
		var task Responsibility
		if err := rows.Scan(&task.ID, &task.Group, &task.Title, &task.Email); err != nil {
			rows.Close()
			return err
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	rows, err = tx.QueryContext(ctx, `SELECT responsibility_id,person_id FROM responsibility_members ORDER BY position,person_id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var taskID, personID string
		if err := rows.Scan(&taskID, &personID); err != nil {
			rows.Close()
			return err
		}
		for i := range tasks {
			if tasks[i].ID == taskID {
				tasks[i].MemberIDs = append(tasks[i].MemberIDs, personID)
				break
			}
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.people, s.responsibilities = people, tasks
	return nil
}

func (s *Store) Resend(ctx context.Context, identity *users.Service, personID string) (bool, error) {
	if s.db == nil {
		return false, errors.New("persistent membership storage is unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var name, email string
	err = tx.QueryRowContext(ctx, `SELECT p.name,u.email FROM people p JOIN users u ON u.id=p.user_id WHERE p.id=? AND u.deleted_at IS NULL`, personID).Scan(&name, &email)
	if err != nil {
		return false, err
	}
	user, err := identity.Invite(ctx, tx, name, email)
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return user.Verified != 1, nil
}

func (s *Store) Invite(ctx context.Context, identity *users.Service, name, email string) (map[string]string, error) {
	fields := map[string]string{}
	name = strings.TrimSpace(name)
	nameError, _, _ := users.RegistrationErrors(name, "valid@studmail.w-hs.de", "valid password")
	if nameError != "" {
		fields["name"] = nameError
	}
	if message := users.EmailError(email); message != "" {
		fields["email"] = message
	}
	if len(fields) != 0 {
		return fields, nil
	}
	if s.db == nil {
		return nil, errors.New("persistent membership storage is unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	user, err := identity.Invite(ctx, tx, name, email)
	if errors.Is(err, users.ErrEmailExists) {
		return map[string]string{"email": err.Error()}, nil
	}
	if err != nil {
		return nil, err
	}
	var existing int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM people WHERE user_id=?`, user.ID).Scan(&existing); err != nil {
		return nil, err
	}
	if existing != 0 {
		return map[string]string{"email": "Dieses Benutzerkonto ist bereits einem Mitglied zugeordnet."}, nil
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO people(id,name,user_id,position) VALUES (?,?,?,?)`, id.New().String(), name, user.ID, len(s.people)); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return nil, s.refresh(ctx)
}
