-- +goose Up
-- +goose StatementBegin

-- 1. Migrate Events (already used Snowflake IDs in PostEvents but schema was INTEGER)
CREATE TABLE events_new (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  cover_path TEXT,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
) STRICT;

INSERT INTO events_new (id, title, cover_path, created_at)
SELECT CAST(id AS TEXT), title, cover_path, created_at FROM events;

-- 2. Migrate Programs
CREATE TABLE programs_new (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
) STRICT;

INSERT INTO programs_new (id, name, created_at)
SELECT CAST(id AS TEXT), name, created_at FROM programs;

-- 3. Migrate Program Versions
CREATE TABLE program_versions_new (
  programid TEXT NOT NULL REFERENCES programs_new(id) ON DELETE CASCADE,
  name TEXT NOT NULL CHECK (name in ('PO2016','PO2023')),
  PRIMARY KEY (programid, name)
) STRICT;

INSERT INTO program_versions_new (programid, name)
SELECT CAST(programid AS TEXT), name FROM program_versions;

-- 4. Migrate Modules
CREATE TABLE modules_new (
  id TEXT PRIMARY KEY,
  programid TEXT NOT NULL REFERENCES programs_new(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  alias TEXT,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE(programid, name)
) STRICT;

INSERT INTO modules_new (id, programid, name, alias, created_at)
SELECT CAST(id AS TEXT), CAST(programid AS TEXT), name, alias, created_at FROM modules;

-- 5. Update Users (programid FK)
CREATE TABLE users_new (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL,
  name TEXT NOT NULL,
  password TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user','editor','admin')),
  active INTEGER NOT NULL DEFAULT 0 CHECK (active IN (0,1)),
  verified INTEGER NOT NULL DEFAULT 0 CHECK (verified IN (0,1)),
  verified_at TEXT,
  verified_until TEXT,
  programid TEXT NOT NULL REFERENCES programs_new(id) ON DELETE RESTRICT ON UPDATE CASCADE,
  theme TEXT NOT NULL DEFAULT 'system',
  private INTEGER NOT NULL DEFAULT 0 CHECK (private IN (0,1)),
  avatar_url TEXT,
  verification_token TEXT,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  CHECK (verified = 0 OR verified_at IS NOT NULL),
  CHECK (verified_at IS NULL OR verified_until IS NULL OR verified_until >= verified_at)
) STRICT;

INSERT INTO users_new 
SELECT id, email, name, password, role, active, verified, verified_at, verified_until, CAST(programid AS TEXT), theme, private, avatar_url, verification_token, created_at, updated_at 
FROM users;

-- 6. Update Exams (programid, moduleid FKs)
CREATE TABLE exams_new (
  id TEXT PRIMARY KEY,
  userid TEXT NOT NULL REFERENCES users_new(id) ON DELETE RESTRICT ON UPDATE CASCADE,
  programid TEXT NOT NULL,
  version TEXT NOT NULL,
  moduleid TEXT REFERENCES modules_new(id) ON DELETE RESTRICT,
  comment TEXT,
  exam_date TEXT NOT NULL,
  uploaded_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  accesskey TEXT NOT NULL,
  mime_type TEXT NOT NULL CHECK (mime_type IN ('application/pdf')),
  nbytes INTEGER NOT NULL,
  checksum TEXT NOT NULL,
  group_id TEXT NOT NULL,
  edit_version INTEGER NOT NULL DEFAULT 1,
  is_latest INTEGER NOT NULL DEFAULT 1,
  FOREIGN KEY (programid, version) REFERENCES program_versions_new(programid, name) ON DELETE RESTRICT ON UPDATE CASCADE
) STRICT;

INSERT INTO exams_new 
SELECT id, userid, CAST(programid AS TEXT), version, CAST(moduleid AS TEXT), comment, exam_date, uploaded_at, accesskey, mime_type, nbytes, checksum, group_id, edit_version, is_latest 
FROM exams;

-- 7. Update Media (event_id FK)
CREATE TABLE media_new (
  id TEXT PRIMARY KEY,
  event_id TEXT NOT NULL REFERENCES events_new(id) ON DELETE CASCADE,
  userid TEXT NOT NULL REFERENCES users_new(id) ON DELETE RESTRICT,
  title TEXT,
  description TEXT,
  accesskey TEXT NOT NULL UNIQUE,
  mime_type TEXT NOT NULL,
  nbytes INTEGER NOT NULL,
  uploaded_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
) STRICT;

INSERT INTO media_new 
SELECT id, CAST(event_id AS TEXT), userid, title, description, accesskey, mime_type, nbytes, uploaded_at 
FROM media;

-- 8. Switch Sessions to Snowflake IDs
DROP TABLE sessions;
CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  userid TEXT NOT NULL REFERENCES users_new(id) ON DELETE CASCADE ON UPDATE CASCADE,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  last_seen TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  expires_at TEXT NOT NULL,
  user_agent TEXT,
  ip_address TEXT
) STRICT;

-- Drop and Rename
DROP TABLE media;
DROP TABLE exams;
DROP TABLE users;
DROP TABLE modules;
DROP TABLE program_versions;
DROP TABLE programs;
DROP TABLE events;

ALTER TABLE events_new RENAME TO events;
ALTER TABLE programs_new RENAME TO programs;
ALTER TABLE program_versions_new RENAME TO program_versions;
ALTER TABLE modules_new RENAME TO modules;
ALTER TABLE users_new RENAME TO users;
ALTER TABLE exams_new RENAME TO exams;
ALTER TABLE media_new RENAME TO media;

-- Recreate indices
CREATE UNIQUE INDEX users_email_unique ON users(lower(email));
CREATE INDEX idx_users_program ON users(programid);
CREATE INDEX idx_users_verified_until ON users(verified_until);
CREATE INDEX idx_exams_date ON exams(programid, exam_date DESC);
CREATE INDEX idx_exams_user ON exams(userid);
CREATE INDEX idx_exams_module ON exams(moduleid);
CREATE INDEX idx_media_event ON media(event_id, uploaded_at DESC);
CREATE INDEX idx_sessions_user ON sessions(userid);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Down migration not fully supported for large schema refactors
-- +goose StatementEnd
