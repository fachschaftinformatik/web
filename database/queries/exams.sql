-- name: CreateExam :one
INSERT INTO exams (
  id, userid, programid, version, moduleid, exam_date, 
  accesskey, mime_type, nbytes, checksum, comment
) VALUES (
  sqlc.arg(id), sqlc.arg(userid), sqlc.arg(programid), sqlc.arg(version), 
  sqlc.arg(moduleid), sqlc.arg(exam_date), sqlc.arg(accesskey), 
  sqlc.arg(mime_type), sqlc.arg(nbytes), sqlc.arg(checksum), sqlc.arg(comment)
) RETURNING *;

-- name: ListExams :many
SELECT e.id, e.programid, e.version, e.exam_date, e.uploaded_at, e.moduleid, e.comment,
       m.name as module_name, u.name as uploader_name
FROM exams e
JOIN modules m ON e.moduleid = m.id
JOIN users u ON e.userid = u.id
WHERE (sqlc.narg('programid') IS NULL OR e.programid = sqlc.narg('programid'))
  AND (sqlc.narg('version') IS NULL OR e.version = sqlc.narg('version'))
  AND (sqlc.narg('moduleid') IS NULL OR e.moduleid = sqlc.narg('moduleid'))
ORDER BY e.exam_date DESC;

-- name: GetExam :one
SELECT * FROM exams WHERE id = sqlc.arg(id) LIMIT 1;

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
