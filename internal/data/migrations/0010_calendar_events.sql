-- +goose Up
CREATE TABLE calendar_events (
  id INTEGER PRIMARY KEY,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  body_format TEXT NOT NULL DEFAULT 'plain' CHECK(body_format IN ('plain','html')),
  event_date TEXT NOT NULL DEFAULT '',
  location TEXT NOT NULL DEFAULT '',
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  cover_key TEXT NOT NULL DEFAULT '',
  cover_alt TEXT NOT NULL DEFAULT '',
  cover_width INTEGER NOT NULL DEFAULT 0,
  cover_height INTEGER NOT NULL DEFAULT 0,
  legacy_post_id INTEGER UNIQUE REFERENCES discussion_posts(id) ON DELETE RESTRICT,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
  deleted_at TEXT
) STRICT;
INSERT INTO calendar_events(id,title,body,body_format,event_date,location,user_id,legacy_post_id,created_at,updated_at,deleted_at)
SELECT id,title,body,body_format,COALESCE(event_date,''),COALESCE(location,''),user_id,id,created_at,updated_at,deleted_at FROM discussion_posts WHERE type='event';

-- +goose Down
DROP TABLE calendar_events;
