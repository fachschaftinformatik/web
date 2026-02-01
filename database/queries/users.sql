-- name: CreateUser :one
INSERT INTO users (
  id, email, name, password, role, active, verified, program_id, verification_token, avatar_url
) VALUES (
  sqlc.arg(id), sqlc.arg(email), sqlc.arg(name), sqlc.arg(password),
  COALESCE(sqlc.arg(role), 'user'),
  COALESCE(sqlc.arg(active), 0),
  0,
  sqlc.arg(program_id),
  sqlc.arg(verification_token),
  sqlc.arg(avatar_url)
)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE lower(email) = lower(sqlc.arg(email)) AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByVerificationToken :one
SELECT *
FROM users
WHERE verification_token = sqlc.arg(verification_token) AND deleted_at IS NULL
LIMIT 1;

-- name: SetUserActive :one
UPDATE users
SET active = sqlc.arg(active)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: SetUserRole :one
UPDATE users
SET role = sqlc.arg(role)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: VerifyUser :one
UPDATE users
SET verified = 1,
    verified_at = strftime('%Y-%m-%dT%H:%M:%SZ','now'),
    verification_token = NULL
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserToken :exec
UPDATE users
SET verification_token = sqlc.arg(verification_token)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: UnverifyUser :one
UPDATE users
SET verified = 0
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: SweepExpiredVerifications :exec
UPDATE users
SET verified = 0
WHERE verified = 1
  AND verified_at < strftime('%Y-%m-%dT%H:%M:%SZ','now', '-1 year'); -- Example logic

-- name: UpdateUser :one
UPDATE users
SET name = sqlc.arg(name),
    program_id = sqlc.arg(program_id),
    theme = sqlc.arg(theme),
    private = sqlc.arg(private),
    avatar_url = sqlc.arg(avatar_url)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserAvatar :one
UPDATE users
SET avatar_url = sqlc.arg(avatar_url)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: ListUsers :many
SELECT *
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: SearchUsers :many
SELECT *
FROM users
WHERE (
  lower(id) LIKE '%' || lower(sqlc.arg(query)) || '%' OR
  lower(name) LIKE '%' || lower(sqlc.arg(query)) || '%' OR
  lower(email) LIKE '%' || lower(sqlc.arg(query)) || '%'
) AND active = 1 AND private = 0 AND deleted_at IS NULL
ORDER BY name ASC
LIMIT 10;
-- name: SetPasswordResetToken :exec
UPDATE users
SET password_reset_token = sqlc.arg(password_reset_token),
    password_reset_expires = sqlc.arg(password_reset_expires)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: GetUserByPasswordResetToken :one
SELECT *
FROM users
WHERE password_reset_token = sqlc.arg(token) AND password_reset_expires > strftime('%Y-%m-%dT%H:%M:%SZ','now') AND deleted_at IS NULL
LIMIT 1;

-- name: ClearPasswordResetToken :exec
UPDATE users
SET password_reset_token = NULL,
    password_reset_expires = NULL
WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: UpdateUserPassword :one
UPDATE users
SET password = sqlc.arg(password),
    password_reset_token = NULL,
    password_reset_expires = NULL
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;