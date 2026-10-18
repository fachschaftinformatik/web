package events

import (
	"bytes"
	"context"
	"errors"
	"image"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fachschaftinformatik/web/internal/publicid"
	"github.com/fachschaftinformatik/web/model/media"
)

func TestImportPreservesCaptionsFullImagesAndEditorialEdits(t *testing.T) {
	s, db, root := eventService(t)
	ctx := context.Background()
	photoData := eventPhoto(t, 16, 24)
	input := ImportedEvent{SourceURL: "https://www.instagram.com/p/Example_123/", PublishedAt: time.Date(2025, 11, 7, 12, 0, 0, 0, time.UTC), Status: "recap", Draft: Draft{Title: "Lerncafé", Body: "Original caption ☕\n\n<script>Literal text</script>"}, Images: []CoverUpload{{Data: coverPhoto(t), Alt: "Gemeinsam lernen"}, {Data: photoData, Alt: "Kaffee im Lerncafé"}, {Data: photoData, Alt: "Im Lerncafé"}}}
	if _, _, err := s.Import(ctx, 3, input); !errors.Is(err, ErrForbidden) {
		t.Fatal("ordinary member imported event")
	}
	event, created, err := s.Import(ctx, 1, input)
	if err != nil || !created {
		t.Fatalf("import: %v %v", created, err)
	}
	e, err := s.Lookup(ctx, event)
	if err != nil {
		t.Fatal(err)
	}
	if e.Body != input.Draft.Body || e.EventDate != "" || e.SourcePublishedAt != input.PublishedAt.Format(time.RFC3339) || e.SourceURL != input.SourceURL || !e.AlbumID.Valid {
		t.Fatalf("imported event: %+v", e)
	}
	media := media.NewService(db, s.storage)
	albums, err := media.Catalog(ctx)
	if err != nil || len(albums) != 1 || len(albums[0].Images) != 3 {
		t.Fatalf("media: %+v %v", albums, err)
	}
	photo := albums[0].Images[0]
	if !publicid.Valid(albums[0].ID) || !publicid.Valid(photo.ID) || e.AlbumPublicID != albums[0].ID || e.AlbumCoverPublicID != photo.ID {
		t.Fatalf("invalid cross-domain public references: event=%+v album=%+v", e, albums[0])
	}
	object, err := media.Open(ctx, photo.ID, "original")
	if err != nil {
		t.Fatal(err)
	}
	original, err := io.ReadAll(object)
	object.Close()
	if err != nil || !bytes.Equal(original, input.Images[0].Data) {
		t.Fatal("original bytes changed")
	}
	decoded, _, err := image.DecodeConfig(bytes.NewReader(original))
	if err != nil || decoded.Width != 800 || decoded.Height != 1800 {
		t.Fatalf("full-sized photograph lost: %+v %v", decoded, err)
	}
	if _, err := db.Exec(`UPDATE calendar_events SET title='Editorial title' WHERE id=?`, e.InternalID); err != nil {
		t.Fatal(err)
	}
	other, created, err := s.Import(ctx, 2, input)
	if err != nil || created || other != event {
		t.Fatalf("idempotence: %s %v %v", other, created, err)
	}
	e, _ = s.Lookup(ctx, event)
	albums, _ = media.Catalog(ctx)
	if e.Title != "Editorial title" || len(albums) != 1 || len(albums[0].Images) != 3 {
		t.Fatal("repeat import overwrote edits or duplicated photos")
	}
	before := 0
	filepath.WalkDir(root, func(_ string, e os.DirEntry, err error) error {
		if err == nil && !e.IsDir() {
			before++
		}
		return err
	})
	if _, err := db.Exec(`CREATE TRIGGER fail_import_photo BEFORE INSERT ON media BEGIN SELECT RAISE(ABORT,'test rollback'); END;`); err != nil {
		t.Fatal(err)
	}
	input.SourceURL = "https://www.instagram.com/p/Another_123/"
	input.Images[0].Data = photoData
	if _, _, err := s.Import(ctx, 1, input); err == nil {
		t.Fatal("failed transaction imported event")
	}
	after := 0
	filepath.WalkDir(root, func(_ string, e os.DirEntry, err error) error {
		if err == nil && !e.IsDir() {
			after++
		}
		return err
	})
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM calendar_events`).Scan(&count); err != nil || count != 1 || before != after {
		t.Fatalf("failed import changed data or storage: events=%d objects=%d->%d %v", count, before, after, err)
	}
	input.SourceURL = "https://example.org/p/Example_123/"
	if _, _, err := s.Import(ctx, 1, input); err == nil {
		t.Fatal("accepted untrusted source URL")
	}
	if _, err := db.Exec(`UPDATE media SET deleted_at='2026-01-01T00:00:00Z' WHERE id=(SELECT MAX(id) FROM media WHERE event_id=?)`, e.AlbumID.Int64); err != nil {
		t.Fatal(err)
	}
	e, err = s.Lookup(ctx, event)
	if err != nil || e.AlbumID.Valid || e.CoverKey == "" {
		t.Fatalf("two-image album linked or event cover lost: %+v %v", e, err)
	}
	if _, err := db.Exec(`UPDATE calendar_events SET deleted_at='2026-01-01T00:00:00Z' WHERE id=?`, e.InternalID); err != nil {
		t.Fatal(err)
	}
	object, err = media.Open(ctx, photo.ID, "original")
	if err != nil {
		t.Fatal("deleting a calendar event removed its independent photo album")
	}
	object.Close()
}

func TestImportNeverReusesDeletedHighestResourceKeys(t *testing.T) {
	s, db, _ := eventService(t)
	ctx := context.Background()
	input := ImportedEvent{
		SourceURL: "https://www.instagram.com/p/First/", PublishedAt: time.Date(2025, 11, 7, 12, 0, 0, 0, time.UTC),
		Draft:  Draft{Title: "Imported event", Body: "An imported event caption."},
		Images: []CoverUpload{{Data: eventPhoto(t, 16, 24), Alt: "A photograph"}},
	}
	input.Images = append(input.Images, input.Images[0], input.Images[0])
	first, _, err := s.Import(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	old, err := s.Lookup(ctx, first)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM calendar_events WHERE id=?`, old.InternalID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM media WHERE event_id=?`, old.AlbumID.Int64); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM events WHERE id=?`, old.AlbumID.Int64); err != nil {
		t.Fatal(err)
	}
	input.SourceURL = "https://www.instagram.com/p/Next/"
	next, _, err := s.Import(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	current, err := s.Lookup(ctx, next)
	if err != nil {
		t.Fatal(err)
	}
	if first == next || old.InternalID == current.InternalID || old.AlbumID == current.AlbumID || old.AlbumPublicID == current.AlbumPublicID || current.AlbumCoverID.Int64 <= old.AlbumCoverID.Int64+2 || old.AlbumCoverPublicID == current.AlbumCoverPublicID {
		t.Fatalf("deleted imported identities were reused: old=%+v next=%+v", old, current)
	}
	if _, err := s.Save(ctx, 1, next, input.Draft, "", nil); err != nil {
		t.Fatal("existing imported event update:", err)
	}
}
