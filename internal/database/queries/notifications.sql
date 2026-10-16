-- name: CreateNotification :one
INSERT INTO notifications (
  id, user_id, title, message, type, link
) VALUES (
  sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(title), sqlc.arg(message), sqlc.arg(type), sqlc.arg(link)
) RETURNING *;

-- name: ListNotifications :many
SELECT * FROM notifications
WHERE user_id = sqlc.arg(user_id) AND read = 0 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 50;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications
WHERE user_id = sqlc.arg(user_id) AND read = 0 AND deleted_at IS NULL;

-- name: MarkNotificationAsRead :one
UPDATE notifications
SET read = 1
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) AND deleted_at IS NULL
RETURNING *;

-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET read = 1
WHERE user_id = sqlc.arg(user_id) AND deleted_at IS NULL;

-- name: ListUsersForNotification :many
SELECT id FROM users WHERE active = 1 AND deleted_at IS NULL;
