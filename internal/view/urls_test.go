package view

import (
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/model/exams"
	"github.com/fachschaftinformatik/web/internal/model/media"
	"github.com/fachschaftinformatik/web/internal/model/posts"
)

func TestOrganizationalDocumentURLsUsePublicIdentity(t *testing.T) {
	const publicID = "m7Qp2Vx9aK4"
	if got := (Document{Stored: true, File: "publication-42", PublicID: publicID}).URL(); got != "/documents/"+publicID {
		t.Fatal(got)
	}
	if got := (Document{File: "synthetic-minutes.pdf", PublicID: publicID}).URL(); got != "/documents/"+publicID {
		t.Fatal(got)
	}
}

func TestCanonicalJourneyFormsAndNavigation(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	submission := NewSubmission(exams.Submission{ID: "42", Status: exams.Pending})
	thread := NewThread(posts.Thread{Post: posts.Post{ID: "42", Title: "Synthetic post"}, Replies: []posts.Comment{{ID: "7"}}})
	for _, tc := range []struct {
		name string
		page Page
		want []string
	}{
		{"exams", Page{Active: "exams"}, []string{`action="/exams"`, `name="search"`, `href="/exams/submissions/create"`, `href="/exams?program=all"`, `name="program" value="all"`, `data-active="false"`}},
		{"exams", Page{Active: "exams", Exams: ExamsPage{Degrees: []FilterOption{{Name: "bachelor", Selected: true}}}}, []string{`data-active="true"`}},
		{"module", Page{Active: "exams", UserID: 1, Exams: ExamsPage{Module: &Module{Module: exams.Module{ID: "42"}, Exams: []Exam{{Exam: exams.Exam{ID: "23", OwnerID: 1}}}}}}, []string{`href="/exams/submissions/create?revision=23"`}},
		{"posts", Page{Active: "posts", Community: true}, []string{`action="/posts"`, `href="/posts/create"`, `href="/news"`, `href="/events"`, `href="/media"`}},
		{"submission", Page{Active: "admin", Admin: &AdminPage{Section: "submissions"}, Submission: &SubmissionPage{Entry: submission, Reviewing: true, EditURL: "/admin/submissions/42/edit"}}, []string{`action="/admin/submissions/42/review"`, `value="approve"`, `value="reject"`, `Freigeben`}},
		{"submissions", Page{Active: "exams", SubmissionList: SubmissionListPage{Submissions: []Submission{submission}}}, []string{`href="/exams/submissions/42"`}},
		{"admin", Page{Active: "admin", Admin: &AdminPage{View: "submissions", Submissions: []Submission{submission}}}, []string{`href="/admin/submissions/42/review"`}},
		{"thread", Page{Active: "posts", SignedIn: true, Posts: PostsPage{Thread: &thread, CanWrite: true, CanEdit: true}}, []string{`action="/posts/42/comments"`, `action="/posts/42/vote"`, `action="/posts/42/comments/7/vote"`, `href="/posts/42/edit"`, `id="comment-7"`}},
		{"token", Page{Token: TokenPage{Kind: "reset", Value: "synthetic-token"}}, []string{`action="/password/reset"`, `name="token" value="synthetic-token"`}},
		{"token", Page{Token: TokenPage{Kind: "confirm", Value: "synthetic-token"}}, []string{`action="/email/confirm"`, `E-Mail-Adresse bestätigen`}},
		{"invite", Page{}, []string{`action="/invitation/accept"`}},
		{"forgot", Page{}, []string{`action="/password/forgot"`}},
		{"upload", Page{Upload: UploadPage{Action: "/admin/submissions/42/edit", SubmissionID: "42", ReviewEditing: true, EditToken: "synthetic-concurrency-token"}}, []string{`method="post"`, `name="edit_token" value="synthetic-concurrency-token"`, `href="/admin/submissions/42/review"`}},
		{"upload", Page{Upload: UploadPage{Action: "/exams/submissions/create", ChooseScope: true, RevisionID: "23"}}, []string{`method="get"`, `name="revision" value="23"`}},
		{"media", Page{Active: "media", Community: true, Media: MediaPage{Albums: []Album{NewAlbum(media.Album{ID: "42", Title: "Synthetic album", Cover: media.Image{ID: "42", AlbumCover: true}})}}}, []string{`href="/media/albums/42"`, `/media/albums/42/cover/400`, `>Fotos`}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			if err := renderer.Render(&out, tc.name, false, tc.page); err != nil {
				t.Fatal(err)
			}
			html := out.String()
			for _, want := range tc.want {
				if !strings.Contains(html, want) {
					t.Errorf("missing %q", want)
				}
			}
			for _, old := range []string{`href="/archive`, `action="/archive`, `href="/verwaltung`, `action="/verwaltung`, `href="/discussions`, `href="/community/events`, `href="/fachschaft`, `?choose_scope=`, `?edit_token=`, `name="q"`, `name="applied"`, `name="scope"`} {
				if strings.Contains(html, old) {
					t.Errorf("retired navigation %q", old)
				}
			}
		})
	}
}
