package files

import (
	"context"
	"errors"
	"testing"
)

func TestReviewDecisionCanBeReversedWithPublication(t *testing.T) {
	s, db := reviewService(t)
	ctx := context.Background()
	saved, err := s.SaveDraft(ctx, 1, submission())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, saved.RequestID, true); !errors.Is(err, ErrConflict) {
		t.Fatal("reviewed an unsubmitted draft")
	}
	if err := s.SendRequest(ctx, 1, saved.RequestID); err != nil {
		t.Fatal(err)
	}
	for _, approve := range []bool{false, true, true, false, false, true} {
		if err := s.Review(ctx, 3, saved.RequestID, approve); !errors.Is(err, ErrForbidden) {
			t.Fatal("ordinary account changed moderation")
		}
		if err := s.Review(ctx, 2, saved.RequestID, approve); err != nil {
			t.Fatal(err)
		}
		request, err := s.Request(ctx, 1, saved.RequestID, false)
		if err != nil {
			t.Fatal(err)
		}
		want := Rejected
		if approve {
			want = Approved
		}
		if request.Status != want {
			t.Fatalf("status %s, want %s", request.Status, want)
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
		if err := db.QueryRow("SELECT is_latest FROM archive WHERE id=?", saved.IDs[0]).Scan(&latest); err != nil {
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
	input := submission()
	input.RevisionOf = "101"
	first, err := s.SaveDraft(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SendRequest(ctx, 1, first.RequestID); err != nil {
		t.Fatal(err)
	}
	for _, approve := range []bool{true, false, true} {
		if err := s.Review(ctx, 2, first.RequestID, approve); err != nil {
			t.Fatal(err)
		}
		var head string
		if err := db.QueryRow("SELECT id FROM archive WHERE group_id=100 AND is_latest=1 AND deleted_at IS NULL").Scan(&head); err != nil {
			t.Fatal(err)
		}
		want := "101"
		if approve {
			want = first.IDs[0]
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
	if err := s.SendRequest(ctx, 1, second.RequestID); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, second.RequestID, true); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, first.RequestID, false); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, first.RequestID, true); !errors.Is(err, ErrConflict) {
		t.Fatal("old approval displaced newer revision")
	}
	var latest int
	if err := db.QueryRow("SELECT is_latest FROM archive WHERE id=?", second.IDs[0]).Scan(&latest); err != nil || latest != 1 {
		t.Fatal("reversal of historical review displaced current revision")
	}
	if err := s.Review(ctx, 2, second.RequestID, false); err != nil {
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
	input := submission()
	input.RevisionOf = "101"
	saved, err := s.SaveDraft(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SendRequest(ctx, 1, saved.RequestID); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, saved.RequestID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_reject BEFORE UPDATE OF status ON archive_requests WHEN NEW.status='rejected' BEGIN SELECT RAISE(FAIL,'forced review failure'); END`); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, saved.RequestID, false); err == nil {
		t.Fatal("failed rejection reported success")
	}
	request, err := s.Request(ctx, 1, saved.RequestID, false)
	if err != nil || request.Status != Approved {
		t.Fatal("failed rejection changed status")
	}
	var latest int
	if err := db.QueryRow("SELECT is_latest FROM archive WHERE id=?", saved.IDs[0]).Scan(&latest); err != nil || latest != 1 {
		t.Fatal("failed rejection unpublished current revision")
	}
	if err := db.QueryRow("SELECT is_latest FROM archive WHERE id=101").Scan(&latest); err != nil || latest != 0 {
		t.Fatal("failed rejection published old revision")
	}
}

func TestMergedApplicabilityCanBeRejectedAndConfirmedAgain(t *testing.T) {
	s, db := reviewService(t)
	ctx := context.Background()
	input := submission()
	first, err := s.SaveDraft(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SendRequest(ctx, 1, first.RequestID); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, first.RequestID, true); err != nil {
		t.Fatal(err)
	}
	input.Assignments = []Assignment{{ModuleID: "11", Version: "2023"}}
	second, err := s.SaveDraft(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SendRequest(ctx, 1, second.RequestID); err != nil {
		t.Fatal(err)
	}
	for _, approve := range []bool{true, false, true, false, true} {
		if err := s.Review(ctx, 2, second.RequestID, approve); err != nil {
			t.Fatal(err)
		}
		var head string
		if err := db.QueryRow("SELECT id FROM archive WHERE group_id=? AND is_latest=1", first.IDs[0]).Scan(&head); err != nil {
			t.Fatal(err)
		}
		want := first.IDs[0]
		if approve {
			want = second.IDs[0]
		}
		if head != want {
			t.Fatalf("merged head %s, want %s", head, want)
		}
	}
	var version, cycles int
	if err := db.QueryRow("SELECT edit_version FROM archive WHERE id=?", second.IDs[0]).Scan(&version); err != nil || version != 2 {
		t.Fatal("review reversals changed revision numbering")
	}
	if err := db.QueryRow("SELECT count(*) FROM archive_group_aliases WHERE group_id=canonical_group_id").Scan(&cycles); err != nil || cycles != 0 {
		t.Fatal("reapproval created a cyclic alias")
	}
}
