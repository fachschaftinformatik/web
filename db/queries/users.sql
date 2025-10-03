-- name: CreateUser :one
INSERT INTO users (
  id, email, password, name, role, active, campusid, disciplineid
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING
  id, email, name, role, active, campusid, disciplineid, created, updated;

-- name: GetUser :one
SELECT
  id, email, name, role, active, campusid, disciplineid, created, updated
FROM users
WHERE id = ?
LIMIT 1;

-- name: ListUsers :many
SELECT
  id, email, name, role, active, campusid, disciplineid, created, updated
FROM users
ORDER BY created DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);
