-- name: ListModulesByProgram :many
SELECT * FROM modules 
WHERE program_id = sqlc.arg(program_id) AND deleted_at IS NULL
ORDER BY name;

-- name: CreateModule :one
INSERT INTO modules (id, program_id, name, alias) 
VALUES (sqlc.arg(id), sqlc.arg(program_id), sqlc.arg(name), sqlc.arg(alias)) 
RETURNING *;

-- name: GetModule :one
SELECT * FROM modules WHERE id = sqlc.arg(id) AND deleted_at IS NULL LIMIT 1;

-- name: SearchModules :many
SELECT * FROM modules
WHERE (name LIKE '%' || sqlc.arg('query') || '%'
   OR alias LIKE '%' || sqlc.arg('query') || '%')
   AND deleted_at IS NULL
LIMIT 20;
