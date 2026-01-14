-- name: CreateEvent :one
INSERT INTO events (title, cover_path) VALUES (?, ?) RETURNING *;

-- name: ListEvents :many
SELECT * FROM events ORDER BY created_at DESC;

-- name: GetEvent :one
SELECT * FROM events WHERE id = ? LIMIT 1;

-- name: CreateMedia :one
INSERT INTO media (
  id, event_id, userid, title, description, 
  accesskey, mime_type, nbytes
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: ListMediaByEvent :many
SELECT m.*, u.name as uploader_name 
FROM media m
JOIN users u ON m.userid = u.id
WHERE event_id = ?
ORDER BY uploaded_at DESC;

-- name: GetMedia :one
SELECT m.*, u.name as uploader_name 
FROM media m
JOIN users u ON m.userid = u.id
WHERE m.id = ? LIMIT 1;

-- name: DeleteMedia :exec
DELETE FROM media WHERE id = ?;
