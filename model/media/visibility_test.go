package media

import (
	"context"
	"errors"
	"testing"
)

func TestHomepageSelectionDoesNotHidePublicPhotos(t *testing.T) {
	s, db := mediaService(t)
	ctx := context.Background()
	if _, err := db.Exec(`UPDATE users SET active=1,verified=1,role='editor' WHERE id=1;
UPDATE media SET homepage_visible=0`); err != nil {
		t.Fatal(err)
	}
	photo := reference(t, db, "media", "101")
	albumID := reference(t, db, "albums", "10")
	if slides, err := s.Carousel(ctx); err != nil || len(slides) != 0 {
		t.Fatalf("unselected carousel: %+v %v", slides, err)
	}
	if err := s.SetHomepageVisibility(ctx, 1, photo, true); err != nil {
		t.Fatal(err)
	}
	slides, err := s.Carousel(ctx)
	if err != nil || len(slides) != 1 || slides[0].Cover.ID != photo || slides[0].ID != albumID {
		t.Fatalf("selected slide: %+v %v", slides, err)
	}
	// Multiple photos from the same album must become distinct slides.
	second := reference(t, db, "media", "100")
	if err := s.SetHomepageVisibility(ctx, 1, second, true); err != nil {
		t.Fatal(err)
	}
	if slides, err := s.Carousel(ctx); err != nil || len(slides) != 2 {
		t.Fatalf("two selected photos: %+v %v", slides, err)
	}
	if err := s.SetHomepageVisibility(ctx, 1, photo, false); err != nil {
		t.Fatal(err)
	}
	a, err := s.Album(ctx, albumID)
	if err != nil || a.OriginalCount != 3 {
		t.Fatalf("public album changed: %+v %v", a, err)
	}
	images, err := s.Photos(ctx, a, 0, 10)
	if err != nil || len(images) != 3 {
		t.Fatalf("public photos changed: %+v %v", images, err)
	}
	if _, err := s.ItemAlbum(ctx, photo); err != nil {
		t.Fatalf("unselected viewer disappeared: %v", err)
	}
	if _, err := s.PhotoPosition(ctx, albumID, photo); err != nil {
		t.Fatalf("unselected position disappeared: %v", err)
	}
	// A one-photo album remains public regardless of homepage selection.
	if a, err := s.Album(ctx, reference(t, db, "albums", "20")); err != nil || a.OriginalCount != 1 {
		t.Fatalf("single-photo album missing: %+v %v", a, err)
	}
	if _, err := db.Exec(`INSERT INTO media(id,event_id,user_id,title,access_key,mime_type,nbytes) VALUES(107,10,1,'New','107','image/jpeg',5)`); err != nil {
		t.Fatal(err)
	}
	seedReferences(t, db)
	var selected bool
	if err := db.QueryRow(`SELECT homepage_visible FROM media WHERE id=107`).Scan(&selected); err != nil || selected {
		t.Fatalf("new photo auto-selected: %v %v", selected, err)
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
				t.Fatal("unselected checkbox checked")
			}
		}
	}
	if !found {
		t.Fatal("unselected photo unavailable for selection")
	}
}

func TestHomepageSelectionRequiresActiveVerifiedStaff(t *testing.T) {
	s, db := mediaService(t)
	ctx := context.Background()
	photo := reference(t, db, "media", "101")
	for _, tc := range []struct {
		role             string
		active, verified int
	}{{"user", 1, 1}, {"admin", 0, 1}, {"editor", 1, 0}} {
		if _, err := db.Exec(`UPDATE users SET role=?,active=?,verified=? WHERE id=1`, tc.role, tc.active, tc.verified); err != nil {
			t.Fatal(err)
		}
		if err := s.SetHomepageVisibility(ctx, 1, photo, true); !errors.Is(err, ErrForbidden) {
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
		if err := s.SetHomepageVisibility(ctx, 1, reference(t, db, "media", id), true); !errors.Is(err, ErrNotFound) {
			t.Fatalf("non-public media %s: %v", id, err)
		}
	}
}
