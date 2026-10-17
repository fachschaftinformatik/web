package media

import (
	"context"
	"errors"
	"testing"
)

func TestAlbumVisibilityCountsOnlyPublicPhotosBeforePagination(t *testing.T) {
	s, db := galleryService(t)
	ctx := context.Background()
	// The fixture has one three-photo album, a single photo plus deleted/unsafe
	// images, an empty album and a deleted album.
	albums, total, err := s.AlbumPage(ctx, 1, 1)
	if err != nil || total != 1 || len(albums) != 1 || albums[0].ID != "10" || albums[0].OriginalCount != 3 || len(albums[0].Images) != 0 {
		t.Fatalf("summaries: %+v %d %v", albums, total, err)
	}
	for _, id := range []string{"20", "30", "40"} {
		if _, err := s.Album(ctx, id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("album %s: %v", id, err)
		}
	}
	for _, id := range []int{107, 108} {
		if _, err := db.Exec(`INSERT INTO media(id,event_id,user_id,access_key,mime_type,nbytes) VALUES(?,20,1,?,'image/jpeg',5)`, id, id); err != nil {
			t.Fatal(err)
		}
		albums, total, err = s.AlbumPage(ctx, 1, 1)
		want := 1
		if id == 108 {
			want = 2
		}
		if err != nil || total != want || len(albums) != 1 {
			t.Fatalf("%d images: %+v %d %v", id, albums, total, err)
		}
	}
	a, err := s.Album(ctx, "10")
	if err != nil {
		t.Fatal(err)
	}
	photos, err := s.Photos(ctx, a, 1, 1)
	if err != nil || len(photos) != 1 || photos[0].ID != "100" {
		t.Fatalf("photo page: %+v %v", photos, err)
	}
	position, err := s.PhotoPosition(ctx, a.ID, "100")
	if err != nil || position != 1 {
		t.Fatalf("position %d %v", position, err)
	}
	if _, err := s.PhotoPosition(ctx, a.ID, "102"); !errors.Is(err, ErrNotFound) {
		t.Fatal("photo from another album resolved")
	}
	if _, err := db.Exec(`UPDATE media SET deleted_at='deleted' WHERE id=108`); err != nil {
		t.Fatal(err)
	}
	_, total, err = s.AlbumPage(ctx, 1, 12)
	if err != nil || total != 1 {
		t.Fatalf("deleted photo counted: %d %v", total, err)
	}
}
