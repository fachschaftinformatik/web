package files

import (
	"context"
	"database/sql"
	"github.com/fachschaftinformatik/web/internal/data"
	"strconv"
)

// Assignment is one valid combination, rather than independent sets whose
// Cartesian product might invent study regulations that do not exist.
type Assignment struct {
	ModuleID string
	Version  string
	Title    string
	Program  string
	Degree   string
}

func (a Assignment) VersionLabel() string { return VersionLabel(a.Version) }

func (s Submission) placements() []Assignment {
	if len(s.Assignments) > 0 {
		return s.Assignments
	}
	return []Assignment{{ModuleID: s.ModuleID, Version: s.Version}}
}

func assignments(ctx context.Context, db data.DBTX, entryID int64) ([]Assignment, error) {
	rows, err := db.QueryContext(ctx, `SELECT CAST(b.module_id AS TEXT),b.version,m.name,COALESCE(sp.name,p.name),COALESCE(pq.degree,'')
    FROM archive_placements b JOIN modules m ON m.id=b.module_id JOIN programs p ON p.id=m.program_id
    LEFT JOIN program_qualifications pq ON pq.legacy_program_id=p.id LEFT JOIN study_programs sp ON sp.id=pq.study_program_id
    WHERE b.archive_id=COALESCE((SELECT canonical_id FROM archive_entry_aliases WHERE legacy_id=?),?)
    ORDER BY m.name,p.name,b.version`, entryID, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Assignment
	for rows.Next() {
		var a Assignment
		if err := rows.Scan(&a.ModuleID, &a.Version, &a.Title, &a.Program, &a.Degree); err != nil {
			return nil, err
		}
		if a.Degree == "" {
			a.Program, a.Degree = ProgramName(a.Program)
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func canonicalEntry(ctx context.Context, db data.DBTX, entryID int64) (data.Archive, error) {
	var canonical int64
	if err := db.QueryRowContext(ctx, `SELECT COALESCE((SELECT canonical_id FROM archive_entry_aliases WHERE legacy_id=?),?)`, entryID, entryID).Scan(&canonical); err != nil {
		return data.Archive{}, err
	}
	return data.New(db).GetArchiveEntry(ctx, canonical)
}

func writeAssignments(ctx context.Context, tx *sql.Tx, entryID int64, values []Assignment) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM archive_assignments WHERE archive_id=?", entryID); err != nil {
		return err
	}
	for _, a := range values {
		module, err := strconv.ParseInt(a.ModuleID, 10, 64)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO archive_assignments(archive_id,module_id,version) VALUES(?,?,?)", entryID, module, a.Version); err != nil {
			return err
		}
	}
	return nil
}
