-- name: GetSetting :one
SELECT value FROM settings WHERE key = ?;

-- name: UpdateSetting :exec
UPDATE settings SET value = ? WHERE key = ?;
