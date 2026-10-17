-- +goose Up
CREATE TABLE web_sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    expires_at TEXT NOT NULL
) STRICT;
CREATE INDEX web_sessions_user ON web_sessions(user_id);
CREATE INDEX web_sessions_expiry ON web_sessions(expires_at);

CREATE TABLE user_tokens (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('verify', 'reset')),
    expires_at TEXT NOT NULL,
    UNIQUE(user_id, kind)
) STRICT;

CREATE TABLE mail_outbox (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('verify', 'reset')),
    recipient TEXT NOT NULL,
    name TEXT NOT NULL,
    link TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
) STRICT;

-- +goose Down
DROP TABLE mail_outbox;
DROP TABLE user_tokens;
DROP TABLE web_sessions;
