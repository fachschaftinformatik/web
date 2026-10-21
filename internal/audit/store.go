package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"
)

// Record contains only the metadata intentionally selected for an audit event.
type Record struct {
	Time                     time.Time
	Event, Outcome           string
	ActorUserID, UserID      int64
	RequestID, Method, Route string
	Details                  []slog.Attr
}

type Recorder interface {
	Record(context.Context, Record) error
}

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// Record appends an event independently of the completed business transaction.
// Audit identifiers have no account foreign keys so history survives deletion.
func (s *Store) Record(ctx context.Context, record Record) error {
	details, err := json.Marshal(attributes(record.Details))
	if err != nil {
		return err
	}
	var actor, user any
	if record.ActorUserID != 0 {
		actor = record.ActorUserID
	}
	if record.UserID != 0 {
		user = record.UserID
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO audit_events
		(occurred_at,event,outcome,actor_user_id,user_id,request_id,method,route,details)
		VALUES(?,?,?,?,?,?,?,?,?)`, record.Time.UTC().Format(time.RFC3339), record.Event, record.Outcome, actor, user, record.RequestID, record.Method, record.Route, string(details))
	return err
}

func attributes(values []slog.Attr) map[string]any {
	result := map[string]any{}
	for _, attribute := range values {
		value := attribute.Value.Resolve()
		if value.Kind() == slog.KindGroup {
			result[attribute.Key] = attributes(value.Group())
		} else {
			result[attribute.Key] = value.Any()
		}
	}
	return result
}
