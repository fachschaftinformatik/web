-- +goose Up
ALTER TABLE media ADD COLUMN gallery_visible INTEGER NOT NULL DEFAULT 1 CHECK (gallery_visible IN (0,1));
CREATE INDEX media_gallery_photos ON media(event_id,created_at DESC,id DESC)
WHERE deleted_at IS NULL AND gallery_visible=1;

-- +goose Down
DROP INDEX media_gallery_photos;
ALTER TABLE media DROP COLUMN gallery_visible;
