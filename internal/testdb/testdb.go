package testdb

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/fachschaftinformatik/web/internal/database"
)

var schema = sync.OnceValues(func() ([]byte, error) {
	dir, err := os.MkdirTemp("", "web-test-schema-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "sqlite.db")
	if err := database.Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		return nil, err
	}
	return os.ReadFile(path)
})

// Path gives each test an isolated copy of the real migrated database.
func Path(t testing.TB) string {
	t.Helper()
	data, err := schema()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "sqlite.db")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
