package files

import (
	"context"
	"errors"
	"io"
	"testing"
)

func TestReviewerEditsPreserveOwnershipStatusAndDownloadID(t *testing.T) {
	for _, status := range []Status{Pending, Approved, Rejected} {
		t.Run(string(status), func(t *testing.T) {
			s, db := reviewService(t)
			ctx := context.Background()
			input := submission()
			saved, err := s.SaveDraft(ctx, 1, input)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.SendRequest(ctx, 1, saved.RequestID); err != nil {
				t.Fatal(err)
			}
			if status != Pending {
				if err := s.Review(ctx, 2, saved.RequestID, status == Approved); err != nil {
					t.Fatal(err)
				}
			}
			request, err := s.Request(ctx, 2, saved.RequestID, true)
			if err != nil {
				t.Fatal(err)
			}
			input.RequestID = saved.RequestID
			input.Comment = "Redaktionelle Korrektur"
			input.Date = "2026-03-27"
			input.Assignments = []Assignment{{ModuleID: "11", Version: "2023"}}
			input.Documents = [][]byte{[]byte("%PDF-1.7\ncorrected PDF\n%%EOF")}
			for _, user := range []int64{1, 3} {
				if _, err := s.EditRequest(ctx, user, input, request.EditToken); !errors.Is(err, ErrForbidden) {
					t.Fatal("non-reviewer edited a submission")
				}
			}
			result, err := s.EditRequest(ctx, 2, input, request.EditToken)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.IDs) != 1 || result.IDs[0] != saved.IDs[0] {
				t.Fatal("edit changed stable download ID")
			}
			updated, err := s.Request(ctx, 2, saved.RequestID, true)
			if err != nil {
				t.Fatal(err)
			}
			if updated.UserID != 1 || updated.Status != status || updated.Comment != input.Comment || updated.Date != input.Date || updated.Assignments[0].ModuleID != "11" {
				t.Fatal("edit lost ownership, status or corrected metadata")
			}
			object, err := s.Open(ctx, 2, saved.IDs[0])
			if err != nil {
				t.Fatal(err)
			}
			value, err := io.ReadAll(object)
			object.Close()
			if err != nil || string(value) != string(input.Documents[0]) {
				t.Fatal("stable download did not return corrected content")
			}
			object, err = s.Open(ctx, 3, saved.IDs[0])
			if status == Approved {
				if err != nil {
					t.Fatal(err)
				}
				object.Close()
			} else if !errors.Is(err, ErrNotFound) {
				t.Fatal("editing unpublished content made it public")
			}
			if _, err := s.EditRequest(ctx, 2, input, request.EditToken); !errors.Is(err, ErrConflict) {
				t.Fatal("stale edit overwrote a correction")
			}
			input.Documents = nil
			input.Comment = "Nur Metadaten"
			if _, err := s.EditRequest(ctx, 2, input, updated.EditToken); err != nil {
				t.Fatal(err)
			}
			var owner int64
			if err := db.QueryRow("SELECT user_id FROM archive WHERE id=?", saved.IDs[0]).Scan(&owner); err != nil || owner != 1 {
				t.Fatal("reviewer became the uploader")
			}
		})
	}
}

func TestReviewerEditRejectsChangedDecisionAndOldRevision(t *testing.T) {
	s, _ := reviewService(t)
	ctx := context.Background()
	input := submission()
	saved, err := s.SaveDraft(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SendRequest(ctx, 1, saved.RequestID); err != nil {
		t.Fatal(err)
	}
	request, err := s.Request(ctx, 2, saved.RequestID, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, saved.RequestID, true); err != nil {
		t.Fatal(err)
	}
	input.RequestID = saved.RequestID
	input.Documents = nil
	if _, err := s.EditRequest(ctx, 2, input, request.EditToken); !errors.Is(err, ErrConflict) {
		t.Fatal("edit ignored a concurrent moderation decision")
	}
	request, err = s.Request(ctx, 2, saved.RequestID, true)
	if err != nil {
		t.Fatal(err)
	}
	newer := submission()
	newer.RevisionOf = saved.IDs[0]
	newer.Documents = [][]byte{[]byte("%PDF-1.7\nnewer\n%%EOF")}
	revision, err := s.SaveDraft(ctx, 1, newer)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SendRequest(ctx, 1, revision.RequestID); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, revision.RequestID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := s.EditRequest(ctx, 2, input, request.EditToken); !errors.Is(err, ErrConflict) {
		t.Fatal("edited a superseded revision")
	}
}

func TestReviewerEditFailureRetainsFileAndMetadata(t *testing.T) {
	for _, mode := range []string{"storage", "database", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			s, db := reviewService(t)
			ctx := context.Background()
			input := submission()
			saved, err := s.SaveDraft(ctx, 1, input)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.SendRequest(ctx, 1, saved.RequestID); err != nil {
				t.Fatal(err)
			}
			request, err := s.Request(ctx, 2, saved.RequestID, true)
			if err != nil {
				t.Fatal(err)
			}
			storage := &failingStorage{writableStorage: s.storage.(writableStorage)}
			s.storage = storage
			switch mode {
			case "storage":
				storage.failAt = 1
			case "database":
				if _, err := db.Exec(`CREATE TRIGGER fail_edit BEFORE UPDATE OF comment ON archive BEGIN SELECT RAISE(FAIL,'forced edit failure'); END`); err != nil {
					t.Fatal(err)
				}
			}
			input.RequestID = saved.RequestID
			input.Comment = "Must roll back"
			input.Documents = [][]byte{[]byte("%PDF-1.7\nreplacement\n%%EOF")}
			if mode == "invalid" {
				input.Assignments = []Assignment{{ModuleID: "999", Version: "2023"}}
			}
			if _, err := s.EditRequest(ctx, 2, input, request.EditToken); err == nil {
				t.Fatal("invalid or failed edit reported success")
			}
			if mode == "invalid" && storage.calls != 0 {
				t.Fatal("invalid edit contacted storage")
			}
			updated, err := s.Request(ctx, 2, saved.RequestID, true)
			if err != nil || updated.EditToken != request.EditToken {
				t.Fatal("failed edit changed persisted state")
			}
			object, err := s.Open(ctx, 2, saved.IDs[0])
			if err != nil {
				t.Fatal(err)
			}
			value, err := io.ReadAll(object)
			object.Close()
			if err != nil || string(value) != string(submission().Documents[0]) {
				t.Fatal("failed edit replaced original bytes")
			}
		})
	}
}
