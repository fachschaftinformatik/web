package database

import (
	"context"
	"database/sql"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestDatabaseRejectsNonCanonicalTimestamps(t *testing.T) {
	path := filepath.Join(t.TempDir(), "utc.db")
	if err := Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	// Use a separate plain SQLite connection: no application serializer or
	// connection-local functions may be needed to enforce the invariant.
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT INTO users(id,email,name,password) VALUES(1,'utc@example.org','Name','hash'); INSERT INTO calendar_events(id,title,body,user_id) VALUES(1,'Event','Body',1)`); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{
		"2026-10-17T12:34:56", "2026-10-17T12:34:56+00:00", "2026-10-17T12:34:56+02:00",
		"2026-10-17 12:34:56Z", "2026-10-17T12:34:56z", "2026-10-17", "deleted",
		"2026-02-30T12:34:56Z", "2026-10-17T24:00:00Z", "2026-10-17T12:60:00Z",
		"2026-10-17T12:34:60Z", "2026-10-17T12:34:56.123Z",
	} {
		t.Run(value, func(t *testing.T) {
			for _, statement := range []string{
				`INSERT INTO users(email,name,password,created_at) VALUES('bad@example.org','Name','hash',?)`,
				`UPDATE users SET created_at=? WHERE id=1`,
				`UPDATE users SET deleted_at=? WHERE id=1`,
				`INSERT INTO web_sessions(id,user_id,expires_at) VALUES('bad',1,?)`,
			} {
				if _, err := db.Exec(statement, value); err == nil || !strings.Contains(err.Error(), "timestamps must be UTC") {
					t.Fatalf("accepted %q or wrong error: %s: %v", value, statement, err)
				}
			}
		})
	}
	for _, value := range []any{"2026-10-17T12:34:56Z", "2024-02-29T00:00:00Z", "", nil} {
		if _, err := db.Exec(`UPDATE users SET deleted_at=? WHERE id=1`, value); err != nil {
			t.Fatalf("rejected valid optional timestamp %v: %v", value, err)
		}
	}
	for _, value := range []string{"2026-10-17", "2024-02-29", "2026-10-17T12:34:56Z", ""} {
		if _, err := db.Exec(`UPDATE calendar_events SET event_date=? WHERE id=1`, value); err != nil {
			t.Fatal(value, err)
		}
	}
	for _, value := range []string{"2026-02-30", "2026-10-17T12:34", "2026-10-17T12:34:56+00:00"} {
		if _, err := db.Exec(`UPDATE calendar_events SET event_date=? WHERE id=1`, value); err == nil {
			t.Fatalf("accepted invalid event date %q", value)
		}
	}
	// Ensure every selected column's table has both enforcement triggers.
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	rules, err := utcRules(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}
	for table := range rules {
		var count int
		if err := tx.QueryRow(`SELECT count(*) FROM sqlite_schema WHERE type='trigger' AND name IN (?,?)`, "trg_utc_"+table+"_insert", "trg_utc_"+table+"_update").Scan(&count); err != nil || count != 2 {
			t.Fatalf("missing enforcement on %s: %d %v", table, count, err)
		}
	}
}

func TestUTCEnforcementMigrationRejectsExistingInvalidValuesAtomically(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "migration.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	source, err := fs.Sub(migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, source, goose.WithGoMigrations(utcEventMigration(), utcEnforcementMigration()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(context.Background(), 15); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users(id,email,name,password,created_at) VALUES(1,'old@example.org','Name','hash','2026-10-17T12:00:00+02:00')`); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(context.Background()); err == nil || !strings.Contains(err.Error(), "users.created_at") {
		t.Fatalf("did not identify invalid existing timestamp: %v", err)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_schema WHERE type='trigger' AND name GLOB 'trg_utc_*'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("failed migration left partial enforcement: %d %v", count, err)
	}
	if _, err := db.Exec(`UPDATE users SET created_at='2026-10-17T10:00:00Z' WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE users SET created_at='2026-10-17' WHERE id=1`); err == nil {
		t.Fatal("corrected migration did not enforce UTC")
	}
}
