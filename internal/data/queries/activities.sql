-- name: CreateActivity :one
INSERT INTO activities (
    id, user_id, type, target_id, target_name
) VALUES (
    sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(type), sqlc.arg(target_id), sqlc.arg(target_name)
)
RETURNING *;

-- name: ListUserActivities :many
SELECT a.*, CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as user_name
FROM activities a
JOIN users u ON a.user_id = u.id
WHERE a.user_id = sqlc.arg(user_id) 
  AND a.deleted_at IS NULL
  AND a.created_at >= date('now', '-30 days')
ORDER BY a.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: CountUserActivities :one
SELECT COUNT(*)
FROM activities a
WHERE a.user_id = sqlc.arg(user_id) 
  AND a.deleted_at IS NULL
  AND a.created_at >= date('now', '-30 days');

-- name: ListAllActivities :many
SELECT a.*, CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as user_name
FROM activities a
JOIN users u ON a.user_id = u.id
WHERE a.deleted_at IS NULL
  AND a.created_at >= date('now', '-30 days')
ORDER BY a.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);
