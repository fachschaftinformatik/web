package view

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/model/archive"
	"github.com/fachschaftinformatik/web/internal/model/events"
	"github.com/fachschaftinformatik/web/internal/model/forum"
	"github.com/fachschaftinformatik/web/internal/model/gallery"
	"github.com/fachschaftinformatik/web/internal/model/people"
	"github.com/fachschaftinformatik/web/internal/model/users"
)

func TestEventPresentationKeepsDateAndPublicationDistinct(t *testing.T) {
	e := Event(events.Event{ID: "42", EventDate: "2026-11-27", SourcePublishedAt: "2025-11-07T12:00:00Z", SourceStatus: "cancelled", CoverKey: "calendar/covers/42.jpg"})
	if e.EventLabel() != "27.11.2026" || e.SourceDateLabel() != "07.11.2025" || e.StatusLabel() != "Abgesagt" {
		t.Fatalf("incorrect event presentation: %s, %s, %s", e.EventLabel(), e.SourceDateLabel(), e.StatusLabel())
	}
	if e.CoverURL() != "/community/veranstaltungen/42/cover" || e.FullImageURL() != e.CoverURL() || e.GalleryURL() != "" {
		t.Fatal("invalid cover fallback")
	}
	e.GalleryAlbumID = sql.NullInt64{Int64: 10, Valid: true}
	e.GalleryCoverID = sql.NullInt64{Int64: 20, Valid: true}
	if e.GalleryURL() != "/community/events/10" || e.FullImageURL() != "/media/20/original" {
		t.Fatal("gallery linkage lost")
	}
}

func TestGalleryPresentationPreservesCoverAndPageReturn(t *testing.T) {
	album := gallery.Album{ID: "20", Cover: gallery.Image{ID: "20", AlbumCover: true}}
	for i := 0; i < 26; i++ {
		album.Images = append(album.Images, gallery.Image{ID: "photo"})
	}
	a := NewAlbum(album)
	if a.ReturnURL(25) != "/community/events/20?page=2#image-photo" {
		t.Fatal("viewer return lost its page or image anchor")
	}
	cover := a.CoverImage()
	if cover.Source() != "/media/events/20/cover/1200" || !strings.Contains(cover.ThumbnailSources(), "/media/events/20/cover/800 2x") {
		t.Fatal("album cover uses photograph URLs")
	}
	photo := NewImage(gallery.Image{ID: "123"})
	if photo.Source() != "/media/123/1200" || photo.Thumbnail() != "/media/123/400" {
		t.Fatal("invalid photograph previews")
	}
}

func TestArchivePresentationFormatsCanonicalRecords(t *testing.T) {
	raw := archive.Exam{ID: "42", Date: "2026-03-27", Version: "PO2017, PO2023", Bytes: 1024}
	e := NewExam(raw)
	if e.Date() != "27.03.2026" || e.VersionLabel() != "2017, 2023" || e.Size() != "1 kB" || e.URL() != "/archive/files/42" {
		t.Fatal("archive formatting changed")
	}
	if raw.Date != "2026-03-27" || raw.Version != "PO2017, PO2023" {
		t.Fatal("presentation mutated domain data")
	}
}

func TestAuthorPresentationRespectsVisibility(t *testing.T) {
	p := NewPost(forum.Post{ID: "42", AuthorID: 7, AuthorPublic: true})
	if p.URL() != "/d/42" || p.AuthorURL() != "/users/7" {
		t.Fatal("public author not linked")
	}
	p.AuthorPublic = false
	if p.AuthorURL() != "" {
		t.Fatal("private author linked")
	}
}

func TestAccountAndMemberLabels(t *testing.T) {
	u := User{User: users.User{Role: "editor", Active: true}}
	if u.Permission() != "Redaktion" || u.Status() != "Aktiv" {
		t.Fatal("account labels changed")
	}
	for _, tc := range []struct {
		person people.Person
		want   string
	}{
		{people.Person{}, "Ohne Benutzerkonto"},
		{people.Person{UserID: 7}, "Eingeladen"},
		{people.Person{UserID: 7, Confirmed: true}, "Bestätigt"},
	} {
		if got := (Person{Person: tc.person}).Status(); got != tc.want {
			t.Fatalf("status=%q want %q", got, tc.want)
		}
	}
}
