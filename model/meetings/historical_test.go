package meetings

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"testing"

	"github.com/fachschaftinformatik/web/internal/publicid"
)

func TestHistoricalIDsResolveOriginalObjectsAcrossRestarts(t *testing.T) {
	s, blobs, reopen := persistentFixture(t)
	ctx := context.Background()
	files := []string{"same.pdf", "same.PDF", "historical-ä.pdf"}
	for _, filename := range files {
		if _, err := s.db.Exec(`INSERT INTO historical_meeting_documents(file,year,number,date,kind,note,bytes,sha256,sources,blob_key) VALUES(?,2020,1,'2020-01-02','Einladung','',1,'hash','[]',?)`, filename, "original/"+filename); err != nil {
			t.Fatal(err)
		}
		blobs.objects["original/"+filename] = []byte("original bytes: " + filename)
	}
	reads := blobs.gets
	s = reopen()
	if blobs.gets != reads {
		t.Fatal("initializing historical IDs read PDF objects")
	}
	seen := make(map[string]bool)
	ids := make(map[string]string)
	for _, filename := range files {
		id, err := publicid.Lookup(ctx, s.db, "historical_documents", filename)
		if err != nil {
			t.Fatal(err)
		}
		ids[filename] = id
		if !publicid.Valid(id) || seen[id] {
			t.Fatalf("invalid or duplicate historical ID %q", id)
		}
		seen[id] = true
		object, err := s.HistoricalDocumentByID(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(object)
		err = errors.Join(err, object.Close())
		if err != nil || !bytes.Equal(data, blobs.objects["original/"+filename]) {
			t.Fatalf("original bytes changed for %q: %q, %v", filename, data, err)
		}
	}
	if blobs.gets != reads+len(files) {
		t.Fatal("resolving IDs scanned unrelated objects")
	}
	for _, id := range []string{"same", "same.pdf", "h", publicid.New()} {
		if _, err := s.HistoricalDocumentByID(ctx, id); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("unknown ID %q: %v", id, err)
		}
	}
	object, err := reopen().HistoricalDocumentByID(ctx, ids["same.pdf"])
	if err != nil {
		t.Fatalf("historical ID changed across restart: %v", err)
	}
	object.Close()
}
