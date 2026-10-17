package data

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
	"io"
	"io/fs"
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

func TestArchiveMigrationPreservesAliasesHistoriesAndQualifications(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "archive.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	source, err := fs.Sub(migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(context.Background(), 8); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO programs(id,name) VALUES(1,'Informatik (B. Sc.)'),(2,'Wirtschaftsinformatik (B. Sc.)'),(3,'Informatik (M. Sc.)');
    INSERT INTO program_versions(program_id,name) VALUES(1,'PO2016'),(2,'PO2023'),(3,'PO2016');
    INSERT INTO modules(id,program_id,name) VALUES(10,1,'Mathematik'),(20,2,'Mathematik'),(30,3,'Mathematik');
    INSERT INTO users(id,email,name,password) VALUES(1,'migration@fsv-wh.de','Name','unchanged');
    INSERT INTO archive(id,group_id,user_id,module_id,version,exam_date,edit_version,is_latest,access_key,mime_type,nbytes,checksum) VALUES
    (100,100,1,10,'PO2016','2026-01-01',1,0,'a','application/pdf',5,'old'),
    (101,100,1,10,'PO2016','2026-01-01',2,1,'b','application/pdf',6,'same'),
    (200,200,1,20,'PO2023','2026-01-01',1,0,'c','application/pdf',5,'old'),
    (201,200,1,20,'PO2023','2026-01-01',2,1,'d','application/pdf',6,'same'),
    (300,300,1,30,'PO2016','2026-01-01',1,1,'e','application/pdf',6,'same');`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatal(err)
	}
	var count int
	for query, want := range map[string]int{
		"SELECT COUNT(*) FROM archive":                                                        5,
		"SELECT COUNT(*) FROM archive_objects":                                                2,
		"SELECT COUNT(*) FROM archive_entry_aliases WHERE legacy_id!=canonical_id":            2,
		"SELECT COUNT(*) FROM archive_assignments WHERE archive_id=101":                       2,
		"SELECT COUNT(*) FROM study_programs":                                                 2,
		"SELECT COUNT(*) FROM program_qualifications WHERE degree='Bachelor'":                 2,
		"SELECT COUNT(*) FROM archive_entry_aliases WHERE legacy_id=300 AND canonical_id=300": 1,
		"SELECT COUNT(*) FROM archive WHERE is_latest=1":                                      3,
	} {
		if err := db.QueryRow(query).Scan(&count); err != nil || count != want {
			t.Fatalf("%s: %d want %d: %v", query, count, want, err)
		}
	}
	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestCalendarMigrationPreservesLegacyDiscussions(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "calendar.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	source, err := fs.Sub(migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(context.Background(), 9); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO users(id,email,name,password) VALUES(1,'calendar@fsv-wh.de','Name','unchanged');
	INSERT INTO discussion_posts(id,title,body,user_id,type,event_date,location,deleted_at) VALUES
	(10,'Grillabend','Original body',1,'event','2026-06-09T17:00','Innenhof',NULL),
	(20,'Diskussion','Original discussion',1,'discussion',NULL,NULL,NULL),
	(30,'Gelöscht','Deleted',1,'event',NULL,NULL,'deleted');
	INSERT INTO discussion_comments(id,post_id,user_id,text) VALUES(100,10,1,'Original comment');
	INSERT INTO discussion_votes(post_id,user_id,vote) VALUES(10,1,1);`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatal(err)
	}
	for query, want := range map[string]int{
		`SELECT COUNT(*) FROM calendar_events`:                          2,
		`SELECT COUNT(*) FROM calendar_events WHERE deleted_at IS NULL`: 1,
		`SELECT COUNT(*) FROM calendar_events WHERE id=10 AND legacy_post_id=10 AND body='Original body' AND event_date='2026-06-09T17:00' AND location='Innenhof'`: 1,
		`SELECT COUNT(*) FROM discussion_posts`:                                                 3,
		`SELECT COUNT(*) FROM discussion_comments WHERE post_id=10 AND text='Original comment'`: 1,
		`SELECT COUNT(*) FROM discussion_votes WHERE post_id=10 AND vote=1`:                     1,
	} {
		var count int
		if err := db.QueryRow(query).Scan(&count); err != nil || count != want {
			t.Fatalf("%s: %d %v", query, count, err)
		}
	}
	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatal(err)
	}
}
