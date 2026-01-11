-- +goose Up
-- +goose StatementBegin
CREATE TABLE exams_new (
  id           TEXT PRIMARY KEY,
  userid       TEXT NOT NULL
                 REFERENCES users(id)
                 ON DELETE RESTRICT ON UPDATE CASCADE,
  programid    INTEGER NOT NULL,
  version      TEXT NOT NULL,
  moduleid     INTEGER REFERENCES modules(id) ON DELETE RESTRICT,
  comment      TEXT,
  exam_date    TEXT NOT NULL,
  uploaded_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  accesskey    TEXT NOT NULL,
  mime_type    TEXT NOT NULL CHECK (mime_type IN ('application/pdf')),
  nbytes       INTEGER NOT NULL,
  checksum     TEXT NOT NULL,
  group_id     TEXT NOT NULL,
  edit_version INTEGER NOT NULL,
  is_latest    INTEGER NOT NULL DEFAULT 1,
  FOREIGN KEY (programid, version) REFERENCES program_versions(programid, name) ON DELETE RESTRICT ON UPDATE CASCADE
) STRICT;

INSERT INTO exams_new (
  id, userid, programid, version, moduleid, comment, exam_date, 
  uploaded_at, accesskey, mime_type, nbytes, checksum, 
  group_id, edit_version, is_latest
)
SELECT 
  id, userid, programid, version, moduleid, comment, exam_date, 
  uploaded_at, accesskey, mime_type, nbytes, checksum, 
  id, 1, 1
FROM exams;

DROP TABLE exams;
ALTER TABLE exams_new RENAME TO exams;

CREATE INDEX idx_exams_date ON exams(programid, exam_date DESC);
CREATE INDEX idx_exams_user ON exams(userid);
CREATE INDEX idx_exams_module ON exams(moduleid);
CREATE INDEX idx_exams_group ON exams(group_id);
CREATE INDEX idx_exams_latest ON exams(is_latest);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE exams;
-- Note: Reverting fully would require recreating the old schema and dropping the new columns.
-- For simplicity in this dev environment, we assume migrations go forward.
-- +goose StatementEnd
