// Package events manages standalone Fachschaft events, independently of forum posts.
package events

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/fachschaftinformatik/web/internal/content/posts"
	"github.com/fachschaftinformatik/web/internal/data/blob"
	"github.com/fachschaftinformatik/web/internal/id"
)

var ErrNotFound = errors.New("event not found")
var ErrForbidden = posts.ErrForbidden

type ValidationError = posts.ValidationError
type FieldError = posts.FieldError
type querier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}
type Service struct {
	db      *sql.DB
	storage blob.Provider
}

func NewService(db *sql.DB, storage blob.Provider) *Service {
	return &Service{db: db, storage: storage}
}

type Event struct {
	ID, Title, Body, BodyFormat, EventDate, Location, Created, Updated string
	CoverKey, CoverAlt                                                 string
	CoverWidth, CoverHeight                                            int
	LegacyPostID                                                       sql.NullInt64
}

func (e Event) URL() string { return "/community/veranstaltungen/" + e.ID }
func (e Event) CoverURL() string {
	if e.CoverKey != "" {
		return e.URL() + "/cover"
	}
	return ""
}
func (e Event) LegacyURL() string {
	if e.LegacyPostID.Valid {
		return "/d/" + strconv.FormatInt(e.LegacyPostID.Int64, 10)
	}
	return ""
}
func (e Event) EventLabel() string {
	if len(e.EventDate) == 10 {
		return (posts.Post{Created: e.EventDate + "T00:00:00Z"}).DateLabel()
	}
	return (posts.Thread{EventDate: e.EventDate}).EventLabel()
}
func (e Event) RichBody() string {
	if e.BodyFormat == "html" {
		return posts.SanitizeRichText(e.Body)
	}
	return posts.PlainRichText(e.Body)
}
func (e Event) Summary() string {
	text := e.Body
	if e.BodyFormat == "html" {
		text = posts.RichTextVisible(text)
	}
	if r := []rune(text); len(r) > 240 {
		return string(r[:240]) + "…"
	}
	return text
}

type Draft struct{ Title, Body, Date, Time, Location string }

func (d Draft) Validate() []FieldError {
	return (posts.Draft{Title: d.Title, Body: d.Body, Date: d.Date, Time: d.Time, Location: d.Location, Type: "event"}).Validate()
}

const selectEvent = `SELECT CAST(id AS TEXT),title,body,body_format,event_date,location,created_at,updated_at,cover_key,cover_alt,cover_width,cover_height,legacy_post_id FROM calendar_events WHERE deleted_at IS NULL`

func scan(row interface{ Scan(...any) error }) (Event, error) {
	var e Event
	err := row.Scan(&e.ID, &e.Title, &e.Body, &e.BodyFormat, &e.EventDate, &e.Location, &e.Created, &e.Updated, &e.CoverKey, &e.CoverAlt, &e.CoverWidth, &e.CoverHeight, &e.LegacyPostID)
	if errors.Is(err, sql.ErrNoRows) {
		return e, ErrNotFound
	}
	if e.BodyFormat == "html" {
		e.Body = posts.SanitizeRichText(e.Body)
	}
	return e, err
}
func (s *Service) Find(ctx context.Context) ([]Event, error) {
	rows, err := s.db.QueryContext(ctx, selectEvent+` ORDER BY event_date,created_at DESC,id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Event
	for rows.Next() {
		e, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
func (s *Service) Lookup(ctx context.Context, event string) (Event, error) {
	return scan(s.db.QueryRowContext(ctx, selectEvent+` AND id=?`, event))
}

func (s *Service) Save(ctx context.Context, user int64, event string, d Draft, rich string, cover *CoverUpload) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if err := permission(ctx, tx, user); err != nil {
		return "", err
	}
	var oldKey, coverAlt string
	var width, height int
	if event != "" {
		err = tx.QueryRowContext(ctx, `SELECT cover_key,cover_alt,cover_width,cover_height FROM calendar_events WHERE id=? AND deleted_at IS NULL`, event).Scan(&oldKey, &coverAlt, &width, &height)
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		if err != nil {
			return "", err
		}
	}
	body, format := d.Body, "plain"
	if rich != "" {
		body, format = posts.SanitizeRichText(rich), "html"
		d.Body = posts.RichTextVisible(body)
	}
	if fields := d.Validate(); len(fields) > 0 {
		return "", &ValidationError{Fields: fields}
	}
	coverKey := oldKey
	newKey := ""
	committed := false
	defer func() {
		if !committed && newKey != "" {
			_ = s.storage.Delete(context.WithoutCancel(ctx), newKey)
		}
	}()
	if cover != nil {
		if cover.Remove && len(cover.Data) > 0 {
			return "", &ValidationError{Fields: []FieldError{{Field: "cover", Message: "Wähle entweder ein neues Titelbild oder entferne das bisherige."}}}
		}
		if cover.Remove {
			coverKey, coverAlt, width, height = "", "", 0, 0
		}
		if len(cover.Data) > 0 {
			if strings.TrimSpace(cover.Alt) == "" || utf8.RuneCountInString(cover.Alt) > 500 {
				return "", &ValidationError{Fields: []FieldError{{Field: "cover_alt", Message: "Beschreibe das Titelbild mit höchstens 500 Zeichen."}}}
			}
			photo, w, h, err := prepareCover(cover.Data)
			if err != nil {
				return "", err
			}
			if s.storage == nil {
				return "", ErrCoverStorage
			}
			newKey = "calendar/covers/" + id.New().String() + ".jpg"
			if err := s.storage.Upload(ctx, newKey, bytes.NewReader(photo), int64(len(photo)), "image/jpeg"); err != nil {
				return "", ErrCoverStorage
			}
			coverKey, width, height = newKey, w, h
		}
		if coverKey != "" && !cover.Remove {
			if strings.TrimSpace(cover.Alt) == "" || utf8.RuneCountInString(cover.Alt) > 500 {
				return "", &ValidationError{Fields: []FieldError{{Field: "cover_alt", Message: "Beschreibe das Titelbild mit höchstens 500 Zeichen."}}}
			}
			coverAlt = strings.TrimSpace(cover.Alt)
		}
	}
	// Reuse the established date parser without creating a forum post.
	date := ""
	if d.Date != "" {
		parsed, _ := parseDate(d.Date)
		clock := strings.TrimSpace(d.Time)
		date = parsed
		if clock != "" {
			date += "T" + clock
		}
	}
	if event == "" {
		event = id.New().String()
		_, err = tx.ExecContext(ctx, `INSERT INTO calendar_events(id,title,body,body_format,event_date,location,user_id,cover_key,cover_alt,cover_width,cover_height) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, event, strings.TrimSpace(d.Title), body, format, date, strings.TrimSpace(d.Location), user, coverKey, coverAlt, width, height)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE calendar_events SET title=?,body=?,body_format=?,event_date=?,location=?,cover_key=?,cover_alt=?,cover_width=?,cover_height=?,updated_at=strftime('%Y-%m-%dT%H:%M:%SZ','now') WHERE id=?`, strings.TrimSpace(d.Title), body, format, date, strings.TrimSpace(d.Location), coverKey, coverAlt, width, height, event)
	}
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	committed = true
	if validCoverKey(oldKey) && oldKey != coverKey && s.storage != nil {
		_ = s.storage.Delete(context.WithoutCancel(ctx), oldKey)
	}
	return event, nil
}
