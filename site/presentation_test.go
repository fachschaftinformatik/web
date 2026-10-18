package site

import (
	"testing"

	"github.com/fachschaftinformatik/web/model/exams"
	"github.com/fachschaftinformatik/web/model/organization"
	"github.com/fachschaftinformatik/web/model/posts"
)

func TestPresentationSnapshotsDoNotMutateDomainRecords(t *testing.T) {
	post := posts.Post{ID: "post-public", Tags: []string{"original"}}
	presentedPost := postView(post)
	presentedPost.Tags[0] = "changed"
	if post.Tags[0] != "original" {
		t.Fatal("view tags alias the domain record")
	}

	details := organization.Details{SocialLinks: []organization.Link{{Title: "Original", URL: "https://example.org"}}}
	presentedOrganization := organizationView(details)
	presentedOrganization.SocialLinks[0].Title = "Changed"
	if details.SocialLinks[0].Title != "Original" {
		t.Fatal("view links alias the organization settings")
	}

	submitted := "2026-06-30T15:00:00Z"
	submission := exams.Submission{
		ID: "submission-public", SubmittedAt: &submitted,
		Documents: []exams.Exam{{ID: "document-public", Assignments: []exams.Assignment{{Title: "Original"}}}},
	}
	presentedSubmission := submissionView(submission)
	*presentedSubmission.SubmittedAt = "2027-06-30T15:00:00Z"
	presentedSubmission.Documents[0].Assignments[0].Title = "Changed"
	if submitted != "2026-06-30T15:00:00Z" || submission.Documents[0].Assignments[0].Title != "Original" {
		t.Fatal("view submission aliases domain state")
	}
	if presentedSubmission.ID != "submission-public" || presentedSubmission.Documents[0].ID != "document-public" {
		t.Fatal("presentation lost public identities")
	}
}
