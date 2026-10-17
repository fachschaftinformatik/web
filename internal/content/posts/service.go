package posts

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fachschaftinformatik/web/internal/id"
)

var ErrForbidden = errors.New("forum access denied")
var ErrNotFound = errors.New("forum record not found")

type ValidationError struct{ Fields []FieldError }

func (e *ValidationError) Error() string { return "invalid forum input" }

type Service struct{ db *sql.DB }

func NewService(db *sql.DB) *Service { return &Service{db: db} }

type querier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func permission(ctx context.Context, db querier, user, owner int64, comments bool) error {
	var role string
	err := db.QueryRowContext(ctx, `SELECT role FROM users WHERE id=? AND active=1 AND verified=1 AND deleted_at IS NULL`, user).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		return err
	}
	if owner != 0 && owner != user && role != "admin" && (comments || role != "editor") {
		return ErrForbidden
	}
	return nil
}

func (s *Service) CanWrite(ctx context.Context, user, owner int64, comments bool) error {
	return permission(ctx, s.db, user, owner, comments)
}

const postSelect = `SELECT CAST(p.id AS TEXT),p.title,p.body,p.body_format,p.user_id,CASE WHEN u.private=1 THEN 'Anonym' ELSE u.name END,p.created_at,p.updated_at,p.type,COALESCE(p.event_date,''),COALESCE(p.location,''),(SELECT COUNT(*) FROM discussion_comments c WHERE c.post_id=p.id AND c.deleted_at IS NULL),COALESCE((SELECT SUM(vote) FROM discussion_votes WHERE post_id=p.id),0),COALESCE((SELECT vote FROM discussion_votes WHERE post_id=p.id AND user_id=?),0) FROM discussion_posts p JOIN users u ON u.id=p.user_id WHERE p.deleted_at IS NULL`

func scanThread(row interface{ Scan(...any) error }) (Thread, error) {
	var t Thread
	err := row.Scan(&t.ID, &t.Title, &t.Body, &t.BodyFormat, &t.AuthorID, &t.Author, &t.Created, &t.Updated, &t.Type, &t.EventDate, &t.Location, &t.Comments, &t.Votes, &t.UserVote)
	if errors.Is(err, sql.ErrNoRows) {
		return t, ErrNotFound
	}
	text := t.Body
	if t.BodyFormat == "html" {
		t.Body = SanitizeRichText(t.Body)
		text = RichTextVisible(t.Body)
	}
	t.Summary = text
	if r := []rune(text); len(r) > 240 {
		t.Summary = string(r[:240]) + "…"
	}
	return t, err
}

func (s *Service) tags(ctx context.Context, t *Thread) error {
	rows, err := s.db.QueryContext(ctx, `SELECT tag FROM discussion_post_tags WHERE post_id=? ORDER BY tag`, t.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return err
		}
		t.Tags = append(t.Tags, tag)
	}
	return rows.Err()
}

