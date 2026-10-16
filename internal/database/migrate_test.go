package database

import (
	"database/sql"
	"io"
	"log"
	"path/filepath"
	"testing"
)

func TestEmbeddedMigrationsPreserveDataOnRestart(t *testing.T) {
	// The executable must migrate without SQL files in its working directory.
	t.Chdir(t.TempDir())
	dsn := filepath.Join(t.TempDir(), "sqlite.db")
	logger := log.New(io.Discard, "", 0)
	if err := Migrate(dsn, logger); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT INTO users (id, email, name, password, password_reset_token) VALUES (42, 'kim@studmail.w-hs.de', 'Kim', 'existing-hash', 'existing-token')`); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(dsn, logger); err != nil {
		t.Fatal(err)
	}
	var password, token string
	if err := db.QueryRow(`SELECT password, password_reset_token FROM users WHERE id = 42`).Scan(&password, &token); err != nil {
		t.Fatal(err)
	}
	if password != "existing-hash" || token != "existing-token" {
		t.Fatal("reapplying migrations changed an existing account")
	}
}
