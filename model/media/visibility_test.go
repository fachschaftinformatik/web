package media

import (
	"context"
	"errors"
	"testing"
)

func TestGallerySelection(t *testing.T) {
	s, db := mediaService(t)
	ctx := context.Background()
	if _, err := db.Exec(`UPDATE users SET active=1,verified=1,role='editor' WHERE id=1;
INSERT INTO media(id,event_id,user_id,title,access_key,mime_type,nbytes,created_at,gallery_visible) VALUES (107,10,1,'Oldest','107','image/jpeg',5,'2024-01-01T00:00:00Z',1)`); err != nil {
		t.Fatal(err)
	}
	photo := reference(t, db, "media", "101")
	albumID := reference(t, db, "albums", "10")
	if err := s.SetGalleryVisibility(ctx, 1, photo, false); err != nil {
		t.Fatal(err)
	}
	a, err := s.Album(ctx, albumID)
	if err != nil || a.OriginalCount != 3 || a.Cover.ID == photo {
		t.Fatalf("album count/cover: %+v %v", a, err)
	}
	images, err := s.Photos(ctx, a, 0, 10)
	if err != nil || len(images) != 3 {
		t.Fatalf("photos: %+v %v", images, err)
	}
	for _, image := range images {
		if image.ID == photo {
			t.Fatal("excluded photo still in album")
		}
	}
	if _, err := s.ItemAlbum(ctx, photo); !errors.Is(err, ErrNotFound) {
		t.Fatalf("excluded viewer lookup: %v", err)
	}
	if _, err := s.PhotoPosition(ctx, albumID, photo); !errors.Is(err, ErrNotFound) {
		t.Fatalf("excluded viewer position: %v", err)
	}
	position, err := s.PhotoPosition(ctx, albumID, reference(t, db, "media", "100"))
	if err != nil || position != 0 {
		t.Fatalf("visible position: %d %v", position, err)
	}
	items, _, err := s.GalleryPage(ctx, 1, 1, 12)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range items {
		if item.ID == photo {
			found = true
			if item.Visible {
				t.Fatal("excluded checkbox still selected")
			}
		}
	}
	if !found {
		t.Fatal("excluded photo cannot be restored")
	}
	if err := s.SetGalleryVisibility(ctx, 1, reference(t, db, "media", "100"), false); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Album(ctx, albumID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("album with fewer than three selected photos: %v", err)
	}
	if err := s.SetGalleryVisibility(ctx, 1, photo, true); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Album(ctx, albumID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM media WHERE event_id=10`).Scan(&count); err != nil || count != 4 {
		t.Fatalf("selection changed stored media: %d %v", count, err)
	}
}

func TestGallerySelectionRequiresActiveVerifiedStaff(t *testing.T) {
	s, db := mediaService(t)
	ctx := context.Background()
	photo := reference(t, db, "media", "101")
	for _, tc := range []struct {
		role             string
		active, verified int
	}{
		{"user", 1, 1}, {"admin", 0, 1}, {"editor", 1, 0},
	} {
		if _, err := db.Exec(`UPDATE users SET role=?,active=?,verified=? WHERE id=1`, tc.role, tc.active, tc.verified); err != nil {
			t.Fatal(err)
		}
		if err := s.SetGalleryVisibility(ctx, 1, photo, false); !errors.Is(err, ErrForbidden) {
			t.Fatalf("unauthorized mutation: %v", err)
		}
		if _, _, err := s.GalleryPage(ctx, 1, 1, 12); !errors.Is(err, ErrForbidden) {
			t.Fatalf("unauthorized management: %v", err)
		}
	}
	if _, err := db.Exec(`UPDATE users SET role='admin',active=1,verified=1 WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"103", "104", "105", "106"} {
		if err := s.SetGalleryVisibility(ctx, 1, reference(t, db, "media", id), false); !errors.Is(err, ErrNotFound) {
			t.Fatalf("non-gallery media %s: %v", id, err)
		}
	}
}
