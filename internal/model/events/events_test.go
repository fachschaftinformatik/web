package events

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/storage"
	"github.com/fachschaftinformatik/web/internal/storage/disk"
)

func eventService(t *testing.T) (*Service, *sql.DB, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "events.db")
	if err := database.Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	db, err := database.NewConnection(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec(`INSERT INTO users(id,email,name,password,role,active,verified) VALUES
	(1,'admin@fsv-wh.de','Admin','unused','admin',1,1),(2,'editor@fsv-wh.de','Editor','unused','editor',1,1),
	(3,'user@fsv-wh.de','User','unused','user',1,1),(4,'inactive@fsv-wh.de','Inactive','unused','admin',0,1),
	(5,'unverified@fsv-wh.de','Unverified','unused','editor',1,0),(6,'deleted@fsv-wh.de','Deleted','unused','editor',1,1);
	UPDATE users SET deleted_at='deleted' WHERE id=6;`)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	storage, err := disk.New(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { storage.Close() })
	return NewService(db, storage), db, root
}
func coverPhoto(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	i := image.NewNRGBA(image.Rect(0, 0, 800, 1800))
	i.Set(0, 0, color.NRGBA{R: 255, A: 255})
	if err := png.Encode(&b, i); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestStandaloneEventsPersistCoversAndRestrictPublishing(t *testing.T) {
	s, db, _ := eventService(t)
	ctx := context.Background()
	d := Draft{Title: "Grillabend", Body: "Gemeinsam grillen im Innenhof.", Date: "30.01.2027", Time: "17:00", Location: "Innenhof"}
	for _, user := range []int64{0, 3, 4, 5, 6, 999} {
		if _, err := s.Save(ctx, user, "", d, "", nil); !errors.Is(err, ErrForbidden) {
			t.Fatalf("user %d: %v", user, err)
		}
	}
	value, err := s.Save(ctx, 1, "", d, `<p><strong>Gemeinsam grillen</strong></p><script>bad()</script>`, &CoverUpload{Data: coverPhoto(t), Alt: "Ein Grill im Innenhof"})
	if err != nil {
		t.Fatal(err)
	}
	e, err := s.Lookup(ctx, value)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(e.Body, "bad()") || e.EventDate != "2027-01-30T17:00" || e.CoverWidth != 711 || e.CoverHeight != 1600 || e.LegacyURL() != "" {
		t.Fatalf("event: %+v", e)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM discussion_posts`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("created forum posts: %d %v", count, err)
	}
	object, err := s.OpenCover(ctx, value)
	if err != nil {
		t.Fatal(err)
	}
	photo, err := jpeg.Decode(object)
	object.Close()
	if err != nil || photo.Bounds().Dy() != 1600 || object.Info.ContentType != "image/jpeg" {
		t.Fatal("stored JPEG preview missing")
	}
	oldKey := e.CoverKey
	d.Title = "Grillabend mit Spielen"
	if _, err := s.Save(ctx, 2, value, d, "", nil); err != nil {
		t.Fatal(err)
	}
	e, err = NewService(db, s.storage).Lookup(ctx, value)
	if err != nil || e.CoverKey != oldKey || e.Title != d.Title {
		t.Fatal("edit lost cover or title")
	}
	if _, err := s.Save(ctx, 3, value, d, "", nil); !errors.Is(err, ErrForbidden) {
		t.Fatal("ordinary member edited event")
	}
	if _, err := s.Save(ctx, 2, value, d, "", &CoverUpload{Data: []byte("<svg></svg>"), Alt: "Bild"}); err == nil {
		t.Fatal("accepted active SVG")
	}
	e, _ = s.Lookup(ctx, value)
	if e.CoverKey != oldKey {
		t.Fatal("invalid upload changed registered image")
	}
	if _, err := s.Save(ctx, 2, value, d, "", &CoverUpload{Data: coverPhoto(t), Alt: "Neues Titelbild"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.storage.GetObject(ctx, oldKey); !errors.Is(err, storage.ErrNotFound) {
		t.Fatal("replaced image was not removed")
	}
	if _, err := s.Save(ctx, 2, value, d, "", &CoverUpload{Remove: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.OpenCover(ctx, value); !errors.Is(err, ErrNotFound) {
		t.Fatal("removed image still public")
	}
	if _, err := db.Exec(`UPDATE calendar_events SET deleted_at='deleted' WHERE id=?`, value); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Lookup(ctx, value); !errors.Is(err, ErrNotFound) {
		t.Fatal("deleted event visible")
	}
	if _, err := s.Save(ctx, 1, value, d, "", nil); !errors.Is(err, ErrNotFound) {
		t.Fatal("edited deleted event")
	}
}

func TestInvalidUploadsAndFailedTransactionsPreserveStorage(t *testing.T) {
	s, db, root := eventService(t)
	ctx := context.Background()
	d := Draft{Title: "Party", Body: "Ein Abend mit Musik und Spielen."}
	for _, cover := range []CoverUpload{{Data: []byte("not an image"), Alt: "Test"}, {Data: bytes.Repeat([]byte{'x'}, MaxCoverBytes+1), Alt: "Test"}, {Data: coverPhoto(t)}, {Data: coverPhoto(t), Remove: true, Alt: "Test"}} {
		if _, err := s.Save(ctx, 1, "", d, "", &cover); err == nil {
			t.Fatal("invalid upload saved")
		}
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_event BEFORE INSERT ON calendar_events BEGIN SELECT RAISE(ABORT,'test rollback'); END;`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Save(ctx, 1, "", d, "", &CoverUpload{Data: coverPhoto(t), Alt: "Test"}); err == nil {
		t.Fatal("failed transaction succeeded")
	}
	files := 0
	if err := filepath.WalkDir(root, func(_ string, entry os.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			files++
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if files != 0 {
		t.Fatalf("rollback left %d objects", files)
	}
	if _, err := db.Exec(`DROP TRIGGER fail_event; INSERT INTO calendar_events(id,title,body,user_id,cover_key) VALUES(10,'Hidden key','Body',1,'avatars/1/source')`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.OpenCover(ctx, "10"); !errors.Is(err, ErrNotFound) {
		t.Fatal("exposed arbitrary storage key")
	}
}

func TestEventWithoutStartTimeDoesNotInventMidnight(t *testing.T) {
	s, _, _ := eventService(t)
	ctx := context.Background()
	value, err := s.Save(ctx, 1, "", Draft{Title: "LAN-Party", Body: "Uhrzeit und Ort folgen.", Date: "27.11.2026"}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	event, err := s.Lookup(ctx, value)
	if err != nil || event.EventDate != "2026-11-27" || event.EventLabel() != "27. November 2026" {
		t.Fatalf("date-only event: %+v %v", event, err)
	}
}
