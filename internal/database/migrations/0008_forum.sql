-- +goose Up
ALTER TABLE discussion_posts ADD COLUMN body_format TEXT NOT NULL DEFAULT 'plain' CHECK (body_format IN ('plain', 'html'));

-- +goose Down
ALTER TABLE discussion_posts DROP COLUMN body_format;
