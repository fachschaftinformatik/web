-- +goose Up
-- +goose StatementBegin

CREATE TABLE ref_user_roles (
    role TEXT PRIMARY KEY
) STRICT;

INSERT INTO ref_user_roles (role) VALUES ('admin'), ('editor'), ('user');

CREATE TABLE ref_discussion_types (
    type TEXT PRIMARY KEY
) STRICT;

INSERT INTO ref_discussion_types (type) VALUES ('discussion'), ('news'), ('event');

CREATE TABLE ref_notification_types (
    type TEXT PRIMARY KEY
) STRICT;

INSERT INTO ref_notification_types (type) VALUES ('archive'), ('discussion'), ('news');

CREATE TABLE ref_activity_types (
    type TEXT PRIMARY KEY
) STRICT;

INSERT INTO ref_activity_types (type) VALUES 
('POST_CREATED'), 
('COMMENT_ADDED'), 
('ARCHIVE_UPLOADED'),
('EXAM_UPLOADED'),
('EVENT_CREATED'),
('MEDIA_UPLOADED');

CREATE TABLE programs (
    id         INTEGER PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at TEXT
) STRICT;

CREATE TABLE program_versions (
    program_id INTEGER NOT NULL REFERENCES programs(id) ON DELETE RESTRICT,
    name       TEXT NOT NULL,
    PRIMARY KEY (program_id, name)
) STRICT;

CREATE TABLE modules (
    id         INTEGER PRIMARY KEY,
    program_id INTEGER NOT NULL REFERENCES programs(id) ON DELETE RESTRICT,
    name       TEXT NOT NULL,
    alias      TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at TEXT,
    UNIQUE(program_id, name)
) STRICT;

CREATE TABLE users (
    id                 INTEGER PRIMARY KEY,
    email              TEXT NOT NULL UNIQUE CHECK(email LIKE '%@studmail.w-hs.de' OR email LIKE '%@fsv-wh.de'),
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
    deleted_at         TEXT
) STRICT;

CREATE TABLE sessions (
    id         TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    user_agent TEXT,
    ip_address TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    last_seen  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    expires_at TEXT NOT NULL
) STRICT;

CREATE TABLE archive (
    id           INTEGER PRIMARY KEY,
    group_id     INTEGER NOT NULL,
    user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    module_id    INTEGER NOT NULL REFERENCES modules(id) ON DELETE RESTRICT,
    version      TEXT NOT NULL, -- PO version 
    exam_date    TEXT NOT NULL,
    comment      TEXT,
    edit_version INTEGER NOT NULL DEFAULT 1,
    is_latest    INTEGER NOT NULL DEFAULT 1 CHECK (is_latest IN (0,1)),
    access_key   TEXT NOT NULL UNIQUE,
    mime_type    TEXT NOT NULL CHECK (mime_type IN ('application/pdf')),
    nbytes       INTEGER NOT NULL,
    checksum     TEXT NOT NULL,
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at   TEXT
) STRICT;

CREATE TABLE events (
    id         INTEGER PRIMARY KEY,
    title      TEXT NOT NULL,
    cover_path TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at TEXT
) STRICT;

CREATE TABLE media (
    id          INTEGER PRIMARY KEY,
    event_id    INTEGER NOT NULL REFERENCES events(id) ON DELETE RESTRICT,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title       TEXT,
    description TEXT,
    access_key  TEXT NOT NULL UNIQUE,
    mime_type   TEXT NOT NULL,
    nbytes      INTEGER NOT NULL,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at  TEXT
) STRICT;

CREATE TABLE discussion_posts (
    id         INTEGER PRIMARY KEY,
    title      TEXT NOT NULL,
    body       TEXT NOT NULL,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    pinned     INTEGER NOT NULL DEFAULT 0 CHECK (pinned IN (0,1)),
    type       TEXT NOT NULL DEFAULT 'discussion' REFERENCES ref_discussion_types(type),
    event_date TEXT,
    location   TEXT,
    image_url  TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at TEXT
) STRICT;

CREATE TABLE discussion_post_programs (
    post_id    INTEGER NOT NULL REFERENCES discussion_posts(id) ON DELETE CASCADE,
    program_id INTEGER NOT NULL REFERENCES programs(id) ON DELETE RESTRICT,
    PRIMARY KEY (post_id, program_id)
) STRICT;

CREATE TABLE discussion_post_tags (
    post_id INTEGER NOT NULL REFERENCES discussion_posts(id) ON DELETE CASCADE,
    tag     TEXT NOT NULL,
    PRIMARY KEY (post_id, tag)
) STRICT;

CREATE TABLE discussion_post_links (
    id      INTEGER PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES discussion_posts(id) ON DELETE CASCADE,
    url     TEXT NOT NULL,
    label   TEXT
) STRICT;

