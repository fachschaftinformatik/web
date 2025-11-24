-- name: ListProgramsWithVersions :many
SELECT p.id, p.name, pv.name as version
FROM programs p
JOIN program_versions pv ON p.id = pv.programid
ORDER BY p.name, pv.name DESC;

-- name: GetProgramWithVersions :many
SELECT p.id, p.name, pv.name as version
FROM programs p
JOIN program_versions pv ON p.id = pv.programid
WHERE p.id = sqlc.arg(id)
ORDER BY pv.name DESC;
