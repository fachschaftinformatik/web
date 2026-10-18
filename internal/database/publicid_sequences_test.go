package database

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"io/fs"
	"log"
	"math"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/fachschaftinformatik/web/internal/publicid"
	"github.com/pressly/goose/v3"
)

func TestIntegerSequencePreservesDeletedAndExplicitKeys(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "sequences.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	source, err := fs.Sub(migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, source, goose.WithGoMigrations(utcEnforcementMigration(), publicIDsMigration()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	const legacy int64 = 987654321012345
	if _, err := db.Exec(`INSERT INTO events(id,title) VALUES(?,'Deleted legacy album')`, legacy); err != nil {
		t.Fatal(err)
	}
	oldPublic, err := publicid.Ensure(ctx, db, "albums", strconv.FormatInt(legacy, 10))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM events WHERE id=?`, legacy); err != nil {
		t.Fatal(err)
	}
	// An unregistered explicit import must also advance the sequence baseline.
	if _, err := db.Exec(`INSERT INTO events(id,title) VALUES(?,'Raw import')`, legacy-1); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	var sequenceCount int
	if err := db.QueryRow(`SELECT count(*) FROM public_id_sequences`).Scan(&sequenceCount); err != nil || sequenceCount != 9 {
		t.Fatal("incorrect INTEGER namespace discovery", sequenceCount, err)
	}
	create := func(title string) (int64, string) {
		t.Helper()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		id, err := publicid.NextInteger(ctx, tx, "albums")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO events(id,title) VALUES(?,?)`, id, title); err != nil {
			t.Fatal(err)
		}
		value, err := publicid.Ensure(ctx, tx, "albums", strconv.FormatInt(id, 10))
		if err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
		return id, value
	}
	first, firstPublic := create("First")
	if first != legacy+1 {
		t.Fatalf("lost retired high-water mark: %d", first)
	}
	if _, err := db.Exec(`DELETE FROM events WHERE id=?`, first); err != nil {
		t.Fatal(err)
	}
	second, secondPublic := create("Second")
	if second != first+1 || secondPublic == firstPublic || secondPublic == oldPublic {
		t.Fatal("reused retired identity", second, secondPublic)
	}
	for _, value := range []string{oldPublic, firstPublic} {
		if _, err := publicid.Resolve(ctx, db, "albums", value); !errors.Is(err, sql.ErrNoRows) {
			t.Fatal("resolved old identity", err)
		}
	}
	// UPSERT is an UPDATE and must retain identity; REPLACE must never recreate
	// a registered identity even with SQLite's default recursive_triggers=OFF.
	if _, err := db.Exec(`INSERT INTO events(id,title) VALUES(?,'Updated') ON CONFLICT(id) DO UPDATE SET title=excluded.title`, second); err != nil {
		t.Fatal("upsert rejected", err)
	}
	for _, id := range []int64{first, second} {
		if _, err := db.Exec(`INSERT OR REPLACE INTO events(id,title) VALUES(?,'Replacement')`, id); err == nil {
			t.Fatal("replaced reserved record", id)
		}
	}
	var title string
	if err := db.QueryRow(`SELECT title FROM events WHERE id=?`, second).Scan(&title); err != nil || title != "Updated" {
		t.Fatal("replacement changed resource", title, err)
	}
	if value, err := publicid.Lookup(ctx, db, "albums", strconv.FormatInt(second, 10)); err != nil || value != secondPublic {
		t.Fatal("replacement changed mapping", value, err)
	}
	const explicit = legacy + 100
	if _, err := db.Exec(`INSERT INTO events(id,title) VALUES(?,'Explicit')`, explicit); err != nil {
		t.Fatal(err)
	}
	third, _ := create("After explicit import")
	if third != explicit+1 {
		t.Fatal("explicit insert not tracked", third)
	}
	if err := Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	fourth, _ := create("After restart")
	if fourth != third+1 {
		t.Fatal("restart reset sequence", fourth)
	}
	// Allocation is transactional; rolled-back allocations are not published.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rolledBack, err := publicid.NextInteger(ctx, tx, "albums")
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	fifth, _ := create("After rollback")
	if fifth != rolledBack {
		t.Fatal("allocation did not roll back", fifth, rolledBack)
	}
	for _, statement := range []string{
		`UPDATE public_id_sequences SET last_id=0 WHERE namespace='albums'`,
		`DELETE FROM public_id_sequences WHERE namespace='albums'`,
	} {
		if _, err := db.Exec(statement); err == nil {
			t.Fatal("released integer reservations", statement)
		}
	}
	if _, err := db.Exec(`INSERT OR REPLACE INTO public_id_sequences(namespace,last_id) VALUES('albums',0)`); err != nil {
		t.Fatal(err)
	}
	var highWater int64
	if err := db.QueryRow(`SELECT last_id FROM public_id_sequences WHERE namespace='albums'`).Scan(&highWater); err != nil || highWater != fifth {
		t.Fatal("replace reset sequence", highWater, err)
	}
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := publicid.NextInteger(ctx, tx, "submissions"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("allocated integer for TEXT namespace", err)
	}
	if _, err := tx.Exec(`UPDATE public_id_sequences SET last_id=? WHERE namespace='albums'`, int64(math.MaxInt64-1)); err != nil {
		t.Fatal(err)
	}
	if id, err := publicid.NextInteger(ctx, tx, "albums"); err != nil || id != math.MaxInt64 {
		t.Fatal("lost final integer", id, err)
	}
	if _, err := publicid.NextInteger(ctx, tx, "albums"); !errors.Is(err, publicid.ErrIntegerExhausted) {
		t.Fatal("overflow not guarded", err)
	}
	var storage string
	if err := tx.QueryRow(`SELECT typeof(last_id) FROM public_id_sequences WHERE namespace='albums'`).Scan(&storage); err != nil || storage != "integer" {
		t.Fatal("sequence overflowed", storage, err)
	}
}
