package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/fachschaftinformatik/web/internal/datetime"
	"github.com/pressly/goose/v3"
	"time"
)

// Legacy event times were entered and displayed as Berlin wall-clock times.
// Preserve the instant and publication metadata when converting them to UTC.
func utcEventTimes(ctx context.Context, tx *sql.Tx) error {
	var trigger string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='trigger' AND name='trg_discussion_posts_updated_at'`).Scan(&trigger); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DROP TRIGGER trg_discussion_posts_updated_at`); err != nil {
		return err
	}
	for _, table := range []string{"calendar_events", "discussion_posts"} {
		rows, err := tx.QueryContext(ctx, `SELECT id,event_date FROM `+table+` WHERE length(event_date)>10`)
		if err != nil {
			return err
		}
		type update struct {
			id    int64
			value string
		}
		var updates []update
		for rows.Next() {
			var id int64
			var value string
			if err := rows.Scan(&id, &value); err != nil {
				rows.Close()
				return err
			}
			var canonical string
			if date, err := time.Parse(time.RFC3339, value); err == nil {
				canonical = datetime.FormatUTC(date)
			} else {
				canonical, err = datetime.LocalTimestamp(value)
				if err != nil {
					rows.Close()
					return fmt.Errorf("convert %s event %d: %w", table, id, err)
				}
			}
			if canonical != value {
				updates = append(updates, update{id, canonical})
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		for _, row := range updates {
			if _, err := tx.ExecContext(ctx, `UPDATE `+table+` SET event_date=? WHERE id=?`, row.value, row.id); err != nil {
				return err
			}
		}
	}
	_, err := tx.ExecContext(ctx, trigger)
	return err
}

func utcEventMigration() *goose.Migration {
	return goose.NewGoMigration(15, &goose.GoFunc{RunTx: utcEventTimes}, nil)
}
