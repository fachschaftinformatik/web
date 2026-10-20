-- +goose Up
DROP INDEX media_gallery_photos;
ALTER TABLE media RENAME COLUMN gallery_visible TO homepage_visible;
CREATE INDEX media_homepage_photos ON media(created_at DESC,id DESC)
WHERE deleted_at IS NULL AND homepage_visible=1;

-- +goose Down
DROP INDEX media_homepage_photos;
ALTER TABLE media RENAME COLUMN homepage_visible TO gallery_visible;
CREATE INDEX media_gallery_photos ON media(event_id,created_at DESC,id DESC)
WHERE deleted_at IS NULL AND gallery_visible=1;
