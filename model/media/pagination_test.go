package media

import (
	"context"
	"errors"
	"testing"
)

func TestAlbumVisibilityCountsOnlyPublicPhotosBeforePagination(t *testing.T) {
	s, db := mediaService(t)
	ctx := context.Background()
	// The fixture has one three-photo album, a single photo plus deleted/unsafe
	// images, an empty album and a deleted album.
	albums, total, err := s.AlbumPage(ctx, 1, 1)
	if err != nil || total != 1 || len(albums) != 1 || albums[0].ID != reference(t, db, "albums", "10") || albums[0].OriginalCount != 3 || len(albums[0].Images) != 0 {
		t.Fatalf("summaries: %+v %d %v", albums, total, err)
	}
	for _, id := range []string{"20", "30", "40"} {
		id = reference(t, db, "albums", id)
		if _, err := s.Album(ctx, id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("album %s: %v", id, err)
		}
	}
	for _, id := range []int{107, 108} {
		if _, err := db.Exec(`INSERT INTO media(id,event_id,user_id,access_key,mime_type,nbytes,gallery_visible) VALUES(?,20,1,?,'image/jpeg',5,1)`, id, id); err != nil {
			t.Fatal(err)
		}
		seedReferences(t, db)
		albums, total, err = s.AlbumPage(ctx, 1, 1)
		want := 1
		if id == 108 {
			want = 2
		}
		if err != nil || total != want || len(albums) != 1 {
			t.Fatalf("%d images: %+v %d %v", id, albums, total, err)
		}
	}
	a, err := s.Album(ctx, reference(t, db, "albums", "10"))
	if err != nil {
		t.Fatal(err)
	}
	photos, err := s.Photos(ctx, a, 1, 1)
	if err != nil || len(photos) != 1 || photos[0].ID != reference(t, db, "media", "100") {
		t.Fatalf("photo page: %+v %v", photos, err)
	}
	position, err := s.PhotoPosition(ctx, a.ID, reference(t, db, "media", "100"))
	if err != nil || position != 1 {
		t.Fatalf("position %d %v", position, err)
	}
	if _, err := s.PhotoPosition(ctx, a.ID, reference(t, db, "media", "102")); !errors.Is(err, ErrNotFound) {
		t.Fatal("photo from another album resolved")
	}
	if _, err := db.Exec(`UPDATE media SET deleted_at='2026-01-01T00:00:00Z' WHERE id=108`); err != nil {
		t.Fatal(err)
	}
	_, total, err = s.AlbumPage(ctx, 1, 12)
	if err != nil || total != 1 {
		t.Fatalf("deleted photo counted: %d %v", total, err)
	}
}

func TestItemAlbumResolvesIdentityAndCurrentPublicVisibility(t *testing.T) {
	s, db := mediaService(t)
	ctx := context.Background()
	for _, item := range []string{"99", "100", "101"} {
		if album, err := s.ItemAlbum(ctx, reference(t, db, "media", item)); err != nil || album != reference(t, db, "albums", "10") {
			t.Fatalf("item %s: album %q, %v", item, album, err)
		}
	}
	// Item 102 belongs to a hidden one-photo album. Deleted items, deleted
	// albums, unsafe formats and noncanonical identities must also stay hidden.
	for _, item := range []string{"100", "102", "103", "104", "105", "106", "999", "0100", "+100", "0", "-1", "", "100/400", "9223372036854775808"} {
		if item == "102" || item == "103" || item == "104" || item == "105" || item == "106" {
			item = reference(t, db, "media", item)
		}
		if album, err := s.ItemAlbum(ctx, item); !errors.Is(err, ErrNotFound) || album != "" {
			t.Errorf("invisible item %q: album %q, %v", item, album, err)
		}
	}
	for _, item := range []int{107, 108} {
		if _, err := db.Exec(`INSERT INTO media(id,event_id,user_id,access_key,mime_type,nbytes,gallery_visible) VALUES(?,20,1,?,'image/jpeg',5,1)`, item, item); err != nil {
			t.Fatal(err)
		}
	}
	if album, err := s.ItemAlbum(ctx, reference(t, db, "media", "102")); err != nil || album != reference(t, db, "albums", "20") {
		t.Fatalf("newly visible item: album %q, %v", album, err)
	}
	// The owner comes from the item itself. Moving it never retains a stale
	// caller-selected album, and dropping below three photos hides its old album.
	if _, err := db.Exec(`UPDATE media SET event_id=20 WHERE id=100`); err != nil {
		t.Fatal(err)
	}
	if album, err := s.ItemAlbum(ctx, reference(t, db, "media", "100")); err != nil || album != reference(t, db, "albums", "20") {
		t.Fatalf("moved item: album %q, %v", album, err)
	}
	if _, err := s.ItemAlbum(ctx, reference(t, db, "media", "99")); !errors.Is(err, ErrNotFound) {
		t.Fatalf("item in newly hidden album: %v", err)
	}
	if _, err := db.Exec(`UPDATE events SET deleted_at='2026-10-18T00:00:00Z' WHERE id=20`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ItemAlbum(ctx, reference(t, db, "media", "100")); !errors.Is(err, ErrNotFound) {
		t.Fatalf("item in newly deleted album: %v", err)
	}
}
