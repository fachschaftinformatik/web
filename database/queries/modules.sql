-- name: ListModulesByProgram :many
SELECT * FROM modules 
WHERE programid = sqlc.arg(programid) 
ORDER BY name;

-- name: CreateModule :one
INSERT INTO modules (programid, name, alias) 
VALUES (sqlc.arg(programid), sqlc.arg(name), sqlc.arg(alias)) 
RETURNING *;
-- name: GetModule :one
SELECT * FROM modules WHERE id = sqlc.arg(id) LIMIT 1;
