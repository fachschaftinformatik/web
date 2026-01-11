-- name: CreateExam :one
INSERT INTO exams (
  id, userid, programid, version, moduleid, exam_date, 
  accesskey, mime_type, nbytes, checksum, comment,
  group_id, edit_version, is_latest
) VALUES (
  sqlc.arg(id), sqlc.arg(userid), sqlc.arg(programid), sqlc.arg(version), 
  sqlc.arg(moduleid), sqlc.arg(exam_date), sqlc.arg(accesskey), 
  sqlc.arg(mime_type), sqlc.arg(nbytes), sqlc.arg(checksum), sqlc.arg(comment),
  sqlc.arg(group_id), sqlc.arg(edit_version), sqlc.arg(is_latest)
) RETURNING *;

-- name: ListExams :many
SELECT e.id, e.programid, e.version, e.exam_date, e.uploaded_at, e.moduleid, e.comment,
       m.name as module_name, u.name as uploader_name, e.group_id, e.edit_version, e.is_latest
FROM exams e
JOIN modules m ON e.moduleid = m.id
JOIN users u ON e.userid = u.id
WHERE (sqlc.narg('programid') IS NULL OR e.programid = sqlc.narg('programid'))
  AND (sqlc.narg('version') IS NULL OR e.version = sqlc.narg('version'))
  AND (sqlc.narg('moduleid') IS NULL OR e.moduleid = sqlc.narg('moduleid'))
  AND e.is_latest = 1
ORDER BY e.exam_date DESC;

-- name: GetExam :one
SELECT * FROM exams WHERE id = sqlc.arg(id) LIMIT 1;

-- name: GetLatestByGroupId :one
SELECT * FROM exams WHERE group_id = sqlc.arg(group_id) AND is_latest = 1 LIMIT 1;

-- name: ClearLatestFlag :exec
UPDATE exams SET is_latest = 0 WHERE group_id = sqlc.arg(group_id);

-- name: UpdateExam :one
UPDATE exams
SET programid = sqlc.arg(programid),
    version = sqlc.arg(version),
    moduleid = sqlc.arg(moduleid),
    exam_date = sqlc.arg(exam_date),
    comment = sqlc.arg(comment)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteExam :exec
DELETE FROM exams WHERE id = sqlc.arg(id);

-- name: ListExamVersions :many
SELECT e.id, e.programid, e.version, e.exam_date, e.uploaded_at, e.moduleid, e.comment,
       m.name as module_name, u.name as uploader_name, e.group_id, e.edit_version, e.is_latest
FROM exams e
JOIN modules m ON e.moduleid = m.id
JOIN users u ON e.userid = u.id
WHERE e.group_id = sqlc.arg(group_id)
ORDER BY e.edit_version DESC;
