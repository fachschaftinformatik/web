package view

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/model/events"
	"github.com/fachschaftinformatik/web/model/exams"
	"github.com/fachschaftinformatik/web/model/media"
	"github.com/fachschaftinformatik/web/model/members"
	"github.com/fachschaftinformatik/web/model/organization"
	"github.com/fachschaftinformatik/web/model/posts"
	"github.com/fachschaftinformatik/web/model/users"
)

func TestEventPresentationKeepsDateAndPublicationDistinct(t *testing.T) {
	e := Event(events.Event{ID: "42", EventDate: "2026-11-27", SourcePublishedAt: "2025-11-07T12:00:00Z", SourceStatus: "cancelled", CoverKey: "calendar/covers/42.jpg"})
	if e.EventLabel() != "27.11.2026" || e.SourceDateLabel() != "07.11.2025" || e.StatusLabel() != "" {
		t.Fatalf("incorrect event presentation: %s, %s, %s", e.EventLabel(), e.SourceDateLabel(), e.StatusLabel())
	}
	if e.CoverURL() != "/events/42/cover" || e.FullImageURL() != e.CoverURL() || e.AlbumURL() != "" {
		t.Fatal("invalid cover fallback")
	}
	e.AlbumID = sql.NullInt64{Int64: 10, Valid: true}
	e.AlbumCoverID = sql.NullInt64{Int64: 20, Valid: true}
	e.AlbumPublicID = "m7Qp2Vx9aK4"
	e.AlbumCoverPublicID = "8vJd3Rk5nQ0"
	if e.AlbumURL() != "/media/albums/m7Qp2Vx9aK4" || e.FullImageURL() != "/images/8vJd3Rk5nQ0/original" {
		t.Fatal("album linkage lost")
	}
}

func TestEventMarkdownPreservesSourceAndPlainSummary(t *testing.T) {
	e := Event(events.Event{Body: "Am **22.04.2026 ab 16:00 Uhr** geht es los.\n\n**16:30 Uhr:** Nächster Vortrag", BodyFormat: "plain"})
	if !strings.Contains(e.RichBody(), "<strong>22.04.2026 ab 16:00 Uhr</strong>") {
		t.Fatal("event description did not render bold timetable text", e.RichBody())
	}
	if strings.Contains(e.Summary(), "**") || strings.Contains(e.Summary(), "<strong>") || !strings.Contains(e.Summary(), "16:30 Uhr:") {
		t.Fatal("card summary contains formatting markers or lost text", e.Summary())
	}
	if !strings.Contains(e.Body, "**16:30 Uhr:**") {
		t.Fatal("rendering changed the original event source")
	}
}

func TestMediaPresentationPreservesCoverAndPageReturn(t *testing.T) {
	album := media.Album{ID: "20", Cover: media.Image{ID: "20", AlbumCover: true}}
	for i := 0; i < 26; i++ {
		album.Images = append(album.Images, media.Image{ID: "photo"})
	}
	a := NewAlbum(album)
	if a.ReturnURL(25) != "/media/albums/20?page=2#image-photo" {
		t.Fatal("viewer return lost its page or image anchor")
	}
	cover := a.CoverImage()
	if cover.Source() != "/media/albums/20/cover/1200" || !strings.Contains(cover.ThumbnailSources(), "/media/albums/20/cover/800 2x") {
		t.Fatal("album cover uses photograph URLs")
	}
	photo := NewImage(media.Image{ID: "123"})
	if photo.Source() != "/images/123/1200" || photo.Thumbnail() != "/images/123/400" {
		t.Fatal("invalid photograph previews")
	}
}

func TestExamsPresentationFormatsCanonicalRecords(t *testing.T) {
	raw := exams.Exam{ID: "42", Date: "2026-03-27", Version: "PO2017, PO2023", Bytes: 1024}
	e := NewExam(raw)
	if e.Date() != "27.03.2026" || e.VersionLabel() != "2017, 2023" || e.Size() != "1 kB" || e.URL() != "/exams/documents/42" {
		t.Fatal("exams formatting changed")
	}
	if raw.Date != "2026-03-27" || raw.Version != "PO2017, PO2023" {
		t.Fatal("presentation mutated domain data")
	}
}

func TestAuthorPresentationRespectsVisibility(t *testing.T) {
	p := NewPost(posts.Post{ID: "42", AuthorID: 7, AuthorPublicID: "m7Qp2Vx9aK4", AuthorPublic: true})
	if p.URL() != "/posts/42" || p.AuthorURL() != "" {
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
		member members.Member
		want   string
	}{
		{members.Member{}, "Ohne Benutzerkonto"},
		{members.Member{UserID: 7}, "Eingeladen"},
		{members.Member{UserID: 7, Confirmed: true}, "Bestätigt"},
	} {
		if got := (Member{Member: tc.member}).Status(); got != tc.want {
			t.Fatalf("status=%q want %q", got, tc.want)
		}
	}
}

func TestOrganizationalRoleOptionsSeparateIdentityFromLabels(t *testing.T) {
	const publicID = "m7Qp2Vx9aK4"
	fields := OrganizationSettingsFields("contact", organization.Details{ChairRole: "legacy-role"}, []members.MemberRole{{ID: "legacy-role", PublicID: publicID, Group: "board", Title: "Vorsitz"}})
	for _, field := range fields {
		if field.ID != "ChairRole" {
			continue
		}
		for _, option := range field.Options {
			if option.Label == "Vorsitz" && option.Selected && option.Value == publicID {
				return
			}
		}
	}
	t.Fatal("role form must submit the public identity while selecting the stored internal role")
}
