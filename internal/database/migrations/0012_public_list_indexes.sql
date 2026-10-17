-- +goose Up
CREATE INDEX media_album_order ON media(event_id,created_at DESC,id DESC) WHERE deleted_at IS NULL;
CREATE INDEX discussion_posts_type_order ON discussion_posts(type,pinned DESC,created_at DESC,id DESC) WHERE deleted_at IS NULL;
CREATE INDEX discussion_posts_author_order ON discussion_posts(user_id,pinned DESC,created_at DESC,id DESC) WHERE deleted_at IS NULL;
CREATE INDEX discussion_comments_post_order ON discussion_comments(post_id,created_at DESC,id DESC) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX discussion_comments_post_order;
DROP INDEX discussion_posts_author_order;
DROP INDEX discussion_posts_type_order;
DROP INDEX media_album_order;
