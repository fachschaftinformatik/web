-- name: CreateArchiveEntry :one
INSERT INTO archive (
  id, group_id, user_id, module_id, version, exam_date,
  comment, edit_version, is_latest, access_key,
  mime_type, nbytes, checksum
) VALUES (
  sqlc.arg(id), sqlc.arg(group_id), sqlc.arg(user_id), sqlc.arg(module_id), sqlc.arg(version), sqlc.arg(exam_date),
  sqlc.arg(comment), sqlc.arg(edit_version), sqlc.arg(is_latest), sqlc.arg(access_key),
  sqlc.arg(mime_type), sqlc.arg(nbytes), sqlc.arg(checksum)
) RETURNING *;

-- name: ListArchiveEntries :many
SELECT a.*,
       m.name as module_name, m.program_id,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as uploader_name
FROM archive a
JOIN modules m ON a.module_id = m.id
JOIN users u ON a.user_id = u.id
WHERE (sqlc.narg('program_id') IS NULL OR m.program_id = sqlc.narg('program_id'))
  AND (sqlc.narg('version') IS NULL OR a.version = sqlc.narg('version'))
  AND (sqlc.narg('module_id') IS NULL OR a.module_id = sqlc.narg('module_id'))
  AND a.is_latest = 1
  AND a.deleted_at IS NULL
ORDER BY a.exam_date DESC;

-- name: GetArchiveEntry :one
SELECT * FROM archive WHERE id = sqlc.arg(id) AND deleted_at IS NULL LIMIT 1;

-- name: ListArchiveCatalog :many
SELECT a.id, b.module_id, b.version, a.exam_date, a.comment, a.nbytes,
       CAST((SELECT original.user_id FROM archive original WHERE original.group_id=a.group_id ORDER BY original.edit_version, original.id LIMIT 1) AS INTEGER) AS owner_id,
       m.name AS module_name, m.alias, COALESCE(sp.name,p.name) AS program_name, COALESCE(pq.degree,'') AS degree
FROM archive a
JOIN archive_placements b ON b.archive_id=a.id
JOIN modules m ON m.id = b.module_id AND m.deleted_at IS NULL
JOIN programs p ON p.id = m.program_id AND p.deleted_at IS NULL
LEFT JOIN program_qualifications pq ON pq.legacy_program_id=p.id
LEFT JOIN study_programs sp ON sp.id=pq.study_program_id
WHERE a.is_latest = 1 AND a.deleted_at IS NULL AND a.mime_type = 'application/pdf'
  AND NOT EXISTS (SELECT 1 FROM archive_entry_aliases x WHERE x.legacy_id=a.id AND x.canonical_id!=a.id)
  AND (a.request_id IS NULL OR EXISTS (SELECT 1 FROM archive_requests r WHERE r.id=a.request_id AND r.status='approved'))
ORDER BY a.exam_date DESC, a.created_at DESC, a.id DESC;

-- name: ListArchiveModules :many
SELECT m.id, m.name, m.program_id, p.name AS program_name,COALESCE(sp.name,p.name) AS study_program,COALESCE(pq.degree,'') AS degree
FROM modules m
JOIN programs p ON p.id = m.program_id AND p.deleted_at IS NULL
LEFT JOIN program_qualifications pq ON pq.legacy_program_id=p.id
LEFT JOIN study_programs sp ON sp.id=pq.study_program_id
WHERE m.deleted_at IS NULL
ORDER BY m.name, m.id;

-- name: GetArchiveEntryDetails :one
SELECT a.*,
       m.name as module_name, m.program_id,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as uploader_name
FROM archive a
JOIN modules m ON a.module_id = m.id
JOIN users u ON a.user_id = u.id
WHERE a.id = sqlc.arg(id) AND a.deleted_at IS NULL LIMIT 1;

-- name: GetLatestByGroupId :one
SELECT * FROM archive WHERE group_id = sqlc.arg(group_id) AND is_latest = 1 AND deleted_at IS NULL LIMIT 1;

-- name: ClearLatestFlagByGroupId :exec
UPDATE archive SET is_latest = 0 WHERE group_id = sqlc.arg(group_id);

-- name: DeleteArchiveEntry :exec
UPDATE archive SET deleted_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now'), is_latest = 0 WHERE id = sqlc.arg(id);

-- name: PromoteLatestInGroup :exec
UPDATE archive SET is_latest = 1 
WHERE id = (
    SELECT a2.id FROM archive a2
    WHERE a2.group_id = sqlc.arg(group_id) AND a2.deleted_at IS NULL 
    ORDER BY a2.edit_version DESC LIMIT 1
);

-- name: ListArchiveVersions :many
SELECT a.*,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as uploader_name
FROM archive a
JOIN users u ON a.user_id = u.id
WHERE a.group_id = sqlc.arg(group_id) AND a.deleted_at IS NULL
ORDER BY a.edit_version DESC;

-- name: SearchArchive :many
SELECT a.*,
       m.name as module_name, m.program_id,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as uploader_name
FROM archive a
JOIN modules m ON a.module_id = m.id
JOIN users u ON a.user_id = u.id
WHERE a.is_latest = 1 AND a.deleted_at IS NULL
  AND (m.name LIKE '%' || sqlc.arg('query') || '%'
       OR m.alias LIKE '%' || sqlc.arg('query') || '%'
       OR a.comment LIKE '%' || sqlc.arg('query') || '%')
ORDER BY a.exam_date DESC
LIMIT 20;

-- name: GetArchiveEntryByMetadata :one
SELECT * FROM archive 
WHERE module_id = sqlc.arg(module_id) 
  AND version = sqlc.arg(version) 
  AND exam_date = sqlc.arg(exam_date) 
  AND deleted_at IS NULL 
LIMIT 1;
