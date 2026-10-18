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

func TestProgramVersionsMigrationStoresNumericYears(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "years.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	source, err := fs.Sub(migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, source, goose.WithGoMigrations(utcEnforcementMigration(), publicIDsMigration(), publicIDSequencesMigration(), programNamesMigration()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(ctx, 19); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO programs(id,name) VALUES(1,'Informatik');
	INSERT INTO program_versions(program_id,name) VALUES(1,'PO2016'),(1,'PO2023');
	INSERT INTO modules(id,program_id,name) VALUES(10,1,'Grundlagen');
	INSERT INTO users(id,email,name,password,program_id) VALUES(1,'test@example.org','Person','unchanged',1);
	INSERT INTO archive(id,group_id,user_id,module_id,version,exam_date,access_key,mime_type,nbytes,checksum)
	VALUES(100,100,1,10,'PO2023','2026-01-01','original/object/key','application/pdf',1234,'unchanged');
	INSERT INTO archive_assignments(archive_id,module_id,version) VALUES(100,10,'PO2023');
	INSERT INTO user_archive_preferences(user_id,programs,degrees,versions) VALUES(1,'["legacy"]','["bachelor"]','["PO2016","PO2023"]');`); err != nil {
		t.Fatal(err)
	}
	ref, err := publicid.Ensure(ctx, db, "documents", "100")
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{
		`SELECT count(*)=2 AND min(name)=2016 AND max(name)=2023 AND min(typeof(name))='integer' FROM program_versions`,
		`SELECT version=2023 AND typeof(version)='integer' AND checksum='unchanged' AND access_key='original/object/key' AND nbytes=1234 FROM archive WHERE id=100`,
		`SELECT version=2023 AND typeof(version)='integer' FROM archive_assignments WHERE archive_id=100`,
		`SELECT count(*)=1 FROM archive_placements b JOIN program_versions p ON p.program_id=1 AND p.name=b.version WHERE b.archive_id=100`,
		`SELECT versions='["2016","2023"]' AND programs='["legacy"]' AND degrees='["bachelor"]' FROM user_archive_preferences WHERE user_id=1`,
	} {
		var ok bool
		if err := db.QueryRow(query).Scan(&ok); err != nil || !ok {
			t.Fatal("numeric year migration broke a reference", query, err)
		}
	}
	if value, err := publicid.Lookup(ctx, db, "documents", "100"); err != nil || value != ref {
		t.Fatal("document public identity changed", value, err)
	}
	for _, query := range []string{
		`INSERT INTO program_versions(program_id,name) VALUES(1,'PO2024')`,
		`UPDATE archive SET version='PO2023' WHERE id=100`,
		`UPDATE archive_assignments SET version=23 WHERE archive_id=100`,
		`UPDATE archive SET updated_at='2026-01-01T10:00:00+01:00' WHERE id=100`,
		`INSERT OR REPLACE INTO archive SELECT * FROM archive WHERE id=100`,
	} {
		if _, err := db.Exec(query); err == nil {
			t.Fatal("numeric year, UTC or public identity constraint lost", query)
		}
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
