-- +goose Up
-- Replace the column to change its SQLite default while retaining selections.
DROP INDEX media_gallery_photos;
ALTER TABLE media ADD COLUMN gallery_selection INTEGER NOT NULL DEFAULT 0 CHECK (gallery_selection IN (0,1));
UPDATE media SET gallery_selection=gallery_visible;
ALTER TABLE media DROP COLUMN gallery_visible;
ALTER TABLE media RENAME COLUMN gallery_selection TO gallery_visible;
CREATE INDEX media_gallery_photos ON media(event_id,created_at DESC,id DESC)
WHERE deleted_at IS NULL AND gallery_visible=1;

-- +goose Down
DROP INDEX media_gallery_photos;
ALTER TABLE media ADD COLUMN gallery_selection INTEGER NOT NULL DEFAULT 1 CHECK (gallery_selection IN (0,1));
UPDATE media SET gallery_selection=gallery_visible;
ALTER TABLE media DROP COLUMN gallery_visible;
ALTER TABLE media RENAME COLUMN gallery_selection TO gallery_visible;
CREATE INDEX media_gallery_photos ON media(event_id,created_at DESC,id DESC)
WHERE deleted_at IS NULL AND gallery_visible=1;
