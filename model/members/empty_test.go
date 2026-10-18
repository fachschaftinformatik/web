package members

import (
	"context"
	"github.com/fachschaftinformatik/web/internal/database"
	"io"
	"log"
	"path/filepath"
	"testing"
)

func TestOpenNeverCreatesMembersOrAssignments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.db")
	if err := database.Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	db, err := database.NewConnection(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for range 2 {
		s, err := Open(context.Background(), db)
		if err != nil {
			t.Fatal(err)
		}
		if len(s.Members()) != 0 || len(s.Roles("board")) != 0 {
			t.Fatal("opening the store inserted organization data")
		}
	}
}
