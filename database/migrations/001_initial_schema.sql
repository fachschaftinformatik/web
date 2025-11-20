-- +goose Up
-- +goose StatementBegin

CREATE TABLE programs (
  id         INTEGER PRIMARY KEY,
  name       TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
) STRICT;

CREATE TABLE program_versions (
  programid INTEGER NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
  name      TEXT NOT NULL CHECK (name in ('PO2016','PO2023')),
  PRIMARY KEY (programid, name)
) STRICT;

CREATE TABLE users (
  id             TEXT PRIMARY KEY,
  email          TEXT NOT NULL,
  name           TEXT NOT NULL,
  password       TEXT NOT NULL,
  role           TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user','editor','admin')),
  active         INTEGER NOT NULL DEFAULT 0 CHECK (active IN (0,1)),
  verified       INTEGER NOT NULL DEFAULT 0 CHECK (verified IN (0,1)),
  verified_at    TEXT,
  verified_until TEXT,
  programid      INTEGER NOT NULL
                   REFERENCES programs(id)
                   ON DELETE RESTRICT ON UPDATE CASCADE,
  created_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  CHECK (verified = 0 OR verified_at IS NOT NULL),
  CHECK (verified_at IS NULL OR verified_until IS NULL OR verified_until >= verified_at)
) STRICT;

CREATE UNIQUE INDEX users_email_unique ON users(lower(email));
CREATE INDEX idx_users_program         ON users(programid);
CREATE INDEX idx_users_verified_until  ON users(verified_until);

CREATE TRIGGER trg_users_update
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
  UPDATE users
     SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
   WHERE id = OLD.id;
END;

CREATE TABLE posts (
  id         TEXT PRIMARY KEY,
  userid     TEXT NOT NULL
               REFERENCES users(id)
               ON DELETE RESTRICT ON UPDATE CASCADE,
  title      TEXT NOT NULL,
  body       TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  deleted    TEXT
) STRICT;

CREATE INDEX idx_posts_user       ON posts(userid);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);

CREATE TRIGGER trg_posts_update
AFTER UPDATE ON posts
FOR EACH ROW
BEGIN
  UPDATE posts
     SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
   WHERE id = OLD.id;
END;

CREATE TABLE comments (
  id         TEXT PRIMARY KEY,
  postid     TEXT NOT NULL
               REFERENCES posts(id)
               ON DELETE CASCADE ON UPDATE CASCADE,
  userid     TEXT NOT NULL
               REFERENCES users(id)
               ON DELETE RESTRICT ON UPDATE CASCADE,
  body       TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-m-%dT%H:%M:%fZ','now'))
) STRICT;

CREATE INDEX idx_comments_post        ON comments(postid);
CREATE INDEX idx_comments_user        ON comments(userid);
CREATE INDEX idx_comments_created_at  ON comments(created_at DESC);

CREATE TRIGGER trg_comments_update
AFTER UPDATE ON comments
FOR EACH ROW
BEGIN
  UPDATE comments
     SET updated_at = strftime('%Y-m-%dT%H:%M:%fZ','now')
   WHERE id = OLD.id;
END;

CREATE TABLE exams (
  id           TEXT PRIMARY KEY,
  userid       TEXT NOT NULL
                 REFERENCES users(id)
                 ON DELETE RESTRICT ON UPDATE CASCADE,
  programid    INTEGER NOT NULL,
  version      TEXT NOT NULL,
  exam_date    TEXT NOT NULL,
  uploaded_at  TEXT NOT NULL DEFAULT (strftime('%Y-m-%dT%H:%M:%fZ','now')),
  accesskey    TEXT NOT NULL UNIQUE,
  mime_type    TEXT NOT NULL CHECK (mime_type IN ('application/pdf')),
  nbytes       INTEGER NOT NULL,
  checksum     TEXT NOT NULL,
  -- Enforce that the (programid, version) tuple actually exists in the valid versions table
  FOREIGN KEY (programid, version) REFERENCES program_versions(programid, name) ON DELETE RESTRICT ON UPDATE CASCADE
) STRICT;

CREATE INDEX idx_exams_date ON exams(programid, exam_date DESC);
CREATE INDEX idx_exams_user ON exams(userid);

CREATE TRIGGER trg_exams_set_update
AFTER UPDATE ON exams
FOR EACH ROW
BEGIN
  UPDATE exams
     SET uploaded_at = uploaded_at,
         accesskey   = accesskey
   WHERE id = OLD.id;
END;

CREATE TABLE sessions (
  id         TEXT PRIMARY KEY,
  userid     TEXT NOT NULL
               REFERENCES users(id)
               ON DELETE CASCADE ON UPDATE CASCADE,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-m-%dT%H:%M:%fZ','now')),
  last_seen  TEXT NOT NULL DEFAULT (strftime('%Y-m-%dT%H:%M:%fZ','now')),
  expires_at TEXT NOT NULL
) STRICT;

CREATE INDEX idx_sessions_user        ON sessions(userid);
CREATE INDEX idx_sessions_expires_at  ON sessions(expires_at);

-- 1. Insert Programs
INSERT OR IGNORE INTO programs (name) VALUES
('Informatik (B. Sc.)'),
('Informatik und Design (B. Sc.)'),
('Wirtschaftsinformatik (B. Sc.)'),
('Informatik (M. Sc.)'),
('Medieninformatik (M. Sc.)'),
('Wirtschaftsinformatik (M. Sc.)'),
('Internetsicherheit (M. Sc.)');

-- 2. Insert Versions for Programs
-- Define standard programs getting both POs
INSERT OR IGNORE INTO program_versions (programid, name)
SELECT id, 'PO2016' FROM programs WHERE name IN (
    'Informatik (B. Sc.)',
    'Wirtschaftsinformatik (B. Sc.)',
    'Informatik (M. Sc.)',
    'Medieninformatik (M. Sc.)',
    'Wirtschaftsinformatik (M. Sc.)',
    'Internetsicherheit (M. Sc.)'
);

INSERT OR IGNORE INTO program_versions (programid, name)
SELECT id, 'PO2023' FROM programs WHERE name IN (
    'Informatik (B. Sc.)',
    'Wirtschaftsinformatik (B. Sc.)',
    'Informatik (M. Sc.)',
    'Medieninformatik (M. Sc.)',
    'Wirtschaftsinformatik (M. Sc.)',
    'Internetsicherheit (M. Sc.)'
);

-- Specific case: Informatik und Design only gets PO2023
INSERT OR IGNORE INTO program_versions (programid, name)
SELECT id, 'PO2023' FROM programs WHERE name = 'Informatik und Design (B. Sc.)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sessions;
DROP TABLE exams;
DROP TABLE comments;
DROP TABLE posts;
DROP TABLE users;
DROP TABLE program_versions;
DROP TABLE programs;
-- +goose StatementEnd
