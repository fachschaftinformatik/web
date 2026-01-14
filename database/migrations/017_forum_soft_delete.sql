-- +goose Up
-- +goose StatementBegin
ALTER TABLE forum_posts ADD COLUMN active INTEGER NOT NULL DEFAULT 1;
ALTER TABLE forum_comments ADD COLUMN active INTEGER NOT NULL DEFAULT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE forum_posts DROP COLUMN active;
ALTER TABLE forum_comments DROP COLUMN active;
-- +goose StatementEnd
