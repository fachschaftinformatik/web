package events

import (
	"context"
	"database/sql"
	"github.com/fachschaftinformatik/web/internal/publicid"
	"testing"
)

func seedReferences(t *testing.T, db *sql.DB) {
	t.Helper()
	for namespace, table := range map[string]string{"events": "calendar_events", "posts": "discussion_posts", "albums": "events", "media": "media", "users": "users"} {
		rows, err := db.Query("SELECT CAST(id AS TEXT) FROM " + table)
		if err != nil {
			t.Fatal(err)
		}
		var ids []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				t.Fatal(err)
			}
			ids = append(ids, id)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		rows.Close()
		for _, id := range ids {
			if _, err := publicid.Ensure(context.Background(), db, namespace, id); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func reference(t *testing.T, db *sql.DB, namespace, id string) string {
	t.Helper()
	seedReferences(t, db)
	ref, err := publicid.Lookup(context.Background(), db, namespace, id)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}
