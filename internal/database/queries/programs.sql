-- name: ListProgramsWithVersions :many
SELECT p.id, p.name, pv.name as version
FROM programs p
JOIN program_versions pv ON p.id = pv.program_id
WHERE p.deleted_at IS NULL
ORDER BY p.name, pv.name DESC;

-- name: ListPrograms :many
SELECT * FROM programs WHERE deleted_at IS NULL ORDER BY name, id;

-- name: GetProgram :one
SELECT * FROM programs WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: ListProgramVersions :many
SELECT pv.* FROM program_versions pv
JOIN programs p ON p.id = pv.program_id AND p.deleted_at IS NULL
ORDER BY pv.name DESC, pv.program_id;

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
