-- name: ListProgramsWithVersions :many
SELECT p.id, p.name, pv.name as version
FROM programs p
JOIN program_versions pv ON p.id = pv.programid
ORDER BY p.name, pv.name DESC;
