-- +goose Up
-- +goose StatementBegin
CREATE TABLE notifications (
  id         TEXT PRIMARY KEY,
  userid     TEXT NOT NULL
               REFERENCES users(id)
               ON DELETE CASCADE ON UPDATE CASCADE,
  title      TEXT NOT NULL,
  message    TEXT NOT NULL,
  type       TEXT NOT NULL CHECK (type IN ('exam','forum','news')),
  link       TEXT NOT NULL,
  read       INTEGER NOT NULL DEFAULT 0 CHECK (read IN (0,1)),
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
) STRICT;

CREATE INDEX idx_notifications_user       ON notifications(userid);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE notifications;
-- +goose StatementEnd
