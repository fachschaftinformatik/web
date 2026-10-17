-- +goose Up
CREATE TABLE study_programs (id INTEGER PRIMARY KEY,name TEXT NOT NULL UNIQUE) STRICT;
CREATE TABLE program_qualifications (
    legacy_program_id INTEGER PRIMARY KEY REFERENCES programs(id) ON DELETE RESTRICT,
    study_program_id INTEGER NOT NULL REFERENCES study_programs(id) ON DELETE RESTRICT,
    degree TEXT NOT NULL CHECK(degree IN ('Bachelor','Master',''))
) STRICT;
INSERT INTO study_programs(name)
SELECT DISTINCT CASE WHEN instr(name,' (')>0 THEN substr(name,1,instr(name,' (')-1) ELSE name END FROM programs;
INSERT INTO program_qualifications(legacy_program_id,study_program_id,degree)
SELECT p.id,s.id,CASE WHEN p.name LIKE '%(B.%' THEN 'Bachelor' WHEN p.name LIKE '%(M.%' THEN 'Master' ELSE '' END
FROM programs p JOIN study_programs s ON s.name=CASE WHEN instr(p.name,' (')>0 THEN substr(p.name,1,instr(p.name,' (')-1) ELSE p.name END;

CREATE TABLE archive_objects (
    checksum TEXT PRIMARY KEY,
    blob_key TEXT NOT NULL UNIQUE,
    nbytes INTEGER NOT NULL CHECK(nbytes >= 0)
) STRICT;

CREATE TABLE archive_group_aliases (
    group_id INTEGER PRIMARY KEY,
    canonical_group_id INTEGER NOT NULL
) STRICT;

-- Only combine legacy groups whose complete revision histories agree. Module
-- and PO are placements, not part of the file or its revision identity.
INSERT INTO archive_group_aliases(group_id, canonical_group_id)
WITH histories AS (
    SELECT group_id, (
        SELECT json_group_array(json_array(checksum,nbytes,exam_date,comment,user_id,
                                           edit_version,is_latest,deleted_at))
        FROM (SELECT * FROM archive h WHERE h.group_id=g.group_id ORDER BY edit_version,id)
    ) AS signature
    FROM archive g GROUP BY group_id
    HAVING COUNT(request_id)=0
)
SELECT group_id,MIN(group_id) OVER (PARTITION BY signature) FROM histories;

CREATE TABLE archive_entry_aliases (
    legacy_id INTEGER PRIMARY KEY REFERENCES archive(id) ON DELETE CASCADE,
    canonical_id INTEGER NOT NULL REFERENCES archive(id) ON DELETE RESTRICT
) STRICT;
INSERT INTO archive_entry_aliases(legacy_id,canonical_id)
SELECT a.id,MIN(c.id) FROM archive a
JOIN archive_group_aliases g ON g.group_id=a.group_id
JOIN archive c ON c.group_id=g.canonical_group_id
    AND c.edit_version=a.edit_version AND c.checksum=a.checksum
GROUP BY a.id;

CREATE TABLE archive_assignments (
    archive_id INTEGER NOT NULL REFERENCES archive(id) ON DELETE CASCADE,
    module_id INTEGER NOT NULL REFERENCES modules(id) ON DELETE RESTRICT,
    version TEXT NOT NULL,
    PRIMARY KEY(archive_id,module_id,version)
) STRICT;
INSERT INTO archive_assignments(archive_id,module_id,version)
SELECT DISTINCT COALESCE(x.canonical_id,a.id),a.module_id,a.version
FROM archive a LEFT JOIN archive_entry_aliases x ON x.legacy_id=a.id;

INSERT INTO archive_objects(checksum,blob_key,nbytes)
SELECT checksum,'archive/'||MIN(id)||'/source.pdf',CASE WHEN MIN(nbytes)=MAX(nbytes) THEN MIN(nbytes) ELSE -1 END
FROM archive GROUP BY checksum;

CREATE TABLE archive_legacy_copies (
    blob_key TEXT PRIMARY KEY,
    checksum TEXT NOT NULL REFERENCES archive_objects(checksum) ON DELETE CASCADE
) STRICT;
INSERT INTO archive_legacy_copies(blob_key,checksum)
SELECT 'archive/'||a.id||'/source.pdf',a.checksum FROM archive a
JOIN archive_objects o ON o.checksum=a.checksum
WHERE 'archive/'||a.id||'/source.pdf'!=o.blob_key;

CREATE TABLE archive_blob_cleanup (blob_key TEXT PRIMARY KEY) STRICT;

CREATE TABLE user_archive_preferences (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    programs TEXT NOT NULL CHECK(json_valid(programs) AND json_type(programs)='array'),
    degrees TEXT NOT NULL CHECK(json_valid(degrees) AND json_type(degrees)='array'),
    versions TEXT NOT NULL CHECK(json_valid(versions) AND json_type(versions)='array')
) STRICT;

-- The fallback supports legacy database writers and import tools. New writes
-- always populate archive_assignments explicitly.
CREATE VIEW archive_placements AS
SELECT archive_id,module_id,version FROM archive_assignments
UNION ALL
SELECT a.id,a.module_id,a.version FROM archive a
WHERE NOT EXISTS(SELECT 1 FROM archive_assignments b WHERE b.archive_id=a.id)
  AND NOT EXISTS(SELECT 1 FROM archive_entry_aliases x WHERE x.legacy_id=a.id AND x.canonical_id!=a.id);

-- +goose Down
DROP VIEW archive_placements;
DROP TABLE user_archive_preferences;
DROP TABLE archive_blob_cleanup;
DROP TABLE archive_assignments;
DROP TABLE archive_entry_aliases;
DROP TABLE archive_group_aliases;
DROP TABLE archive_legacy_copies;
DROP TABLE archive_objects;
DROP TABLE program_qualifications;
DROP TABLE study_programs;
