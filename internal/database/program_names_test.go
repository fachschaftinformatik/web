package database

import (
	"context"
	"database/sql"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"testing"

	"github.com/fachschaftinformatik/web/internal/publicid"
	"github.com/pressly/goose/v3"
)

func TestProgramNamesMigrationPreservesQualificationsAndIdentities(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "programs.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	source, err := fs.Sub(migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, source, goose.WithGoMigrations(utcEnforcementMigration(), publicIDsMigration(), publicIDSequencesMigration()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO programs(id,name) VALUES(1,'Informatik (B. Sc.)'),(2,'Informatik (M.Sc.)');
	INSERT INTO study_programs(id,name) VALUES(10,'Informatik');
	INSERT INTO program_qualifications(legacy_program_id,study_program_id,degree) VALUES(1,10,'Bachelor'),(2,10,'Master');
	INSERT INTO program_versions(program_id,name) VALUES(1,'2023'),(2,'2024');
	INSERT INTO modules(id,program_id,name) VALUES(11,1,'Grundlagen'),(12,2,'Vertiefung');
	INSERT INTO users(id,email,name,password,program_id) VALUES(1,'test@example.org','Person','unchanged',2);`); err != nil {
		t.Fatal(err)
	}
	var refs []string
	for _, id := range []string{"1", "2"} {
		value, err := publicid.Ensure(ctx, db, "programs", id)
		if err != nil {
			t.Fatal(err)
		}
		refs = append(refs, value)
	}
	var triggers int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_schema WHERE type='trigger' AND tbl_name='programs'`).Scan(&triggers); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	for i, id := range []string{"1", "2"} {
		var name string
		if err := db.QueryRow(`SELECT name FROM programs WHERE id=?`, id).Scan(&name); err != nil || name != "Informatik" {
			t.Fatal("program name still includes its degree", name, err)
		}
		if value, err := publicid.Lookup(ctx, db, "programs", id); err != nil || value != refs[i] {
			t.Fatal("program public identity changed", value, err)
		}
	}
	for _, query := range []string{
		`SELECT count(*)=2 FROM program_qualifications WHERE (legacy_program_id=1 AND degree='Bachelor') OR (legacy_program_id=2 AND degree='Master')`,
		`SELECT count(*)=2 FROM modules WHERE (id=11 AND program_id=1) OR (id=12 AND program_id=2)`,
		`SELECT count(*)=2 FROM program_versions WHERE (program_id=1 AND name='2023') OR (program_id=2 AND name='2024')`,
		`SELECT program_id=2 AND password='unchanged' FROM users WHERE id=1`,
	} {
		var ok bool
		if err := db.QueryRow(query).Scan(&ok); err != nil || !ok {
			t.Fatal("qualification or relationship changed", query, err)
		}
	}
	var after int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_schema WHERE type='trigger' AND tbl_name='programs'`).Scan(&after); err != nil || after != triggers {
		t.Fatal("program triggers changed", triggers, after, err)
	}
	if _, err := db.Exec(`UPDATE programs SET updated_at='2026-10-18T02:00:00+02:00' WHERE id=1`); err == nil {
		t.Fatal("UTC enforcement lost")
	}
	if _, err := db.Exec(`INSERT OR REPLACE INTO programs(id,name) VALUES(1,'Replacement')`); err == nil {
		t.Fatal("public ID replacement protection lost")
	}
	rows, err := db.Query(`PRAGMA foreign_key_check`)
	if err != nil {
		t.Fatal(err)
	}
	if rows.Next() {
		t.Fatal("foreign key integrity changed")
	}
	rows.Close()
	if err := Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal("migration restart", err)
	}
}
