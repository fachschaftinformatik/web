package main

import (
	"database/sql"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/fachschaftinformatik/web/internal/database"
)

func TestImportPreservesExistingRecordsAndVerifiesObjects(t *testing.T) {
	root := t.TempDir()
	path, blobs := filepath.Join(root, "site.db"), filepath.Join(root, "blobs")
	if err := database.Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT INTO organization(id,data) VALUES(1,'{"Name":"Existing organization"}')`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := importData(path, blobs); err != nil {
			t.Fatal(err)
		}
	}
	var settings string
	if err := db.QueryRow(`SELECT data FROM organization WHERE id=1`).Scan(&settings); err != nil || settings != `{"Name":"Existing organization"}` {
		t.Fatal("existing settings changed", settings, err)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM historical_meeting_documents`).Scan(&count); err != nil || count != 48 {
		t.Fatal("missing or duplicate documents", count, err)
	}
	var key string
	if err := db.QueryRow(`SELECT blob_key FROM historical_meeting_documents LIMIT 1`).Scan(&key); err != nil {
		t.Fatal(err)
	}
	object := filepath.Join(blobs, filepath.FromSlash(key))
	if err := os.WriteFile(object, []byte("different bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := importData(path, blobs); err == nil {
		t.Fatal("overwrote a conflicting object")
	}
	data, err := os.ReadFile(object)
	if err != nil || string(data) != "different bytes" {
		t.Fatal("changed conflicting object", err)
	}
	if _, err := db.Exec(`DELETE FROM organization`); err != nil {
		t.Fatal(err)
	}
	// Verification of all published metadata must happen inside the transaction,
	// including records that were already in the target database.
	if _, err := db.Exec(`UPDATE historical_meeting_documents SET bytes=bytes+1`); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(object); err != nil {
		t.Fatal(err)
	}
	if err := importData(path, blobs); err == nil {
		t.Fatal("accepted mismatched published metadata")
	}
	if err := db.QueryRow(`SELECT count(*) FROM organization`).Scan(&count); err != nil || count != 0 {
		t.Fatal("failed verification published settings", count, err)
	}
}
