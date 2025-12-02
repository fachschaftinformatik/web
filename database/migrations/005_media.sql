-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
  id INTEGER PRIMARY KEY,
  title TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
) STRICT;

CREATE TABLE media (
  id TEXT PRIMARY KEY,
  event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  userid TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  title TEXT,
  description TEXT,
  accesskey TEXT NOT NULL UNIQUE,
  mime_type TEXT NOT NULL,
  nbytes INTEGER NOT NULL,
  uploaded_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
) STRICT;

CREATE INDEX idx_media_event ON media(event_id, uploaded_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_media_event;
DROP TABLE media;
DROP TABLE events;
-- +goose StatementEnd
