-- name: CreateNotification :one
INSERT INTO notifications (
  id, userid, title, message, type, link
) VALUES (
  sqlc.arg(id), sqlc.arg(userid), sqlc.arg(title), sqlc.arg(message), sqlc.arg(type), sqlc.arg(link)
) RETURNING *;

-- name: ListNotifications :many
SELECT * FROM notifications
WHERE userid = sqlc.arg(userid) AND read = 0
ORDER BY created_at DESC
LIMIT 50;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications
WHERE userid = sqlc.arg(userid) AND read = 0;

-- name: MarkNotificationAsRead :one
UPDATE notifications
SET read = 1
WHERE id = sqlc.arg(id) AND userid = sqlc.arg(userid)
RETURNING *;

-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET read = 1
WHERE userid = sqlc.arg(userid);



-- name: ListUsersForNotification :many
SELECT id FROM users WHERE active = 1;
