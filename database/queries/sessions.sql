-- name: CreateSession :one
INSERT INTO sessions (id, userid, expires_at)
VALUES (sqlc.arg(id), sqlc.arg(userid), sqlc.arg(expires_at))
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: TouchSession :one
UPDATE sessions
SET last_seen = strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SlideSession :one
UPDATE sessions
SET last_seen = strftime('%Y-%m-%dT%H:%M:%fZ','now'),
    expires_at = sqlc.arg(expires_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = sqlc.arg(id);

-- name: DeleteUserSessions :exec
DELETE FROM sessions WHERE userid = sqlc.arg(userid);

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < strftime('%Y-%m-%dT%H:%M:%fZ','now');

