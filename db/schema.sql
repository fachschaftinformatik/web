PRAGMA user_version = 2;

PRAGMA foreign_keys = ON;
PRAGMA recursive_triggers = OFF;
PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS campuses (
  id       INTEGER PRIMARY KEY,
  name     TEXT NOT NULL UNIQUE,
  location TEXT,
  created  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
) STRICT;

CREATE TABLE IF NOT EXISTS disciplines (
  id      INTEGER PRIMARY KEY,
  name    TEXT NOT NULL,
  degree  TEXT NOT NULL,
  created TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE(degree, name)
) STRICT;

CREATE TABLE IF NOT EXISTS users (
  id           TEXT PRIMARY KEY,
  email        TEXT NOT NULL,
  password     TEXT NOT NULL,
  name         TEXT NOT NULL,
  role         TEXT NOT NULL CHECK (role IN ('user','editor','admin')),
  active       INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0,1)),
  campusid     INTEGER NOT NULL
                 REFERENCES campuses(id)
                 ON DELETE RESTRICT ON UPDATE CASCADE,
  disciplineid INTEGER NOT NULL
                 REFERENCES disciplines(id)
                 ON DELETE RESTRICT ON UPDATE CASCADE,
  created      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
) STRICT;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users(lower(email));
CREATE INDEX IF NOT EXISTS idx_users_campus          ON users(campusid);
CREATE INDEX IF NOT EXISTS idx_users_discipline      ON users(disciplineid);

CREATE TRIGGER IF NOT EXISTS trg_users_update
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
  UPDATE users
     SET updated = strftime('%Y-%m-%dT%H:%M:%fZ','now')
   WHERE id = OLD.id;
END;

CREATE TABLE IF NOT EXISTS posts (
  id      TEXT PRIMARY KEY,
  userid  TEXT NOT NULL
            REFERENCES users(id)
            ON DELETE RESTRICT ON UPDATE CASCADE,
  title   TEXT NOT NULL,
  body    TEXT NOT NULL,
  created TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  deleted TEXT
) STRICT;

CREATE INDEX IF NOT EXISTS idx_forum_posts_user     ON forum_posts(userid);
CREATE INDEX IF NOT EXISTS idx_forum_posts_created  ON forum_posts(created DESC);

CREATE TRIGGER IF NOT EXISTS trg_posts_update
AFTER UPDATE ON posts
FOR EACH ROW
BEGIN
  UPDATE posts
     SET updated = strftime('%Y-%m-%dT%H:%M:%fZ','now')
   WHERE id = OLD.id;
END;

CREATE TABLE IF NOT EXISTS comments (
  id      TEXT PRIMARY KEY,
  postid  TEXT NOT NULL
            REFERENCES posts(id)
            ON DELETE CASCADE ON UPDATE CASCADE,
  userid  TEXT NOT NULL
            REFERENCES users(id)
            ON DELETE RESTRICT ON UPDATE CASCADE,
  body    TEXT NOT NULL,
  created TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  deleted TEXT
) STRICT;

CREATE INDEX IF NOT EXISTS idx_forum_comments_post    ON forum_comments(postid);
CREATE INDEX IF NOT EXISTS idx_forum_comments_user    ON forum_comments(userid);
CREATE INDEX IF NOT EXISTS idx_forum_comments_created ON forum_comments(created DESC);

CREATE TRIGGER IF NOT EXISTS trg_comments_update
AFTER UPDATE ON comments
FOR EACH ROW
BEGIN
  UPDATE comments
     SET updated = strftime('%Y-%m-%dT%H:%M:%fZ','now')
   WHERE id = OLD.id;
END;

CREATE TABLE IF NOT EXISTS exams (
  id               TEXT PRIMARY KEY,
  userid           TEXT NOT NULL
                    REFERENCES users(id)
                    ON DELETE RESTRICT ON UPDATE CASCADE,
  disciplineid     INTEGER NOT NULL
                     REFERENCES disciplines(id)
                     ON DELETE RESTRICT ON UPDATE CASCADE,
  campusid         INTEGER
                     REFERENCES campuses(id)
                     ON DELETE RESTRICT ON UPDATE CASCADE,
  examdate         TEXT NOT NULL,
  uploaded         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  accesskey        TEXT NOT NULL UNIQUE,
  mimetype         TEXT NOT NULL CHECK (mimetype IN ('application/pdf')),
  sz               INTEGER,
  checksum         TEXT
) STRICT;

CREATE INDEX IF NOT EXISTS idx_exams_date ON exams(disciplineid, examdate DESC);
CREATE INDEX IF NOT EXISTS idx_exams_user ON exams(userid);

CREATE TRIGGER IF NOT EXISTS trg_exams_set_update
AFTER UPDATE ON exams
FOR EACH ROW
BEGIN
  UPDATE exams
     SET uploaded = uploaded,
         accesskey = accesskey
   WHERE id = OLD.id;
END;

CREATE TABLE IF NOT EXISTS sessions (
  id       TEXT PRIMARY KEY,
  userid   TEXT NOT NULL
             REFERENCES users(id)
             ON DELETE CASCADE ON UPDATE CASCADE,
  created  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  lastseen TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  expires  TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_sessions_user    ON sessions(userid);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires);

CREATE INDEX IF NOT EXISTS idx_posts_user_fk       ON posts(userid);
CREATE INDEX IF NOT EXISTS idx_comments_post_fk    ON comments(postid);
CREATE INDEX IF NOT EXISTS idx_comments_user_fk    ON comments(userid);
CREATE INDEX IF NOT EXISTS idx_exams_discipline_fk ON exams(disciplineid);
CREATE INDEX IF NOT EXISTS idx_exams_campus_fk     ON exams(campusid);

