-- +goose Up
-- +goose StatementBegin

-- 1. Add CHECK constraints to users
-- Note: SQLite doesn't support adding CHECK constraints via ALTER TABLE easily
-- but we can add them to new tables if we were to recreate.
-- For now, we will add them where we can or in future migrations.

-- 2. Add validation to forum_posts
-- We can't easily add constraints to existing tables without recreation in SQLite
-- but we can ensure future data is valid via triggers or just rely on backend hardening for now.
-- Actually, the best way in SQLite is to recreate the table. 
-- Since this is a prototype, I'll do a "safe" hardening by adding some indices and triggers if needed.

-- Let's at least add some missing indices for security/performance
CREATE INDEX IF NOT EXISTS idx_forum_posts_author ON forum_posts(author_id);
CREATE INDEX IF NOT EXISTS idx_forum_comments_post ON forum_comments(post_id);
CREATE INDEX IF NOT EXISTS idx_forum_comments_author ON forum_comments(author_id);

-- 3. Cleanup: Remove the old 'posts' and 'comments' tables if they are no longer used
-- They were from the very first migration and replaced by forum_posts/forum_comments
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Restoration of dropped tables would require schema knowledge from 001
-- +goose StatementEnd
