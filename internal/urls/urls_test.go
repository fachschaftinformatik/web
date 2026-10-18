package urls

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildersMatchDeclaredPatterns(t *testing.T) {
	const id = "AbCdEfGhIjA"
	for _, tc := range []struct{ pattern, built, capture string }{
		{PostPattern, Post(id), "post"},
		{EditPostPattern, EditPost(id), "post"},
		{EventPattern, Event(id), "event"},
		{EditEventPattern, EditEvent(id), "event"},
		{MediaItemPattern, MediaItem(id), "image"},
		{AlbumPattern, Album(id), "album"},
		{AlbumCoverPattern, AlbumCover(id, "400"), "id"},
		{ImagePattern, Image(id, "original"), "id"},
		{ExamModulePattern, ExamModule(id), "module"},
		{ExamDocumentPattern, ExamDocument(id), "id"},
		{SubmissionPattern, Submission(id), "submission"},
		{EditSubmissionPattern, EditSubmission(id), "submission"},
		{CheckSubmissionPattern, CheckSubmission(id), "submission"},
		{ReviewSubmissionPattern, ReviewSubmission(id), "submission"},
		{DocumentPattern, Document(id), "id"},
		{UserPattern, User(id), "id"},
		{AdminUserPattern, AdminUser(id), "id"},
		{AdminMemberPattern, AdminMember(id), "id"},
		{AdminMeetingPattern, AdminMeeting(id), "id"},
	} {
		t.Run(tc.pattern, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("GET "+tc.pattern, func(w http.ResponseWriter, r *http.Request) {
				if got := r.PathValue(tc.capture); got != id {
					t.Errorf("capture=%q", got)
				}
				w.WriteHeader(http.StatusNoContent)
			})
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("GET", tc.built, nil))
			if w.Code != http.StatusNoContent {
				t.Fatalf("%s: status=%d", tc.built, w.Code)
			}
		})
	}
}