func (s *Service) Find(ctx context.Context, f Filter, viewer ...int64) ([]Post, error) {
	rows, err := s.db.QueryContext(ctx, postSelect+` ORDER BY p.pinned DESC,p.created_at DESC,p.id DESC`, viewerID(viewer))
	if err != nil {
		return nil, err
	}
	var threads []Thread
	for rows.Next() {
		t, err := scanThread(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		threads = append(threads, t)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	var result []Post
	for _, t := range threads {
		if err := s.tags(ctx, &t); err != nil {
			return nil, err
		}
		if f.Type != "" && t.Type != f.Type {
			continue
		}
		text := strings.ToLower(t.Title + " " + t.Summary + " " + t.Body + " " + strings.Join(t.Tags, " "))
		if slices.ContainsFunc(strings.Fields(strings.ToLower(f.Query)), func(term string) bool { return !strings.Contains(text, term) }) {
			continue
		}
		result = append(result, t.Post)
	}
	return result, nil
}

func (s *Service) Lookup(ctx context.Context, post string, viewer ...int64) (Thread, error) {
	t, err := scanThread(s.db.QueryRowContext(ctx, postSelect+` AND p.id=?`, viewerID(viewer), post))
	if err != nil {
		return t, err
	}
	if err = s.tags(ctx, &t); err != nil {
		return t, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT CAST(c.id AS TEXT),c.user_id,CASE WHEN u.private=1 THEN 'Anonym' ELSE u.name END,c.text,c.created_at,c.updated_at,COALESCE((SELECT SUM(vote) FROM discussion_comment_votes WHERE comment_id=c.id),0),COALESCE((SELECT vote FROM discussion_comment_votes WHERE comment_id=c.id AND user_id=?),0) FROM discussion_comments c JOIN users u ON u.id=c.user_id WHERE c.post_id=? AND c.deleted_at IS NULL ORDER BY c.created_at DESC,c.id DESC`, viewerID(viewer), post)
	if err != nil {
		return t, err
	}
	defer rows.Close()
	for rows.Next() {
		var c Comment
		if err = rows.Scan(&c.ID, &c.AuthorID, &c.Author, &c.Text, &c.Created, &c.Updated, &c.Votes, &c.UserVote); err != nil {
			return t, err
		}
		t.Replies = append(t.Replies, c)
	}
	return t, rows.Err()
}

// Save persists sanitized HTML, event details and tags atomically. An empty post creates a new record.
func (s *Service) Save(ctx context.Context, user int64, post string, d Draft, rich string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var owner int64
	if post != "" {
		if err = tx.QueryRowContext(ctx, `SELECT user_id FROM discussion_posts WHERE id=? AND deleted_at IS NULL`, post).Scan(&owner); errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		} else if err != nil {
			return "", err
		}
	}
	if err = permission(ctx, tx, user, owner, false); err != nil {
		return "", err
	}
	d.Title = strings.TrimSpace(d.Title)
	body, format := d.Body, "plain"
	if rich != "" {
		body = SanitizeRichText(rich)
		format = "html"
		d.Body = RichTextVisible(body)
	}
	if fields := d.Validate(); len(fields) > 0 {
		return "", &ValidationError{fields}
	}
	event := ""
	if d.Type == "event" && strings.TrimSpace(d.Date) != "" {
		date, _ := time.Parse("2.1.2006", strings.TrimSpace(d.Date))
		clock := strings.TrimSpace(d.Time)
		if clock == "" {
			clock = "00:00"
		}
		event = date.Format("2006-01-02") + "T" + clock
	}
	if post == "" {
		post = id.New().String()
		_, err = tx.ExecContext(ctx, `INSERT INTO discussion_posts(id,title,body,body_format,user_id,type,event_date,location) VALUES(?,?,?,?,?,?,?,?)`, post, d.Title, body, format, user, d.Type, event, strings.TrimSpace(d.Location))
		if err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO activities(id,user_id,type,target_id,target_name) VALUES(?,?,'POST_CREATED',?,?)`, id.New().Int64(), user, post, d.Title)
		}
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE discussion_posts SET title=?,body=?,body_format=?,type=?,event_date=?,location=? WHERE id=?`, d.Title, body, format, d.Type, event, strings.TrimSpace(d.Location), post)
	}
	if err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM discussion_post_tags WHERE post_id=?`, post); err != nil {
		return "", err
	}
	for _, tag := range strings.Split(d.Tags, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO discussion_post_tags(post_id,tag) VALUES(?,?)`, post, tag); err != nil {
				return "", err
			}
		}
	}
	return post, tx.Commit()
}

func (s *Service) SaveComment(ctx context.Context, user int64, post, comment, text string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var title string
	if err = tx.QueryRowContext(ctx, `SELECT title FROM discussion_posts WHERE id=? AND deleted_at IS NULL`, post).Scan(&title); errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	} else if err != nil {
		return "", err
	}
	var owner int64
	if comment != "" {
		if err = tx.QueryRowContext(ctx, `SELECT user_id FROM discussion_comments WHERE id=? AND post_id=? AND deleted_at IS NULL`, comment, post).Scan(&owner); errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		} else if err != nil {
			return "", err
		}
	}
	if err = permission(ctx, tx, user, owner, true); err != nil {
		return "", err
	}
	text = strings.TrimSpace(text)
	if n := utf8.RuneCountInString(text); n < 1 || n > 10000 {
		return "", &ValidationError{[]FieldError{{"text", "Der Kommentar muss zwischen 1 und 10.000 Zeichen lang sein."}}}
	}
	if comment == "" {
		comment = id.New().String()
		_, err = tx.ExecContext(ctx, `INSERT INTO discussion_comments(id,post_id,user_id,text) VALUES(?,?,?,?)`, comment, post, user, text)
		if err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO activities(id,user_id,type,target_id,target_name) VALUES(?,?,'COMMENT_ADDED',?,?)`, id.New().Int64(), user, post, title)
		}
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE discussion_comments SET text=? WHERE id=? AND post_id=?`, text, comment, post)
	}
	if err != nil {
		return "", err
	}
	return comment, tx.Commit()
}
