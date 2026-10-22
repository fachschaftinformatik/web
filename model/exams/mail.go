package exams

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/fachschaftinformatik/web/internal/audit"
	"github.com/fachschaftinformatik/web/internal/publicid"
)

// SubmissionMailPreference keeps reviewer authorization in the domain model.
func (s *Service) SubmissionMailPreference(ctx context.Context, userID int64) (enabled, permitted bool, err error) {
	role, err := accountRole(ctx, s.db, userID)
	if err != nil {
		return false, false, err
	}
	if !canReview(role) {
		return false, false, nil
	}
	err = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM submission_mail_subscriptions WHERE user_id=?)`, userID).Scan(&enabled)
	return enabled, true, err
}

func (s *Service) SaveSubmissionMailPreference(ctx context.Context, userID int64, enabled bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	role, err := accountRole(ctx, tx, userID)
	if err != nil {
		return err
	}
	if !canReview(role) {
		return ErrForbidden
	}
	if enabled {
		_, err = tx.ExecContext(ctx, `INSERT INTO submission_mail_subscriptions(user_id) VALUES(?) ON CONFLICT DO NOTHING`, userID)
	} else {
		_, err = tx.ExecContext(ctx, `DELETE FROM submission_mail_subscriptions WHERE user_id=?`, userID)
		if err == nil {
			_, err = tx.ExecContext(ctx, `DELETE FROM submission_mail_outbox WHERE user_id=?`, userID)
		}
	}
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	audit.Audit(ctx, "submission.mail_preference", "success", userID, slog.Bool("enabled", enabled))
	return nil
}

type SubmissionMailer interface {
	SendSubmissionEmail(to, name, submissionID string) error
}

type SubmissionDecisionMailer interface {
	SendSubmissionDecisionEmail(to, name, submissionID string, approved bool) error
}

// DeliverSubmissionDecisionMail notifies the owner using their current address.
// Only the latest undelivered decision is retained; failed deliveries are retried.
func (s *Service) DeliverSubmissionDecisionMail(ctx context.Context, mailer SubmissionDecisionMailer) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM submission_decision_mail_outbox AS o WHERE NOT EXISTS (
		SELECT 1 FROM archive_requests r JOIN users u ON u.id=r.user_id
		WHERE r.id=o.submission_id AND r.status=o.status
		AND u.active=1 AND u.verified=1 AND u.deleted_at IS NULL)`); err != nil {
		return err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT o.id,o.submission_id,o.status,u.id,u.email,u.name
		FROM submission_decision_mail_outbox o JOIN archive_requests r ON r.id=o.submission_id
		JOIN users u ON u.id=r.user_id
		WHERE r.status=o.status AND u.active=1 AND u.verified=1 AND u.deleted_at IS NULL
		ORDER BY o.id LIMIT 20`)
	if err != nil {
		return err
	}
	type message struct {
		id, userID             int64
		submissionID, to, name string
		status                 Status
	}
	var pending []message
	for rows.Next() {
		var m message
		if err := rows.Scan(&m.id, &m.submissionID, &m.status, &m.userID, &m.to, &m.name); err != nil {
			rows.Close()
			return err
		}
		pending = append(pending, m)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return err
	}
	var failures []error
	for _, m := range pending {
		if err := ctx.Err(); err != nil {
			return err
		}
		id, err := publicid.Lookup(ctx, s.db, "submissions", m.submissionID)
		if err != nil {
			return err
		}
		if err := mailer.SendSubmissionDecisionEmail(m.to, m.name, id, m.status == Approved); err != nil {
			audit.Audit(ctx, "submission.decision_mail", "smtp_failed", m.userID, slog.String("submission_id", id))
			failures = append(failures, fmt.Errorf("deliver submission decision mail: %w", err))
			continue
		}
		// A newer decision queued during SMTP delivery has a different ID.
		if _, err := s.db.ExecContext(ctx, `DELETE FROM submission_decision_mail_outbox WHERE id=?`, m.id); err != nil {
			return err
		}
		audit.Audit(ctx, "submission.decision_mail", "smtp_accepted", m.userID, slog.String("submission_id", id))
	}
	return errors.Join(failures...)
}

// DeliverSubmissionMail uses the current address and reviewer permissions at
// delivery time. Failed messages remain queued; drafts never enter this outbox.
func (s *Service) DeliverSubmissionMail(ctx context.Context, mailer SubmissionMailer) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM submission_mail_outbox AS o WHERE NOT EXISTS (
		SELECT 1 FROM users u JOIN submission_mail_subscriptions p ON p.user_id=u.id
		JOIN archive_requests r ON r.id=o.submission_id
		WHERE u.id=o.user_id AND u.active=1 AND u.verified=1 AND u.deleted_at IS NULL
		AND u.role IN ('admin','editor') AND r.status='pending')`); err != nil {
		return err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT o.submission_id,o.user_id,u.email,u.name
		FROM submission_mail_outbox o JOIN users u ON u.id=o.user_id
		JOIN submission_mail_subscriptions p ON p.user_id=u.id JOIN archive_requests r ON r.id=o.submission_id
		WHERE u.active=1 AND u.verified=1 AND u.deleted_at IS NULL AND u.role IN ('admin','editor') AND r.status='pending'
		ORDER BY o.submission_id,o.user_id LIMIT 20`)
	if err != nil {
		return err
	}
	type message struct {
		submissionID, to, name string
		userID                 int64
	}
	var pending []message
	for rows.Next() {
		var m message
		if err := rows.Scan(&m.submissionID, &m.userID, &m.to, &m.name); err != nil {
			rows.Close()
			return err
		}
		pending = append(pending, m)
	}
	err = errors.Join(rows.Err(), rows.Close())
	if err != nil {
		return err
	}
	var failures []error
	for _, m := range pending {
		if err := ctx.Err(); err != nil {
			return err
		}
		id, err := publicid.Lookup(ctx, s.db, "submissions", m.submissionID)
		if err != nil {
			return err
		}
		if err := mailer.SendSubmissionEmail(m.to, m.name, id); err != nil {
			audit.Audit(ctx, "submission.mail", "smtp_failed", m.userID, slog.String("submission_id", id))
			failures = append(failures, fmt.Errorf("deliver submission mail: %w", err))
			continue
		}
		if _, err := s.db.ExecContext(ctx, `DELETE FROM submission_mail_outbox WHERE submission_id=? AND user_id=?`, m.submissionID, m.userID); err != nil {
			return err
		}
		audit.Audit(ctx, "submission.mail", "smtp_accepted", m.userID, slog.String("submission_id", id))
	}
	return errors.Join(failures...)
}
