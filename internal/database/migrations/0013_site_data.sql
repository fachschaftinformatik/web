-- +goose Up
-- Migrate opens a dedicated connection with foreign keys disabled for table rebuilds.
CREATE TABLE users_unrestricted (
    id                 INTEGER PRIMARY KEY,
    email              TEXT NOT NULL UNIQUE,
    name               TEXT NOT NULL,
    password           TEXT NOT NULL,
    role               TEXT NOT NULL DEFAULT 'user' REFERENCES ref_user_roles(role),
    active             INTEGER NOT NULL DEFAULT 0 CHECK (active IN (0,1)),
    verified           INTEGER NOT NULL DEFAULT 0 CHECK (verified IN (0,1)),
    verified_at        TEXT,
    verification_token TEXT,
    program_id         INTEGER REFERENCES programs(id) ON DELETE RESTRICT,
    theme              TEXT NOT NULL DEFAULT 'system' CHECK (theme IN ('light', 'dark', 'system')),
    private            INTEGER NOT NULL DEFAULT 0 CHECK (private IN (0,1)),
    avatar_url         TEXT,
    created_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at         TEXT,
    password_reset_token TEXT,
    password_reset_expires TEXT
) STRICT;
INSERT INTO users_unrestricted SELECT * FROM users;
DROP TABLE users;
ALTER TABLE users_unrestricted RENAME TO users;
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;
-- +goose StatementBegin
CREATE TRIGGER trg_users_updated_at AFTER UPDATE ON users BEGIN
    UPDATE users SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = OLD.id;
END;
-- +goose StatementEnd

CREATE TABLE organization (
    id INTEGER PRIMARY KEY CHECK(id=1),
    data TEXT NOT NULL CHECK(json_valid(data))
) STRICT;
CREATE TABLE historical_meeting_documents (
    file TEXT PRIMARY KEY,
    year INTEGER NOT NULL,
    number INTEGER NOT NULL,
    date TEXT NOT NULL,
    kind TEXT NOT NULL,
    note TEXT NOT NULL,
    bytes INTEGER NOT NULL,
    sha256 TEXT NOT NULL,
    sources TEXT NOT NULL CHECK(json_valid(sources)),
    blob_key TEXT NOT NULL UNIQUE
) STRICT;

-- +goose Down
-- Do not reintroduce a restriction that could invalidate existing accounts.
DROP TABLE historical_meeting_documents;
DROP TABLE organization;
