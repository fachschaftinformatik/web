package posts

import (
	"context"
	"database/sql"
	"errors"
)

func viewerID(viewer []int64) int64 {
	if len(viewer) > 0 {
		return viewer[0]
	}
	return 0
}

// Vote sets an explicit desired state so a retried request cannot toggle twice.
// Zero removes the viewer's vote. Existing votes by other users are preserved.
func (s *Service) Vote(ctx context.Context, user int64, post, comment string, value int) error {
	if value < -1 || value > 1 {
		return &ValidationError{[]FieldError{{"vote", "Wähle eine gültige Bewertung."}}}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := permission(ctx, tx, user, 0, false); err != nil {
		return err
	}
	var exists int
	err = tx.QueryRowContext(ctx, `SELECT 1 FROM discussion_posts WHERE id=? AND deleted_at IS NULL`, post).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	table, column, target := "discussion_votes", "post_id", post
	if comment != "" {
		err = tx.QueryRowContext(ctx, `SELECT 1 FROM discussion_comments WHERE id=? AND post_id=? AND deleted_at IS NULL`, comment, post).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		table, column, target = "discussion_comment_votes", "comment_id", comment
	}
	if value == 0 {
		_, err = tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE "+column+"=? AND user_id=?", target, user)
	} else {
		_, err = tx.ExecContext(ctx, "INSERT INTO "+table+" ("+column+",user_id,vote) VALUES(?,?,?) ON CONFLICT("+column+",user_id) DO UPDATE SET vote=excluded.vote", target, user, value)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}
