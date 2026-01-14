-- name: CreateUser :one
INSERT INTO users (
  id, email, name, password, role, active, verified, programid, verification_token, avatar_url
) VALUES (
  sqlc.arg(id), sqlc.arg(email), sqlc.arg(name), sqlc.arg(password),
  COALESCE(sqlc.arg(role), 'user'),
  COALESCE(sqlc.arg(active), 0),
  0,
  sqlc.arg(programid),
  sqlc.arg(verification_token),
  sqlc.arg(avatar_url)
)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE lower(email) = lower(sqlc.arg(email))
LIMIT 1;

-- name: GetUserByVerificationToken :one
SELECT *
FROM users
WHERE verification_token = sqlc.arg(verification_token)
LIMIT 1;

-- name: SetUserActive :one
UPDATE users
SET active = sqlc.arg(active)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetUserRole :one
UPDATE users
SET role = sqlc.arg(role)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: VerifyUser :one
UPDATE users
SET verified = 1,
    verified_at = strftime('%Y-%m-%dT%H:%M:%fZ','now'),
    verified_until = sqlc.arg(verified_until),
    verification_token = NULL
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateUserToken :exec
UPDATE users
SET verification_token = sqlc.arg(verification_token)
WHERE id = sqlc.arg(id);

-- name: UnverifyUser :one
UPDATE users
SET verified = 0
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateUserVerificationWindow :one
UPDATE users
SET verified_until = sqlc.arg(verified_until)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SweepExpiredVerifications :exec
UPDATE users
SET verified = 0
WHERE verified = 1
  AND verified_until IS NOT NULL
  AND verified_until < strftime('%Y-%m-%dT%H:%M:%fZ','now');

-- name: UpdateUser :one
UPDATE users
SET name = sqlc.arg(name),
    programid = sqlc.arg(programid),
    theme = sqlc.arg(theme),
    private = sqlc.arg(private),
    avatar_url = sqlc.arg(avatar_url),
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateUserAvatar :one
UPDATE users
SET avatar_url = sqlc.arg(avatar_url),
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: SearchUsers :many
SELECT *
FROM users
WHERE (
  lower(id) LIKE '%' || lower(sqlc.arg(query)) || '%' OR
  lower(name) LIKE '%' || lower(sqlc.arg(query)) || '%' OR
  lower(email) LIKE '%' || lower(sqlc.arg(query)) || '%'
) AND active = 1 AND private = 0
ORDER BY name ASC
LIMIT 10;
