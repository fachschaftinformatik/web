-- name: ListProgramsWithVersions :many
SELECT p.id, p.name, pv.name as version
FROM programs p
JOIN program_versions pv ON p.id = pv.program_id
WHERE p.deleted_at IS NULL
ORDER BY p.name, pv.name DESC;

-- name: GetProgramWithVersions :many
SELECT p.id, p.name, pv.name as version
FROM programs p
JOIN program_versions pv ON p.id = pv.program_id
WHERE p.id = sqlc.arg(id) AND p.deleted_at IS NULL
ORDER BY pv.name DESC;

-- name: CreateProgram :one
INSERT INTO programs (id, name) VALUES (sqlc.arg(id), sqlc.arg(name)) RETURNING *;

-- name: CreateProgramVersion :exec
INSERT INTO program_versions (program_id, name) VALUES (sqlc.arg(program_id), sqlc.arg(name));
