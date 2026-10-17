package people

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"

	"github.com/fachschaftinformatik/web/internal/id"
	"github.com/fachschaftinformatik/web/internal/model/users"
)

func Open(ctx context.Context, db *sql.DB) (*Store, error) {
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
	for _, task := range tasks {
		rows, err := tx.QueryContext(ctx, `SELECT person_id FROM responsibility_members WHERE responsibility_id=? ORDER BY position`, task.ID)
		if err != nil {
			return err
		}
		var current []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			current = append(current, id)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		if slices.Equal(current, task.MemberIDs) {
			continue
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM responsibility_members WHERE responsibility_id=?`, task.ID); err != nil {
			return err
		}
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
	rows, err := tx.QueryContext(ctx, `SELECT p.id,COALESCE(u.name,p.name),p.email,COALESCE(p.user_id,0),COALESCE(u.email,''),COALESCE(u.verified=1 AND u.deleted_at IS NULL,0) FROM people p LEFT JOIN users u ON u.id=p.user_id ORDER BY p.position,p.id`)
	if err != nil {
		return err
	}
	var people []Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Email, &p.UserID, &p.AccountEmail, &p.Confirmed); err != nil {
			rows.Close()
			return err
		}
		p.AccountEmail = strings.ToLower(p.AccountEmail)
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
	var userID int64
	err := s.db.QueryRowContext(ctx, `SELECT u.id FROM people p JOIN users u ON u.id=p.user_id WHERE p.id=? AND u.deleted_at IS NULL`, personID).Scan(&userID)
	if err != nil {
		return false, err
	}
	return identity.ResendConfirmation(ctx, userID)
}

func (s *Store) Invite(ctx context.Context, identity *users.Service, name, email string) (map[string]string, error) {
	return s.InviteMember(ctx, identity, "neu", name, email)
}

// InviteMember links an existing public profile or creates a new membership.
// Account creation, profile linkage and queued mail commit together.
func (s *Store) InviteMember(ctx context.Context, identity *users.Service, personID, name, email string) (map[string]string, error) {
	fields := map[string]string{}
	name = strings.TrimSpace(name)
	nameError := users.NameError(name)
	if nameError != "" {
		fields["name"] = nameError
	}
	if message := identity.EmailError(email); message != "" {
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
	if personID != "neu" {
		var linked sql.NullInt64
		if err := tx.QueryRowContext(ctx, `SELECT user_id FROM people WHERE id=?`, personID).Scan(&linked); err != nil {
			return nil, err
		}
		if linked.Valid {
			return map[string]string{"email": "Dieses Mitglied hat bereits ein Benutzerkonto."}, nil
		}
	}
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
	if personID == "neu" {
		if _, err := tx.ExecContext(ctx, `INSERT INTO people(id,name,user_id,position) VALUES (?,?,?,?)`, id.New().String(), user.Name, user.ID, len(s.people)); err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `UPDATE people SET user_id=?,name=? WHERE id=? AND user_id IS NULL`, user.ID, user.Name, personID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return nil, s.refresh(ctx)
}

func (s *Store) saveLinked(ctx context.Context, p Person) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE users SET name=? WHERE id=? AND deleted_at IS NULL`, p.Name, p.UserID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return sql.ErrNoRows
	}
	if _, err := tx.ExecContext(ctx, `UPDATE people SET name=?,email=? WHERE id=? AND user_id=?`, p.Name, p.Email, p.ID, p.UserID); err != nil {
		return err
	}
	return tx.Commit()
}
