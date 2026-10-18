package exams

import (
	"context"
	"errors"
	"testing"
)

func TestReviewDecisionCanBeReversedWithPublication(t *testing.T) {
	s, db := reviewService(t)
	ctx := context.Background()
	saved, err := s.SaveDraft(ctx, 1, submission(t, db))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, saved.SubmissionID, true); !errors.Is(err, ErrConflict) {
		t.Fatal("reviewed an unsubmitted draft")
	}
	if err := s.Submit(ctx, 1, saved.SubmissionID); err != nil {
		t.Fatal(err)
	}
	for _, approve := range []bool{false, true, true, false, false, true} {
		if err := s.Review(ctx, 3, saved.SubmissionID, approve); !errors.Is(err, ErrForbidden) {
			t.Fatal("ordinary account changed moderation")
		}
		if err := s.Review(ctx, 2, saved.SubmissionID, approve); err != nil {
			t.Fatal(err)
		}
		current, err := s.Submission(ctx, 1, saved.SubmissionID, false)
		if err != nil {
			t.Fatal(err)
		}
		want := Rejected
		if approve {
			want = Approved
		}
		if current.Status != want {
			t.Fatalf("status %s, want %s", current.Status, want)
		}
		object, err := s.Open(ctx, 3, saved.IDs[0])
		if approve {
			if err != nil {
				t.Fatal(err)
			}
			object.Close()
		} else if !errors.Is(err, ErrNotFound) {
			t.Fatal("rejected PDF remains readable")
		}
		var latest int
		if err := db.QueryRow("SELECT is_latest FROM archive WHERE id=?", testInternalID(t, db, "documents", saved.IDs[0])).Scan(&latest); err != nil {
			t.Fatal(err)
		}
		if (latest == 1) != approve {
			t.Fatal("publication flags disagree with decision")
		}
	}
}

func TestReversedRevisionRestoresPredecessorAndProtectsNewerHead(t *testing.T) {
	s, db := reviewService(t)
	ctx := context.Background()
	input := submission(t, db)
	input.RevisionOf = testPublicID(t, db, "documents", "101")
	first, err := s.SaveDraft(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(ctx, 1, first.SubmissionID); err != nil {
		t.Fatal(err)
	}
	for _, approve := range []bool{true, false, true} {
		if err := s.Review(ctx, 2, first.SubmissionID, approve); err != nil {
			t.Fatal(err)
		}
		var head string
		if err := db.QueryRow("SELECT id FROM archive WHERE group_id=100 AND is_latest=1 AND deleted_at IS NULL").Scan(&head); err != nil {
			t.Fatal(err)
		}
		want := "101"
		if approve {
			want = testInternalID(t, db, "documents", first.IDs[0])
		}
		if head != want {
			t.Fatalf("head %s, want %s", head, want)
		}
	}
	input.RevisionOf = first.IDs[0]
	input.Documents = [][]byte{[]byte("%PDF-1.7\nnewer revision\n%%EOF")}
	second, err := s.SaveDraft(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(ctx, 1, second.SubmissionID); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, second.SubmissionID, true); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, first.SubmissionID, false); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, first.SubmissionID, true); !errors.Is(err, ErrConflict) {
		t.Fatal("old approval displaced newer revision")
	}
	var latest int
	if err := db.QueryRow("SELECT is_latest FROM archive WHERE id=?", testInternalID(t, db, "documents", second.IDs[0])).Scan(&latest); err != nil || latest != 1 {
		t.Fatal("reversal of historical review displaced current revision")
	}
	if err := s.Review(ctx, 2, second.SubmissionID, false); err != nil {
		t.Fatal(err)
	}
	var head string
	if err := db.QueryRow("SELECT id FROM archive WHERE group_id=100 AND is_latest=1").Scan(&head); err != nil || head != "101" {
		t.Fatal("rejection restored an unapproved predecessor")
	}
}

func TestRejectionFailureRollsBackPublication(t *testing.T) {
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
	if err := s.Review(ctx, 2, saved.SubmissionID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_reject BEFORE UPDATE OF status ON archive_requests WHEN NEW.status='rejected' BEGIN SELECT RAISE(FAIL,'forced review failure'); END`); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, saved.SubmissionID, false); err == nil {
		t.Fatal("failed rejection reported success")
	}
	current, err := s.Submission(ctx, 1, saved.SubmissionID, false)
	if err != nil || current.Status != Approved {
		t.Fatal("failed rejection changed status")
	}
	var latest int
	if err := db.QueryRow("SELECT is_latest FROM archive WHERE id=?", testInternalID(t, db, "documents", saved.IDs[0])).Scan(&latest); err != nil || latest != 1 {
		t.Fatal("failed rejection unpublished current revision")
	}
	if err := db.QueryRow("SELECT is_latest FROM archive WHERE id=101").Scan(&latest); err != nil || latest != 0 {
		t.Fatal("failed rejection published old revision")
	}
}

func TestMergedApplicabilityCanBeRejectedAndConfirmedAgain(t *testing.T) {
	s, db := reviewService(t)
	ctx := context.Background()
	input := submission(t, db)
	first, err := s.SaveDraft(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(ctx, 1, first.SubmissionID); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, first.SubmissionID, true); err != nil {
		t.Fatal(err)
	}
	input.Assignments = []Assignment{{ModuleID: testPublicID(t, db, "modules", "11"), Version: "2023"}}
	second, err := s.SaveDraft(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(ctx, 1, second.SubmissionID); err != nil {
		t.Fatal(err)
	}
	for _, approve := range []bool{true, false, true, false, true} {
		if err := s.Review(ctx, 2, second.SubmissionID, approve); err != nil {
			t.Fatal(err)
		}
		var head string
		if err := db.QueryRow("SELECT id FROM archive WHERE group_id=? AND is_latest=1", testInternalID(t, db, "documents", first.IDs[0])).Scan(&head); err != nil {
			t.Fatal(err)
		}
		want := testInternalID(t, db, "documents", first.IDs[0])
		if approve {
			want = testInternalID(t, db, "documents", second.IDs[0])
		}
		if head != want {
			t.Fatalf("merged head %s, want %s", head, want)
		}
	}
	var version, cycles int
	if err := db.QueryRow("SELECT edit_version FROM archive WHERE id=?", testInternalID(t, db, "documents", second.IDs[0])).Scan(&version); err != nil || version != 2 {
		t.Fatal("review reversals changed revision numbering")
	}
	if err := db.QueryRow("SELECT count(*) FROM archive_group_aliases WHERE group_id=canonical_group_id").Scan(&cycles); err != nil || cycles != 0 {
		t.Fatal("reapproval created a cyclic alias")
	}
}
