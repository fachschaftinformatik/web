package exams

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/fachschaftinformatik/web/internal/database"
)

type qualification struct {
	program, legacy, name, degree string
}

func (c *Catalog) loadPrograms(ctx context.Context, db database.DBTX) error {
	rows, err := db.QueryContext(ctx, `SELECT p.id,p.name,pq.study_program_id,COALESCE(sp.name,p.name),COALESCE(pq.degree,'')
		FROM programs p LEFT JOIN program_qualifications pq ON pq.legacy_program_id=p.id
		LEFT JOIN study_programs sp ON sp.id=pq.study_program_id WHERE p.deleted_at IS NULL`)
	if err != nil {
		return err
	}
	var records []struct {
		id, studyID         sql.NullInt64
		name, label, degree string
	}
	for rows.Next() {
		var record struct {
			id, studyID         sql.NullInt64
			name, label, degree string
		}
		if err := rows.Scan(&record.id, &record.name, &record.studyID, &record.label, &record.degree); err != nil {
			rows.Close()
			return err
		}
		records = append(records, record)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	c.qualifications = make(map[int64]qualification)
	c.programLabels = make(map[string]string)
	for _, record := range records {
		legacy, err := lookupID(ctx, db, "programs", record.id.Int64)
		if err != nil {
			return err
		}
		program := legacy
		if record.studyID.Valid {
			program, err = lookupID(ctx, db, "study_programs", record.studyID.Int64)
			if err != nil {
				return err
			}
		}
		degree := record.degree
		if degree == "" {
			_, degree = ProgramName(record.name)
		}
		c.qualifications[record.id.Int64] = qualification{program: program, legacy: legacy, name: record.label, degree: DegreeValue(degree)}
		c.programLabels[program] = record.label
	}
	return nil
}

// NormalizeFilter translates legacy qualification references at the boundary.
// A qualification implies its degree only when no separate degree was selected.
func (c Catalog) NormalizeFilter(f Filter) Filter {
	var programs, inferred []string
	infer := len(f.Programs) > 0
	for _, value := range f.Programs {
		matched := false
		for _, q := range c.qualifications {
			if value == q.legacy && q.legacy != q.program {
				if f.legacyPrograms == nil {
					f.legacyPrograms = slices.Clone(f.Programs)
				}
				programs = appendUnique(programs, q.program)
				inferred = appendUnique(inferred, q.degree)
				infer = infer && q.degree != ""
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		name, degree := ProgramName(value)
		for id, label := range c.programLabels {
			base, _ := ProgramName(label)
			if value == label || name == base {
				programs = appendUnique(programs, id)
				matched = true
			}
		}
		if !matched {
			programs = appendUnique(programs, value)
		}
		infer = infer && matched && degree != ""
		inferred = appendUnique(inferred, DegreeValue(degree))
	}
	f.Programs = programs
	if len(f.Degrees) == 0 && infer {
		f.Degrees = inferred
	}
	f.Degrees = slices.Clone(f.Degrees)
	for i := range f.Degrees {
		f.Degrees[i] = DegreeValue(f.Degrees[i])
	}
	return f
}

func (c Catalog) sortPrograms() {
	slices.SortFunc(c.programs, func(a, b string) int {
		if order := strings.Compare(strings.ToLower(c.ProgramLabel(a)), strings.ToLower(c.ProgramLabel(b))); order != 0 {
			return order
		}
		return strings.Compare(a, b)
	})
}

func studyProgramID(ctx context.Context, db database.DBTX, legacy int64) (string, error) {
	var id int64
	err := db.QueryRowContext(ctx, "SELECT study_program_id FROM program_qualifications WHERE legacy_program_id=?", legacy).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return lookupID(ctx, db, "programs", legacy)
	}
	if err != nil {
		return "", err
	}
	return lookupID(ctx, db, "study_programs", id)
}

func (c Catalog) legacyPrograms(f Filter) []string {
	var ids []string
	for id, q := range c.qualifications {
		if slices.Contains(f.Programs, q.program) && c.matchesLegacyProgram(f, q.program, q.degree) {
			ids = append(ids, strconv.FormatInt(id, 10))
		}
	}
	return ids
}

func (c Catalog) matchesLegacyProgram(f Filter, program, degree string) bool {
	if len(f.legacyPrograms) == 0 || slices.Contains(f.legacyPrograms, "all") || slices.Contains(f.legacyPrograms, program) {
		return true
	}
	for _, q := range c.qualifications {
		if q.program == program && q.degree == DegreeValue(degree) && slices.Contains(f.legacyPrograms, q.legacy) {
			return true
		}
	}
	return false
}
