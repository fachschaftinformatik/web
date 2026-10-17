-- +goose Up
ALTER TABLE calendar_events ADD COLUMN source_url TEXT NOT NULL DEFAULT '';
ALTER TABLE calendar_events ADD COLUMN source_published_at TEXT NOT NULL DEFAULT '';
ALTER TABLE calendar_events ADD COLUMN source_status TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX calendar_events_source ON calendar_events(source_url) WHERE source_url != '';
ALTER TABLE calendar_events ADD COLUMN gallery_album_id INTEGER REFERENCES events(id) ON DELETE RESTRICT;
ALTER TABLE calendar_events ADD COLUMN gallery_cover_id INTEGER REFERENCES media(id) ON DELETE RESTRICT;

-- +goose Down
ALTER TABLE calendar_events DROP COLUMN gallery_cover_id;
ALTER TABLE calendar_events DROP COLUMN gallery_album_id;
DROP INDEX calendar_events_source;
ALTER TABLE calendar_events DROP COLUMN source_status;
ALTER TABLE calendar_events DROP COLUMN source_published_at;
ALTER TABLE calendar_events DROP COLUMN source_url;
