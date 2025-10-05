-- name: CreateUser :one
INSERT INTO users (
  id, email, name, password, role, active, verified, campusid, disciplineid
) VALUES (
  sqlc.arg(id), sqlc.arg(email), sqlc.arg(name), sqlc.arg(password),
  COALESCE(sqlc.arg(role), 'user'),
  COALESCE(sqlc.arg(active), 0),
  0,
  sqlc.arg(campusid), sqlc.arg(disciplineid)
)
RETURNING
  id, email, name, role, active, verified, verified_at, verified_until,
  campusid, disciplineid, created_at, updated_at;

-- name: GetUser :one
SELECT
  id, email, name, role, active, verified, verified_at, verified_until,
  campusid, disciplineid, created_at, updated_at
FROM users
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetUserByEmail :one
SELECT
  id, email, name, role, active, verified, verified_at, verified_until,
  campusid, disciplineid, created_at, updated_at
FROM users
WHERE lower(email) = lower(sqlc.arg(email))
LIMIT 1;

-- name: SetUserActive :one
UPDATE users
SET active = sqlc.arg(active)
WHERE id = sqlc.arg(id)
RETURNING
  id, email, name, role, active, verified, verified_at, verified_until,
  campusid, disciplineid, created_at, updated_at;

-- name: SetUserRole :one
UPDATE users
SET role = sqlc.arg(role)
WHERE id = sqlc.arg(id)
RETURNING
  id, email, name, role, active, verified, verified_at, verified_until,
  campusid, disciplineid, created_at, updated_at;

-- name: VerifyUser :one
UPDATE users
SET verified = 1,
    verified_at = strftime('%Y-%m-%dT%H:%M:%fZ','now'),
    verified_until = sqlc.arg(verified_until)
WHERE id = sqlc.arg(id)
RETURNING
  id, email, name, role, active, verified, verified_at, verified_until,
  campusid, disciplineid, created_at, updated_at;

-- name: UnverifyUser :one
UPDATE users
SET verified = 0
WHERE id = sqlc.arg(id)
RETURNING
  id, email, name, role, active, verified, verified_at, verified_until,
  campusid, disciplineid, created_at, updated_at;

-- name: UpdateUserVerificationWindow :one
UPDATE users
SET verified_until = sqlc.arg(verified_until)
WHERE id = sqlc.arg(id)
RETURNING
  id, email, name, role, active, verified, verified_at, verified_until,
  campusid, disciplineid, created_at, updated_at;

-- name: SweepExpiredVerifications :exec
UPDATE users
SET verified = 0
WHERE verified = 1
  AND verified_until IS NOT NULL
  AND verified_until < strftime('%Y-%m-%dT%H:%M:%fZ','now');

-- name: ListUsers :many
SELECT
  id, email, name, role, active, verified, verified_at, verified_until,
  campusid, disciplineid, created_at, updated_at
FROM users
ORDER BY created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