CREATE TABLE discussion_comments (
    id         INTEGER PRIMARY KEY,
    post_id    INTEGER NOT NULL REFERENCES discussion_posts(id) ON DELETE RESTRICT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    parent_id  INTEGER REFERENCES discussion_comments(id) ON DELETE RESTRICT,
    text       TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at TEXT
) STRICT;

CREATE TABLE discussion_votes (
    post_id INTEGER NOT NULL REFERENCES discussion_posts(id) ON DELETE RESTRICT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    vote    INTEGER NOT NULL CHECK (vote IN (1, -1)),
    PRIMARY KEY (post_id, user_id)
) STRICT;

CREATE TABLE discussion_comment_votes (
    comment_id INTEGER NOT NULL REFERENCES discussion_comments(id) ON DELETE RESTRICT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    vote       INTEGER NOT NULL CHECK (vote IN (1, -1)),
    PRIMARY KEY (comment_id, user_id)
) STRICT;

CREATE TABLE notifications (
    id         INTEGER PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title      TEXT NOT NULL,
    message    TEXT NOT NULL,
    type       TEXT NOT NULL REFERENCES ref_notification_types(type),
    link       TEXT NOT NULL,
    read       INTEGER NOT NULL DEFAULT 0 CHECK (read IN (0,1)),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at TEXT
) STRICT;

CREATE TABLE activities (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    type        TEXT NOT NULL REFERENCES ref_activity_types(type),
    target_id   TEXT NOT NULL,
    target_name TEXT,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at  TEXT
) STRICT;

CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_programs_deleted_at ON programs(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_modules_deleted_at ON modules(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_archive_deleted_at ON archive(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_archive_group_id ON archive(group_id);
CREATE INDEX idx_events_deleted_at ON events(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_deleted_at ON media(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_discussion_posts_deleted_at ON discussion_posts(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_discussion_comments_deleted_at ON discussion_comments(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_notifications_deleted_at ON notifications(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_activities_deleted_at ON activities(deleted_at) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_programs_updated_at AFTER UPDATE ON programs BEGIN
    UPDATE programs SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = OLD.id;
END;

CREATE TRIGGER trg_modules_updated_at AFTER UPDATE ON modules BEGIN
    UPDATE modules SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = OLD.id;
END;

CREATE TRIGGER trg_users_updated_at AFTER UPDATE ON users BEGIN
    UPDATE users SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = OLD.id;
END;

CREATE TRIGGER trg_archive_updated_at AFTER UPDATE ON archive BEGIN
    UPDATE archive SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = OLD.id;
END;

CREATE TRIGGER trg_events_updated_at AFTER UPDATE ON events BEGIN
    UPDATE events SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = OLD.id;
END;

CREATE TRIGGER trg_media_updated_at AFTER UPDATE ON media BEGIN
    UPDATE media SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = OLD.id;
END;

CREATE TRIGGER trg_discussion_posts_updated_at AFTER UPDATE ON discussion_posts BEGIN
    UPDATE discussion_posts SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = OLD.id;
END;

CREATE TRIGGER trg_discussion_comments_updated_at AFTER UPDATE ON discussion_comments BEGIN
    UPDATE discussion_comments SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = OLD.id;
END;

CREATE TRIGGER trg_notifications_updated_at AFTER UPDATE ON notifications BEGIN
    UPDATE notifications SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = OLD.id;
END;

CREATE TRIGGER trg_activities_updated_at AFTER UPDATE ON activities BEGIN
    UPDATE activities SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = OLD.id;
END;

CREATE TRIGGER trg_archive_validate_version
BEFORE INSERT ON archive
FOR EACH ROW
BEGIN
    SELECT CASE
        WHEN (SELECT COUNT(*) FROM program_versions pv 
              JOIN modules m ON m.program_id = pv.program_id 
              WHERE m.id = NEW.module_id AND pv.name = NEW.version) = 0
        THEN RAISE(ABORT, 'Invalid version for the program associated with this module')
    END;
END;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS discussion_comment_votes;
DROP TABLE IF EXISTS discussion_votes;
DROP TABLE IF EXISTS discussion_comments;
DROP TABLE IF EXISTS discussion_post_links;
DROP TABLE IF EXISTS discussion_post_tags;
DROP TABLE IF EXISTS discussion_post_programs;
DROP TABLE IF EXISTS discussion_posts;
DROP TABLE IF EXISTS media;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS archive;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS modules;
DROP TABLE IF EXISTS program_versions;
DROP TABLE IF EXISTS programs;
DROP TABLE IF EXISTS ref_activity_types;
DROP TABLE IF EXISTS ref_notification_types;
DROP TABLE IF EXISTS ref_discussion_types;
DROP TABLE IF EXISTS ref_user_roles;
-- +goose StatementEnd
