package posts

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// CommentPage keeps edit/vote redirects on the page containing the comment.
func (s *Service) CommentPage(ctx context.Context, post, comment string) (int, error) {
	var before int
	err := s.db.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM discussion_comments preceding WHERE preceding.post_id=c.post_id AND preceding.deleted_at IS NULL AND (preceding.created_at>c.created_at OR (preceding.created_at=c.created_at AND preceding.id>c.id))) FROM discussion_comments c JOIN discussion_posts p ON p.id=c.post_id AND p.deleted_at IS NULL WHERE c.id=? AND c.post_id=? AND c.deleted_at IS NULL`, comment, post).Scan(&before)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return before/20 + 1, err
}

// FindPage applies filters before limiting results, including tag searches.
func (s *Service) FindPage(ctx context.Context, f Filter, page, size int, viewer ...int64) ([]Post, int, error) {
	size = min(100, max(1, size))
	where := ""
	args := []any{}
	if f.Type != "" {
		where += " AND p.type=?"
		args = append(args, f.Type)
	}
	if f.AuthorID != 0 {
		where += " AND p.user_id=? AND u.private=0 AND u.deleted_at IS NULL"
		args = append(args, f.AuthorID)
	}
	for _, term := range strings.Fields(strings.ToLower(f.Query)) {
		where += ` AND (instr(lower(p.title||' '||p.body),?)>0 OR EXISTS(SELECT 1 FROM discussion_post_tags t WHERE t.post_id=p.id AND instr(lower(t.tag),?)>0))`
		args = append(args, term, term)
	}
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM discussion_posts p JOIN users u ON u.id=p.user_id WHERE p.deleted_at IS NULL`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	page = max(1, min(page, max(1, (total+size-1)/size)))
	params := append([]any{viewerID(viewer)}, args...)
	params = append(params, size, (page-1)*size)
	rows, err := s.db.QueryContext(ctx, postSelect+where+` ORDER BY p.pinned DESC,p.created_at DESC,p.id DESC LIMIT ? OFFSET ?`, params...)
	if err != nil {
		return nil, 0, err
	}
	var threads []Thread
	for rows.Next() {
		t, err := scanThread(rows)
		if err != nil {
			rows.Close()
			return nil, 0, err
		}
		threads = append(threads, t)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, 0, err
	}
	result := make([]Post, 0, len(threads))
	for _, t := range threads {
		if err := s.tags(ctx, &t); err != nil {
			return nil, 0, err
		}
		result = append(result, t.Post)
	}
	return result, total, nil
}
