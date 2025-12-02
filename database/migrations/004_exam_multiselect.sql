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
  FOREIGN KEY (programid, version) REFERENCES program_versions(programid, name) ON DELETE RESTRICT ON UPDATE CASCADE
) STRICT;

INSERT INTO exams_new 
SELECT id, userid, programid, version, moduleid, comment, exam_date, uploaded_at, accesskey, mime_type, nbytes, checksum 
FROM exams;

DROP TABLE exams;
ALTER TABLE exams_new RENAME TO exams;

CREATE INDEX idx_exams_date ON exams(programid, exam_date DESC);
CREATE INDEX idx_exams_user ON exams(userid);
CREATE INDEX idx_exams_module ON exams(moduleid);

CREATE TRIGGER trg_exams_set_update
AFTER UPDATE ON exams
FOR EACH ROW
BEGIN
  UPDATE exams
     SET uploaded_at = uploaded_at,
         accesskey   = accesskey
   WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
