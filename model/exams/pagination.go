package exams

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/fachschaftinformatik/web/internal/publicid"
)

// CatalogPage selects matching module groups in SQL before loading their exams.
// Keeping a module's exam documents together preserves the catalog's grouping.
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
	result := Catalog{modules: []Module{}, programLabels: map[string]string{}}
	if err := result.loadPrograms(ctx, tx); err != nil {
		return result, 0, err
	}
	f = result.NormalizeFilter(f)
	var moduleTitle string
	if moduleID != "" {
		internalID, err := resolveID(ctx, tx, "modules", moduleID)
		if err != nil {
			return result, 0, err
		}
		err = tx.QueryRowContext(ctx, `SELECT m.name FROM modules m JOIN programs p ON p.id=m.program_id AND p.deleted_at IS NULL WHERE m.id=? AND m.deleted_at IS NULL`, internalID).Scan(&moduleTitle)
		if errors.Is(err, sql.ErrNoRows) {
			return result, 0, ErrNotFound
		}
		if err != nil {
			return result, 0, err
		}
		result.modules = append(result.modules, Module{ID: moduleID, Title: moduleTitle})
	}
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT m.program_id FROM modules m JOIN programs p ON p.id=m.program_id AND p.deleted_at IS NULL WHERE m.deleted_at IS NULL`)
	if err != nil {
		return result, 0, err
	}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return result, 0, err
		}
		q := result.qualifications[id]
		result.programs = appendUnique(result.programs, q.program)
		result.degrees = appendUnique(result.degrees, q.degree)
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
	cte := `WITH entries AS (SELECT a.id,b.module_id,p.id AS program_id,b.version,a.exam_date,COALESCE(a.comment,'') AS comment,a.nbytes,(SELECT original.user_id FROM archive original WHERE original.group_id=a.group_id ORDER BY original.edit_version,original.id LIMIT 1) AS owner_id,m.name AS title,COALESCE(m.alias,'') AS alias,` + programExpr + ` AS program,lower(` + degreeExpr + `) AS degree FROM archive a JOIN archive_placements b ON b.archive_id=a.id JOIN modules m ON m.id=b.module_id AND m.deleted_at IS NULL JOIN programs p ON p.id=m.program_id AND p.deleted_at IS NULL LEFT JOIN program_qualifications pq ON pq.legacy_program_id=p.id LEFT JOIN study_programs sp ON sp.id=pq.study_program_id WHERE a.is_latest=1 AND a.deleted_at IS NULL AND a.mime_type='application/pdf' AND NOT EXISTS(SELECT 1 FROM archive_entry_aliases x WHERE x.legacy_id=a.id AND x.canonical_id!=a.id) AND (a.request_id IS NULL OR EXISTS(SELECT 1 FROM archive_requests r WHERE r.id=a.request_id AND r.status='approved'))),matching AS (SELECT * FROM entries WHERE 1=1`
	args := []any{}
	if moduleID != "" {
		cte += " AND title=?"
		args = append(args, moduleTitle)
	}
	for _, filter := range []struct {
		column string
		values []string
	}{{"program_id", f.Programs}, {"degree", f.Degrees}, {"version", f.Versions}} {
		if filter.column == "program_id" && slices.Contains(filter.values, "all") {
			continue
		}
		if filter.column == "program_id" && len(filter.values) > 0 {
			filter.values = result.legacyPrograms(f)
			if len(filter.values) == 0 {
				filter.values = []string{"-1"}
			}
		}
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
			return result, 0, fmt.Errorf("exam document %s: %w", e.ID, err)
		}
		e.Date = parsed.Format("2006-01-02")
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
	for i := range result.exams {
		e := &result.exams[i]
		internalModule := e.ModuleID
		e.ID, err = publicid.Lookup(ctx, tx, "documents", e.ID)
		if err != nil {
			return result, 0, err
		}
		e.ModuleID, err = publicid.Lookup(ctx, tx, "modules", internalModule)
		if err != nil {
			return result, 0, err
		}
		var programID int64
		if err := tx.QueryRowContext(ctx, "SELECT program_id FROM modules WHERE id=?", internalModule).Scan(&programID); err != nil {
			return result, 0, err
		}
		e.ProgramID = result.qualifications[programID].program
		e.Degree = DegreeLabel(e.Degree)
		e.Assignments = []Assignment{{ModuleID: e.ModuleID, ProgramID: e.ProgramID, Version: e.Version, Title: e.Title, Program: e.Program, Degree: e.Degree}}
	}
	for i := range result.modules {
		if result.modules[i].ID == moduleID {
			continue
		}
		result.modules[i].ID, err = publicid.Lookup(ctx, tx, "modules", result.modules[i].ID)
		if err != nil {
			return result, 0, err
		}
	}
	result.sortPrograms()
	slices.Sort(result.degrees)
	slices.Sort(result.versions)
	slices.Reverse(result.versions)
	if err := tx.Commit(); err != nil {
		return result, 0, err
	}
	return result, total, nil
}
