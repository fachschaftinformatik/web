-- name: CreateSession :one
INSERT INTO sessions (id, user_id, expires_at, user_agent, ip_address)
VALUES (sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(expires_at), sqlc.arg(user_agent), sqlc.arg(ip_address))
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: TouchSession :one
UPDATE sessions
SET last_seen = strftime('%Y-%m-%dT%H:%M:%SZ','now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SlideSession :one
UPDATE sessions
SET last_seen = strftime('%Y-%m-%dT%H:%M:%SZ','now'),
    expires_at = sqlc.arg(expires_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = sqlc.arg(id);

-- name: DeleteUserSessions :exec
DELETE FROM sessions WHERE user_id = sqlc.arg(user_id);

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < strftime('%Y-%m-%dT%H:%M:%SZ','now');
