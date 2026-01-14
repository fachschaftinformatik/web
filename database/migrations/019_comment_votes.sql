-- +goose Up
-- +goose StatementBegin
CREATE TABLE forum_comment_votes (
    comment_id TEXT NOT NULL REFERENCES forum_comments(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id),
    vote INTEGER NOT NULL, -- 1 or -1
    PRIMARY KEY (comment_id, user_id)
);

-- forum_posts already has updated_at, but we should ensure it's used correctly.
-- forum_comments already has updated_at (from migration 018).
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS forum_comment_votes;
-- +goose StatementEnd
