-- +goose Up
-- +goose StatementBegin
CREATE TABLE modules (
  id INTEGER PRIMARY KEY,
  programid INTEGER NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE(programid, name)
) STRICT;

CREATE INDEX idx_modules_program ON modules(programid);

ALTER TABLE exams ADD COLUMN moduleid INTEGER REFERENCES modules(id) ON DELETE RESTRICT;
ALTER TABLE exams ADD COLUMN comment TEXT;

CREATE INDEX idx_exams_module ON exams(moduleid);

INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Logik und diskrete Strukturen' FROM programs WHERE name LIKE 'Informatik%';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Einführung in die Programmierung' FROM programs WHERE name LIKE 'Informatik%';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Datenbanksysteme' FROM programs WHERE name LIKE 'Informatik%';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Mathematische Grundlagen' FROM programs WHERE name LIKE 'Informatik%';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Theoretische Informatik' FROM programs WHERE name LIKE 'Informatik%';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Rechnernetze' FROM programs WHERE name LIKE 'Informatik%';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_exams_module;
ALTER TABLE exams DROP COLUMN comment;
ALTER TABLE exams DROP COLUMN moduleid;
DROP TABLE modules;
-- +goose StatementEnd
