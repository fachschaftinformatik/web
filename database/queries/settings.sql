-- name: GetConfig :one
SELECT * FROM global_config WHERE id = 1;

-- name: UpdateOfficeOccupied :exec
UPDATE global_config SET office_occupied = ? WHERE id = 1;
