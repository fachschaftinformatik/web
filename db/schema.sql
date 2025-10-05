PRAGMA user_version = 3;

PRAGMA foreign_keys = ON;
PRAGMA recursive_triggers = OFF;
PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS campuses (
  id         INTEGER PRIMARY KEY,
  name       TEXT NOT NULL UNIQUE,
  location   TEXT,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
) STRICT;

CREATE TABLE IF NOT EXISTS disciplines (
  id         INTEGER PRIMARY KEY,
  name       TEXT NOT NULL,
  degree     TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE(degree, name)
) STRICT;

CREATE TABLE IF NOT EXISTS users (
  id             TEXT PRIMARY KEY,
  email          TEXT NOT NULL,
  name           TEXT NOT NULL,
  password       TEXT NOT NULL,
  role           TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user','editor','admin')),
  active         INTEGER NOT NULL DEFAULT 0 CHECK (active IN (0,1)),
  verified       INTEGER NOT NULL DEFAULT 0 CHECK (verified IN (0,1)),
  verified_at    TEXT,
  verified_until TEXT,
  campusid       INTEGER NOT NULL
                   REFERENCES campuses(id)
                   ON DELETE RESTRICT ON UPDATE CASCADE,
  disciplineid   INTEGER NOT NULL
                   REFERENCES disciplines(id)
                   ON DELETE RESTRICT ON UPDATE CASCADE,
  created_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  CHECK (verified = 0 OR verified_at IS NOT NULL),
  CHECK (verified_at IS NULL OR verified_until IS NULL OR verified_until >= verified_at)
) STRICT;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users(lower(email));
CREATE INDEX IF NOT EXISTS idx_users_campus          ON users(campusid);
CREATE INDEX IF NOT EXISTS idx_users_discipline      ON users(disciplineid);
CREATE INDEX IF NOT EXISTS idx_users_verified_until  ON users(verified_until);

CREATE TRIGGER IF NOT EXISTS trg_users_update
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
  UPDATE users
     SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
   WHERE id = OLD.id;
END;

CREATE TABLE IF NOT EXISTS posts (
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

CREATE INDEX IF NOT EXISTS idx_posts_user       ON posts(userid);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);

CREATE TRIGGER IF NOT EXISTS trg_posts_update
AFTER UPDATE ON posts
FOR EACH ROW
BEGIN
  UPDATE posts
     SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
   WHERE id = OLD.id;
END;

CREATE TABLE IF NOT EXISTS comments (
  id         TEXT PRIMARY KEY,
  postid     TEXT NOT NULL
               REFERENCES posts(id)
               ON DELETE CASCADE ON UPDATE CASCADE,
  userid     TEXT NOT NULL
               REFERENCES users(id)
               ON DELETE RESTRICT ON UPDATE CASCADE,
  body       TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
) STRICT;

CREATE INDEX IF NOT EXISTS idx_comments_post        ON comments(postid);
CREATE INDEX IF NOT EXISTS idx_comments_user        ON comments(userid);
CREATE INDEX IF NOT EXISTS idx_comments_created_at  ON comments(created_at DESC);

CREATE TRIGGER IF NOT EXISTS trg_comments_update
AFTER UPDATE ON comments
FOR EACH ROW
BEGIN
  UPDATE comments
     SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
   WHERE id = OLD.id;
END;

CREATE TABLE IF NOT EXISTS exams (
  id           TEXT PRIMARY KEY,
  userid       TEXT NOT NULL
                 REFERENCES users(id)
                 ON DELETE RESTRICT ON UPDATE CASCADE,
  disciplineid INTEGER NOT NULL
                 REFERENCES disciplines(id)
                 ON DELETE RESTRICT ON UPDATE CASCADE,
  campusid     INTEGER
                 REFERENCES campuses(id)
                 ON DELETE RESTRICT ON UPDATE CASCADE,
  exam_date    TEXT NOT NULL,
  uploaded_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  accesskey    TEXT NOT NULL UNIQUE,
  mime_type    TEXT NOT NULL CHECK (mime_type IN ('application/pdf')),
  nbytes       INTEGER NOT NULL,
  checksum     TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_exams_date ON exams(disciplineid, exam_date DESC);
CREATE INDEX IF NOT EXISTS idx_exams_user ON exams(userid);

CREATE TRIGGER IF NOT EXISTS trg_exams_set_update
AFTER UPDATE ON exams
FOR EACH ROW
BEGIN
  UPDATE exams
     SET uploaded_at = uploaded_at,
         accesskey   = accesskey
   WHERE id = OLD.id;
END;

CREATE TABLE IF NOT EXISTS sessions (
  id         TEXT PRIMARY KEY,
  userid     TEXT NOT NULL
               REFERENCES users(id)
               ON DELETE CASCADE ON UPDATE CASCADE,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  last_seen  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  expires_at TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_sessions_user        ON sessions(userid);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at  ON sessions(expires_at);

