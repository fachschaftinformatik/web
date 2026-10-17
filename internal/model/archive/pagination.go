package archive

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// CatalogPage selects matching module groups in SQL before loading their exams.
// Keeping a module's reconstructions together preserves the archive's grouping.
func (s *Service) CatalogPage(ctx context.Context, f Filter, page, size int) (Catalog, int, error) {
	return s.catalogPage(ctx, f, page, size, "")
}

func (s *Service) ModulePage(ctx context.Context, id string, f Filter, page, size int) (Module, int, error) {
	c, total, err := s.catalogPage(ctx, f, page, size, id)
	if err != nil {
		return Module{}, 0, err
	}
	m, _ := c.FindModule(id, Filter{})
	return m, total, nil
}

func (s *Service) catalogPage(ctx context.Context, f Filter, page, size int, moduleID string) (Catalog, int, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Catalog{}, 0, err
	}
	defer tx.Rollback()
	result := Catalog{modules: []Module{}}
	var moduleTitle string
	if moduleID != "" {
		err := tx.QueryRowContext(ctx, `SELECT m.name FROM modules m JOIN programs p ON p.id=m.program_id AND p.deleted_at IS NULL WHERE m.id=? AND m.deleted_at IS NULL`, moduleID).Scan(&moduleTitle)
		if errors.Is(err, sql.ErrNoRows) {
			return result, 0, ErrNotFound
		}
		if err != nil {
			return result, 0, err
		}
		result.modules = append(result.modules, Module{ID: moduleID, Title: moduleTitle})
	}
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT COALESCE(sp.name,p.name),COALESCE(pq.degree,'') FROM modules m JOIN programs p ON p.id=m.program_id AND p.deleted_at IS NULL LEFT JOIN program_qualifications pq ON pq.legacy_program_id=p.id LEFT JOIN study_programs sp ON sp.id=pq.study_program_id WHERE m.deleted_at IS NULL`)
	if err != nil {
		return result, 0, err
	}
	for rows.Next() {
		var program, degree string
		if err := rows.Scan(&program, &degree); err != nil {
			rows.Close()
			return result, 0, err
		}
		if degree == "" {
			program, degree = ProgramName(program)
		}
		result.programs = appendUnique(result.programs, program)
		result.degrees = appendUnique(result.degrees, degree)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, 0, err
	}
	rows, err = tx.QueryContext(ctx, `SELECT DISTINCT pv.name FROM program_versions pv JOIN programs p ON p.id=pv.program_id AND p.deleted_at IS NULL UNION SELECT DISTINCT version FROM archive_placements`)
	if err != nil {
		return result, 0, err
	}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			rows.Close()
			return result, 0, err
		}
		result.versions = appendUnique(result.versions, version)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, 0, err
	}
	// Normalize legacy degree suffixes using the same rules as ProgramName.
	rawProgram := `COALESCE(sp.name,p.name)`
	programExpr := rawProgram
	degreeExpr := `COALESCE(pq.degree,'')`
	for _, suffix := range []struct{ value, degree string }{{" (B. Sc.)", "Bachelor"}, {" (M. Sc.)", "Master"}, {" (B.Sc.)", "Bachelor"}, {" (M.Sc.)", "Master"}} {
		programExpr = `CASE WHEN COALESCE(pq.degree,'')='' AND substr(` + rawProgram + `,-` + strconv.Itoa(len(suffix.value)) + `)='` + suffix.value + `' THEN substr(` + rawProgram + `,1,length(` + rawProgram + `)-` + strconv.Itoa(len(suffix.value)) + `) ELSE ` + programExpr + ` END`
		degreeExpr = `CASE WHEN COALESCE(pq.degree,'')='' AND substr(` + rawProgram + `,-` + strconv.Itoa(len(suffix.value)) + `)='` + suffix.value + `' THEN '` + suffix.degree + `' ELSE ` + degreeExpr + ` END`
	}
	cte := `WITH entries AS (SELECT a.id,b.module_id,b.version,a.exam_date,COALESCE(a.comment,'') AS comment,a.nbytes,(SELECT original.user_id FROM archive original WHERE original.group_id=a.group_id ORDER BY original.edit_version,original.id LIMIT 1) AS owner_id,m.name AS title,COALESCE(m.alias,'') AS alias,` + programExpr + ` AS program,` + degreeExpr + ` AS degree FROM archive a JOIN archive_placements b ON b.archive_id=a.id JOIN modules m ON m.id=b.module_id AND m.deleted_at IS NULL JOIN programs p ON p.id=m.program_id AND p.deleted_at IS NULL LEFT JOIN program_qualifications pq ON pq.legacy_program_id=p.id LEFT JOIN study_programs sp ON sp.id=pq.study_program_id WHERE a.is_latest=1 AND a.deleted_at IS NULL AND a.mime_type='application/pdf' AND NOT EXISTS(SELECT 1 FROM archive_entry_aliases x WHERE x.legacy_id=a.id AND x.canonical_id!=a.id) AND (a.request_id IS NULL OR EXISTS(SELECT 1 FROM archive_requests r WHERE r.id=a.request_id AND r.status='approved'))),matching AS (SELECT * FROM entries WHERE 1=1`
	args := []any{}
	if moduleID != "" {
		cte += " AND title=?"
		args = append(args, moduleTitle)
	}
	for _, filter := range []struct {
		column string
		values []string
	}{{"program", f.Programs}, {"degree", f.Degrees}, {"version", f.Versions}} {
		if len(filter.values) > 0 {
			cte += " AND " + filter.column + " IN (" + strings.TrimSuffix(strings.Repeat("?,", len(filter.values)), ",") + ")"
			for _, v := range filter.values {
				args = append(args, v)
			}
		}
	}
	if query := strings.ToLower(strings.TrimSpace(f.Query)); query != "" {
		cte += ` AND instr(lower(title||' '||comment||' '||alias),?)>0`
		args = append(args, query)
	}
	cte += `)`
	var total int
	countColumn := "title"
	if moduleID != "" {
		countColumn = "id"
	}
	if err := tx.QueryRowContext(ctx, cte+` SELECT COUNT(DISTINCT `+countColumn+`) FROM matching`, args...).Scan(&total); err != nil {
		return result, 0, err
	}
	size = min(100, max(1, size))
	page = max(1, min(page, max(1, (total+size-1)/size)))
	args = append(args, size, (page-1)*size)
	selection := `,selected AS (SELECT title,MAX(exam_date) AS newest,MAX(id) AS last_id FROM matching GROUP BY title ORDER BY newest DESC,last_id DESC,title LIMIT ? OFFSET ?) SELECT id,module_id,version,exam_date,comment,nbytes,owner_id,matching.title,alias,program,degree FROM matching JOIN selected USING(title) ORDER BY selected.newest DESC,selected.last_id DESC,matching.title,exam_date DESC,id DESC,module_id`
	if moduleID != "" {
		selection = `,selected AS (SELECT id,MAX(exam_date) AS newest FROM matching GROUP BY id ORDER BY newest DESC,id DESC LIMIT ? OFFSET ?) SELECT id,module_id,version,exam_date,comment,nbytes,owner_id,matching.title,alias,program,degree FROM matching JOIN selected USING(id) ORDER BY exam_date DESC,id DESC,module_id`
	}
	rows, err = tx.QueryContext(ctx, cte+selection, args...)
	if err != nil {
		return result, 0, err
	}
	for rows.Next() {
		var e Exam
		var date string
		var owner sql.NullInt64
		if err := rows.Scan(&e.ID, &e.ModuleID, &e.Version, &date, &e.Description, &e.Bytes, &owner, &e.Title, &e.alias, &e.Program, &e.Degree); err != nil {
			rows.Close()
			return result, 0, err
		}
		parsed, err := parseDate(date)
		if err != nil {
			rows.Close()
			return result, 0, fmt.Errorf("archive entry %s: %w", e.ID, err)
		}
		e.Date = parsed.Format("02.01.2006")
		e.OwnerID = owner.Int64
		e.Assignments = []Assignment{{ModuleID: e.ModuleID, Version: e.Version, Title: e.Title, Program: e.Program, Degree: e.Degree}}
		result.exams = append(result.exams, e)
		if !slices.ContainsFunc(result.modules, func(m Module) bool { return m.Title == e.Title }) {
			result.modules = append(result.modules, Module{ID: e.ModuleID, Title: e.Title})
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, 0, err
	}
	slices.Sort(result.programs)
	slices.Sort(result.degrees)
	slices.Sort(result.versions)
	slices.Reverse(result.versions)
	if err := tx.Commit(); err != nil {
		return result, 0, err
	}
	return result, total, nil
}
