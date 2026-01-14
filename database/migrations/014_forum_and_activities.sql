-- +goose Up
-- +goose StatementBegin
-- Forum and Activities Migration

CREATE TABLE forum_posts (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES users(id),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    pinned INTEGER NOT NULL DEFAULT 0,
    type TEXT NOT NULL DEFAULT 'forum', -- 'forum', 'news', 'event'
    programs TEXT NOT NULL DEFAULT '[]', -- JSON array of program IDs
    tags TEXT NOT NULL DEFAULT '[]', -- JSON array of tags
    event_date TEXT, -- Optional for events
    location TEXT, -- Optional for events
    image_url TEXT, -- Optional image
    links TEXT NOT NULL DEFAULT '[]' -- JSON array of external links
);

CREATE TABLE forum_comments (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL REFERENCES forum_posts(id) ON DELETE CASCADE,
    author_id TEXT NOT NULL REFERENCES users(id),
    parent_id TEXT REFERENCES forum_comments(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE forum_votes (
    post_id TEXT NOT NULL REFERENCES forum_posts(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id),
    vote INTEGER NOT NULL, -- 1 or -1
    PRIMARY KEY (post_id, user_id)
);

CREATE TABLE activities (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    type TEXT NOT NULL, -- 'POST_CREATED', 'COMMENT_ADDED', 'EXAM_UPLOADED'
    target_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- Index for faster activity lookups by user
CREATE INDEX idx_activities_user_created ON activities(user_id, created_at DESC);
CREATE INDEX idx_forum_posts_type_created ON forum_posts(type, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_forum_posts_type_created;
DROP INDEX IF EXISTS idx_activities_user_created;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS forum_votes;
DROP TABLE IF EXISTS forum_comments;
DROP TABLE IF EXISTS forum_posts;
-- +goose StatementEnd
