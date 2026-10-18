package exams

import (
	"context"
	"errors"
	"testing"
)

func TestReviewVersionRejectsEditedContent(t *testing.T) {
	for _, edit := range []string{"PDF", "metadata"} {
		t.Run(edit, func(t *testing.T) {
			s, db := reviewService(t)
			ctx := context.Background()
			input := submission(t, db)
			input.RevisionOf = testPublicID(t, db, "documents", "101")
			saved, err := s.SaveDraft(ctx, 1, input)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Submit(ctx, 1, saved.SubmissionID); err != nil {
				t.Fatal(err)
			}
			viewed, err := s.Submission(ctx, 2, saved.SubmissionID, true)
			if err != nil {
				t.Fatal(err)
			}
			input.SubmissionID = saved.SubmissionID
			input.Documents = nil
			if edit == "PDF" {
				input.Documents = [][]byte{[]byte("%PDF-1.7\nreplacement\n%%EOF")}
			} else {
				input.Comment = "Corrected metadata"
			}
			if _, err := s.EditSubmission(ctx, 2, input, viewed.EditToken); err != nil {
				t.Fatal(err)
			}
			updated, err := s.Submission(ctx, 2, saved.SubmissionID, true)
			if err != nil {
				t.Fatal(err)
			}
			for _, token := range []string{"", viewed.EditToken} {
				for _, approve := range []bool{true, false} {
					if err := s.ReviewVersion(ctx, 2, saved.SubmissionID, approve, token); !errors.Is(err, ErrConflict) {
						t.Fatalf("stale review: %v", err)
					}
				}
			}
			current, err := s.Submission(ctx, 2, saved.SubmissionID, true)
			if err != nil || current.EditToken != updated.EditToken || current.Status != Pending || current.ReviewedAt != nil {
				t.Fatal("stale review changed content or moderation state")
			}
			var latest int
			if err := db.QueryRow("SELECT is_latest FROM archive WHERE id=101").Scan(&latest); err != nil || latest != 1 {
				t.Fatal("stale review displaced the published revision")
			}
			if _, err := s.Open(ctx, 3, saved.IDs[0]); !errors.Is(err, ErrNotFound) {
				t.Fatal("stale review published replacement content")
			}
			if err := s.ReviewVersion(ctx, 2, saved.SubmissionID, true, updated.EditToken); err != nil {
				t.Fatal("freshly inspected content could not be approved:", err)
			}
		})
	}
}

func TestReviewVersionRejectsInterveningDecision(t *testing.T) {
	for _, approve := range []bool{true, false} {
		t.Run(map[bool]string{true: "approved", false: "rejected"}[approve], func(t *testing.T) {
			s, db := reviewService(t)
			ctx := context.Background()
			saved, err := s.SaveDraft(ctx, 1, submission(t, db))
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Submit(ctx, 1, saved.SubmissionID); err != nil {
				t.Fatal(err)
			}
			viewed, err := s.Submission(ctx, 2, saved.SubmissionID, true)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.ReviewVersion(ctx, 2, saved.SubmissionID, approve, viewed.EditToken); err != nil {
				t.Fatal(err)
			}
			decided, err := s.Submission(ctx, 2, saved.SubmissionID, true)
			if err != nil {
				t.Fatal(err)
			}
			// Both a stale reversal and a replay must validate the viewed version.
			for _, decision := range []bool{!approve, approve} {
				if err := s.ReviewVersion(ctx, 2, saved.SubmissionID, decision, viewed.EditToken); !errors.Is(err, ErrConflict) {
					t.Fatalf("stale decision: %v", err)
				}
			}
			current, err := s.Submission(ctx, 2, saved.SubmissionID, true)
			if err != nil || current.EditToken != decided.EditToken {
				t.Fatal("stale decision changed moderation or publication state")
			}
			var latest bool
			if err := db.QueryRow("SELECT is_latest FROM archive WHERE id=?", testInternalID(t, db, "documents", saved.IDs[0])).Scan(&latest); err != nil || latest != approve {
				t.Fatal("stale decision changed public availability")
			}
			if err := s.ReviewVersion(ctx, 2, saved.SubmissionID, !approve, decided.EditToken); err != nil {
				t.Fatal("freshly inspected decision could not be reversed:", err)
			}
		})
	}
}

func TestReviewVersionConcurrentDecisions(t *testing.T) {
	s, db := reviewService(t)
	ctx := context.Background()
	saved, err := s.SaveDraft(ctx, 1, submission(t, db))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(ctx, 1, saved.SubmissionID); err != nil {
		t.Fatal(err)
	}
	viewed, err := s.Submission(ctx, 2, saved.SubmissionID, true)
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		approve bool
		err     error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	for _, approve := range []bool{true, false} {
		go func(approve bool) {
			<-start
			results <- result{approve, s.ReviewVersion(ctx, 2, saved.SubmissionID, approve, viewed.EditToken)}
		}(approve)
	}
	close(start)
	first, second := <-results, <-results
	if first.err != nil {
		first, second = second, first
	}
	if first.err != nil || !errors.Is(second.err, ErrConflict) {
		t.Fatalf("expected one successful decision and one conflict, got %v and %v", first.err, second.err)
	}
	current, err := s.Submission(ctx, 2, saved.SubmissionID, true)
	if err != nil {
		t.Fatal(err)
	}
	want := Rejected
	if first.approve {
		want = Approved
	}
	if current.Status != want {
		t.Fatal("losing concurrent decision overwrote the winning decision")
	}
	var latest bool
	if err := db.QueryRow("SELECT is_latest FROM archive WHERE id=?", testInternalID(t, db, "documents", saved.IDs[0])).Scan(&latest); err != nil || latest != first.approve {
		t.Fatal("losing concurrent decision changed publication state")
	}
}
