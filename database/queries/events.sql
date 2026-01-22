-- name: CreateEvent :one
INSERT INTO events (id, title, cover_path) VALUES (sqlc.arg(id), sqlc.arg(title), sqlc.arg(cover_path)) RETURNING *;

-- name: ListEvents :many
SELECT * FROM events WHERE deleted_at IS NULL ORDER BY created_at DESC;

-- name: GetEvent :one
SELECT * FROM events WHERE id = sqlc.arg(id) AND deleted_at IS NULL LIMIT 1;

-- name: CreateMedia :one
INSERT INTO media (
  id, event_id, user_id, title, description, 
  access_key, mime_type, nbytes
) VALUES (
  sqlc.arg(id), sqlc.arg(event_id), sqlc.arg(user_id), sqlc.arg(title), 
  sqlc.arg(description), sqlc.arg(access_key), sqlc.arg(mime_type), sqlc.arg(nbytes)
) RETURNING *;

-- name: ListMediaByEvent :many
SELECT m.*, CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as uploader_name 
FROM media m
JOIN users u ON m.user_id = u.id
WHERE event_id = sqlc.arg(event_id) AND m.deleted_at IS NULL
ORDER BY m.created_at DESC;

-- name: GetMedia :one
SELECT m.*, CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as uploader_name 
FROM media m
JOIN users u ON m.user_id = u.id
WHERE m.id = sqlc.arg(id) AND m.deleted_at IS NULL LIMIT 1;

-- name: DeleteMedia :exec
UPDATE media SET deleted_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = sqlc.arg(id);
