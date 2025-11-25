-- name: ListModulesByProgram :many
SELECT * FROM modules 
WHERE programid = sqlc.arg(programid) 
ORDER BY name;

-- name: CreateModule :one
INSERT INTO modules (programid, name) 
VALUES (sqlc.arg(programid), sqlc.arg(name)) 
RETURNING *;
