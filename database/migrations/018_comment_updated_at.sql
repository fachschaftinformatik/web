-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE forum_comments_new (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL REFERENCES forum_posts(id) ON DELETE CASCADE,
    author_id TEXT NOT NULL REFERENCES users(id),
    parent_id TEXT REFERENCES forum_comments_new(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    active INTEGER NOT NULL DEFAULT 1
);

INSERT INTO forum_comments_new (id, post_id, author_id, parent_id, text, created_at, updated_at, active)
SELECT id, post_id, author_id, parent_id, text, created_at, created_at, active FROM forum_comments;

DROP TABLE forum_comments;

ALTER TABLE forum_comments_new RENAME TO forum_comments;

PRAGMA foreign_keys=ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE forum_comments_old (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL REFERENCES forum_posts(id) ON DELETE CASCADE,
    author_id TEXT NOT NULL REFERENCES users(id),
    parent_id TEXT REFERENCES forum_comments_old(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    active INTEGER NOT NULL DEFAULT 1
);

INSERT INTO forum_comments_old (id, post_id, author_id, parent_id, text, created_at, active)
SELECT id, post_id, author_id, parent_id, text, created_at, active FROM forum_comments;

DROP TABLE forum_comments;

ALTER TABLE forum_comments_old RENAME TO forum_comments;

PRAGMA foreign_keys=ON;
-- +goose StatementEnd
