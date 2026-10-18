package publicid

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestCanonicalEncoding(t *testing.T) {
	for range 1000 {
		value := New()
		if !Valid(value) {
			t.Fatalf("generated invalid ID %q", value)
		}
	}
	for _, value := range []string{"", "1", "12345678901", "AAAAAAAAAAA=", "AAAAAAAAAAB", "AAAAAAAAAA/", "AAAAAAAAAA+", "AAAAAAAAAA\n"} {
		if Valid(value) {
			t.Errorf("accepted noncanonical ID %q", value)
		}
	}
	if !Valid("AAAAAAAAAAA") || !Valid("__________8") {
		t.Fatal("rejected canonical boundary encodings")
	}
}

func TestAllocationRetriesCollisionsAndRemainsStable(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE public_ids(namespace TEXT,record_id TEXT,public_id TEXT,retired INTEGER NOT NULL DEFAULT 0,UNIQUE(namespace,record_id),UNIQUE(namespace,public_id))`)
	if err != nil {
		t.Fatal(err)
	}
	first, err := Ensure(ctx, db, "posts", "42")
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	second, err := ensure(ctx, db, "posts", "43", func() string {
		calls++
		if calls == 1 {
			return first
		}
		return "AAAAAAAAAAA"
	})
	if err != nil || second == first || calls != 2 {
		t.Fatalf("collision retry: %q %d %v", second, calls, err)
	}
	again, err := Ensure(ctx, db, "posts", "42")
	if err != nil || again != first {
		t.Fatalf("unstable allocation: %q %v", again, err)
	}
	if _, err := Resolve(ctx, db, "posts", "42"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("internal ID resolved", err)
	}
	if _, err := Resolve(ctx, db, "albums", first); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("cross-namespace resolution", err)
	}
	if _, err := db.Exec(`UPDATE public_ids SET retired=1 WHERE record_id='42'`); err != nil {
		t.Fatal(err)
	}
	if _, err := Ensure(ctx, db, "posts", "42"); !errors.Is(err, ErrRetired) {
		t.Fatal("reused retired record", err)
	}
	if _, err := Resolve(ctx, db, "posts", first); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("resolved retired ID", err)
	}
	if _, err := Lookup(ctx, db, "posts", "42"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("looked up retired ID", err)
	}
	_, err = ensure(ctx, db, "posts", "44", func() string { return first })
	if err == nil || !strings.Contains(err.Error(), "collision") {
		t.Fatal("collision exhaustion", err)
	}
}
