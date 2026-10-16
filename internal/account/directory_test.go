package account

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestReadOnlyUserSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "users.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE programs (id INTEGER, name TEXT); CREATE TABLE users (id INTEGER, name TEXT, email TEXT, role TEXT, active INTEGER, verified INTEGER, program_id INTEGER, deleted_at TEXT, password TEXT, verification_token TEXT); INSERT INTO programs VALUES (1,'Informatik'); INSERT INTO users VALUES (42,'Kim Beispiel','kim@studmail.w-hs.de','editor',0,1,1,NULL,'never-import-this-password','never-import-this-token'), (43,'Gelöscht','deleted@example.org','user',0,0,NULL,'2026-01-01','','');`)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	users, err := LoadUsers(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].ID != "42" || users[0].Permission != "Redaktion" || users[0].Program != "Informatik" || users[0].Status != "Inaktiv" || !users[0].Verified {
		t.Fatalf("unexpected display snapshot: %+v", users)
	}
	if len(SearchUsers(users, " KIM@STUDMAIL ")) != 1 || len(SearchUsers(users, "absent")) != 0 {
		t.Fatal("search does not match email case-insensitively")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("loading users changed the source database")
	}
}
