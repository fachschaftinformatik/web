package testdb

import (
	"testing"

	"github.com/fachschaftinformatik/web/internal/database"
)

func TestDatabasesAreIsolated(t *testing.T) {
	first, err := database.NewConnection(Path(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { first.Close() })
	if _, err := first.Exec(`INSERT INTO programs(id,name) VALUES(1001,'Isolated test program')`); err != nil {
		t.Fatal(err)
	}
	second, err := database.NewConnection(Path(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { second.Close() })
	var count int
	if err := second.QueryRow(`SELECT count(*) FROM programs WHERE id=1001`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("database fixtures share mutable state")
	}
}
