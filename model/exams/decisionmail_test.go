package exams

import (
	"context"
	"errors"
	"testing"
)

type decisionMailRecorder struct {
	submissionMailRecorder
	approved bool
}

func (m *decisionMailRecorder) SendSubmissionDecisionEmail(to, name, id string, approved bool) error {
	m.approved = approved
	return m.SendSubmissionEmail(to, name, id)
}

func TestSubmissionDecisionMailRetryAndReversal(t *testing.T) {
	for _, approved := range []bool{true, false} {
		t.Run(map[bool]string{true: "approved", false: "rejected"}[approved], func(t *testing.T) {
			s, db := reviewService(t)
			ctx := context.Background()
			draft, err := s.SaveDraft(ctx, 1, submission(t, db))
			if err != nil {
				t.Fatal(err)
			}
			mailer := &decisionMailRecorder{}
			if err := s.Submit(ctx, 1, draft.SubmissionID); err != nil {
				t.Fatal(err)
			}
			if err := s.DeliverSubmissionDecisionMail(ctx, mailer); err != nil || mailer.calls != 0 {
				t.Fatal("pending submission sent a decision", err)
			}
			for range 2 {
				if err := s.Review(ctx, 2, draft.SubmissionID, approved); err != nil {
					t.Fatal(err)
				}
			}
			mailer.err = errors.New("SMTP unavailable")
			if err := s.DeliverSubmissionDecisionMail(ctx, mailer); err == nil || mailer.calls != 1 {
				t.Fatal("failed decision mail was not attempted", err)
			}
			if _, err := db.Exec(`UPDATE users SET email='submitter@example.org',name='Submitter' WHERE id=1`); err != nil {
				t.Fatal(err)
			}
			mailer.err = nil
			if err := s.DeliverSubmissionDecisionMail(ctx, mailer); err != nil {
				t.Fatal(err)
			}
			if mailer.calls != 2 || mailer.to != "submitter@example.org" || mailer.name != "Submitter" || mailer.id != draft.SubmissionID || mailer.approved != approved {
				t.Fatalf("incorrect decision recipient or contents: %+v", mailer)
			}
			if err := s.DeliverSubmissionDecisionMail(ctx, mailer); err != nil || mailer.calls != 2 {
				t.Fatal("delivered decision repeated", err)
			}
			// Reversals notify the owner, coalescing undelivered intermediate decisions.
			for _, decision := range []bool{!approved, approved, !approved} {
				if err := s.Review(ctx, 2, draft.SubmissionID, decision); err != nil {
					t.Fatal(err)
				}
			}
			if err := s.DeliverSubmissionDecisionMail(ctx, mailer); err != nil || mailer.calls != 3 || mailer.approved == approved {
				t.Fatal("latest reversed decision not delivered", err)
			}
		})
	}
}

func TestReviewAndDecisionMailCommitTogether(t *testing.T) {
	s, db := reviewService(t)
	ctx := context.Background()
	draft, err := s.SaveDraft(ctx, 1, submission(t, db))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(ctx, 1, draft.SubmissionID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_decision_mail BEFORE INSERT ON submission_decision_mail_outbox BEGIN SELECT RAISE(FAIL,'outbox unavailable'); END`); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, draft.SubmissionID, true); err == nil {
		t.Fatal("review succeeded without queuing its notification")
	}
	current, err := s.Submission(ctx, 1, draft.SubmissionID, false)
	if err != nil || current.Status != Pending {
		t.Fatal("failed queue insertion did not roll back review", err)
	}
	var latest int
	if err := db.QueryRow(`SELECT is_latest FROM archive WHERE request_id=?`, testInternalID(t, db, "submissions", draft.SubmissionID)).Scan(&latest); err != nil || latest != 0 {
		t.Fatal("failed queue insertion did not roll back publication", err)
	}
}

func TestDecisionMailDropsInactiveOwners(t *testing.T) {
	for _, change := range []string{
		`UPDATE users SET active=0 WHERE id=1`,
		`UPDATE users SET verified=0 WHERE id=1`,
		`UPDATE users SET deleted_at='2026-10-22T00:00:00Z' WHERE id=1`,
	} {
		t.Run(change, func(t *testing.T) {
			s, db := reviewService(t)
			ctx := context.Background()
			draft, err := s.SaveDraft(ctx, 1, submission(t, db))
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Submit(ctx, 1, draft.SubmissionID); err != nil {
				t.Fatal(err)
			}
			if err := s.Review(ctx, 2, draft.SubmissionID, false); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(change); err != nil {
				t.Fatal(err)
			}
			mailer := &decisionMailRecorder{}
			if err := s.DeliverSubmissionDecisionMail(ctx, mailer); err != nil || mailer.calls != 0 {
				t.Fatal("ineligible owner received mail", err)
			}
			var count int
			if err := db.QueryRow(`SELECT count(*) FROM submission_decision_mail_outbox`).Scan(&count); err != nil || count != 0 {
				t.Fatal("ineligible message retained", err)
			}
		})
	}
}
