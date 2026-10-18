package exams

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"slices"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/publicid"
)

// Assignment is one valid combination, rather than independent sets whose
// Cartesian product might invent study regulations that do not exist.
type Assignment struct {
	ModuleID  string
	ProgramID string
	Version   string
	Title     string
	Program   string
	Degree    string
}

func (a Assignment) Key() string { return a.ModuleID + "|" + a.Version }

// Scope describes the qualifications represented by these assignments.
func Scope(assignments []Assignment) Filter {
	var scope Filter
	add := func(values []string, value string) []string {
		if !slices.Contains(values, value) {
			return append(values, value)
		}
		return values
	}
	for _, a := range assignments {
		scope.Programs = add(scope.Programs, a.ProgramID)
		scope.Degrees = add(scope.Degrees, DegreeValue(a.Degree))
		scope.Versions = add(scope.Versions, a.Version)
	}
	slices.Sort(scope.Programs)
	slices.Sort(scope.Degrees)
	slices.Sort(scope.Versions)
	return scope
}

func (r Submission) Scope() Filter {
	var values []Assignment
	for _, document := range r.Documents {
		values = append(values, document.Assignments...)
	}
	return Scope(values)
}

// Assignments returns only catalogue-backed combinations within the scope.
func (c Catalog) Assignments(scope Filter) []Assignment {
	var result []Assignment
	for _, module := range c.Modules() {
		program, degree := module.Qualification()
		if !slices.Contains(scope.Programs, "all") && !slices.Contains(scope.Programs, module.ProgramID) || !slices.Contains(scope.Degrees, DegreeValue(degree)) {
			continue
		}
		for _, version := range c.ProgramVersions(module.ProgramID) {
			if slices.Contains(scope.Versions, version) {
				result = append(result, Assignment{ModuleID: module.ID, ProgramID: module.ProgramID, Version: version, Title: module.Title, Program: program, Degree: degree})
			}
		}
	}
	return result
}

// ResolveModules expands selected titles without inventing qualification pairs.
func ResolveModules(available []Assignment, titles []string) ([]Assignment, ValidationErrors) {
	fields := ValidationErrors{}
	for _, title := range titles {
		if !slices.ContainsFunc(available, func(a Assignment) bool { return a.Title == title }) {
			fields["module_id"] = "Wähle gültige Module aus."
		}
	}
	var result []Assignment
	for _, a := range available {
		if slices.Contains(titles, a.Title) {
			result = append(result, a)
		}
	}
	return result, fields
}

func (s SubmissionDraft) placements() []Assignment {
	if len(s.Assignments) > 0 {
		return s.Assignments
	}
	return []Assignment{{ModuleID: s.ModuleID, Version: s.Version}}
}

func assignments(ctx context.Context, db database.DBTX, entryID int64) ([]Assignment, error) {
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
	err = errors.Join(rows.Err(), rows.Close())
	if err != nil {
		return nil, err
	}
	for i := range result {
		var programID int64
		if err := db.QueryRowContext(ctx, "SELECT program_id FROM modules WHERE id=?", result[i].ModuleID).Scan(&programID); err != nil {
			return nil, err
		}
		result[i].ProgramID, err = lookupID(ctx, db, "programs", programID)
		if err != nil {
			return nil, err
		}
		result[i].ModuleID, err = publicid.Lookup(ctx, db, "modules", result[i].ModuleID)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func canonicalEntry(ctx context.Context, db database.DBTX, entryID int64) (database.Archive, error) {
	var canonical int64
	if err := db.QueryRowContext(ctx, `SELECT COALESCE((SELECT canonical_id FROM archive_entry_aliases WHERE legacy_id=?),?)`, entryID, entryID).Scan(&canonical); err != nil {
		return database.Archive{}, err
	}
	return database.New(db).GetArchiveEntry(ctx, canonical)
}

func writeAssignments(ctx context.Context, tx *sql.Tx, entryID int64, values []Assignment) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM archive_assignments WHERE archive_id=?", entryID); err != nil {
		return err
	}
	for _, a := range values {
		module, err := resolveID(ctx, tx, "modules", a.ModuleID)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO archive_assignments(archive_id,module_id,version) VALUES(?,?,?)", entryID, module, a.Version); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Settings(ctx context.Context, userID int64) (Filter, error) {
	if _, err := accountRole(ctx, s.db, userID); err != nil {
		return Filter{}, err
	}
	var program, degree, version string
	err := s.db.QueryRowContext(ctx, "SELECT programs,degrees,versions FROM user_archive_preferences WHERE user_id=?", userID).Scan(&program, &degree, &version)
	if err == nil {
		var f Filter
		if err := json.Unmarshal([]byte(program), &f.Programs); err != nil {
			return f, err
		}
		if err := json.Unmarshal([]byte(degree), &f.Degrees); err != nil {
			return f, err
		}
		if err := json.Unmarshal([]byte(version), &f.Versions); err != nil {
			return f, err
		}
		return s.publicFilter(ctx, f)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Filter{}, err
	}
	// Older accounts already have a qualification. Do not guess their PO.
	err = s.db.QueryRowContext(ctx, `SELECT COALESCE(sp.name,p.name),COALESCE(pq.degree,'') FROM users u JOIN programs p ON p.id=u.program_id AND p.deleted_at IS NULL LEFT JOIN program_qualifications pq ON pq.legacy_program_id=p.id LEFT JOIN study_programs sp ON sp.id=pq.study_program_id WHERE u.id=?`, userID).Scan(&program, &degree)
	if errors.Is(err, sql.ErrNoRows) {
		return Filter{}, nil
	}
	if err != nil {
		return Filter{}, err
	}
	if degree == "" {
		program, degree = ProgramName(program)
	}
	var programID int64
	if err := s.db.QueryRowContext(ctx, "SELECT program_id FROM users WHERE id=?", userID).Scan(&programID); err != nil {
		return Filter{}, err
	}
	id, err := lookupID(ctx, s.db, "programs", programID)
	if err != nil {
		return Filter{}, err
	}
	result := Filter{Programs: []string{id}}
	if degree != "" {
		result.Degrees = []string{degree}
	}
	return s.publicFilter(ctx, result)
}

func (s *Service) SaveSettings(ctx context.Context, userID int64, filter Filter) error {
	if _, err := accountRole(ctx, s.db, userID); err != nil {
		return err
	}
	catalog, err := s.Catalog(ctx)
	if err != nil {
		return err
	}
	for _, group := range []struct{ values, known []string }{{filter.Programs, append(catalog.Programs(), "all")}, {filter.Degrees, catalog.Degrees()}, {filter.Versions, catalog.Versions()}} {
		if len(group.values) > 100 || slices.ContainsFunc(group.values, func(value string) bool { return !slices.Contains(group.known, value) }) {
			return ValidationErrors{"settings": "Wähle gültige Studiengänge, Abschlüsse und Prüfungsordnungen aus."}
		}
	}
	encode := func(values []string) string {
		unique := slices.Clone(values)
		slices.Sort(unique)
		unique = slices.Compact(unique)
		if unique == nil {
			unique = []string{}
		}
		value, _ := json.Marshal(unique)
		return string(value)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO user_archive_preferences(user_id,programs,degrees,versions) VALUES(?,?,?,?) ON CONFLICT(user_id) DO UPDATE SET programs=excluded.programs,degrees=excluded.degrees,versions=excluded.versions`, userID, encode(filter.Programs), encode(filter.Degrees), encode(filter.Versions))
	return err
}

// publicFilter translates historical display-name preferences at the boundary.
func (s *Service) publicFilter(ctx context.Context, f Filter) (Filter, error) {
	c, err := s.Catalog(ctx)
	if err != nil {
		return Filter{}, err
	}
	var programs []string
	for _, value := range f.Programs {
		if value == "all" || publicid.Valid(value) {
			programs = appendUnique(programs, value)
			continue
		}
		value, _ = ProgramName(value)
		for _, id := range c.Programs() {
			label := c.ProgramLabel(id)
			base, _ := ProgramName(label)
			if value == label || value == base {
				programs = appendUnique(programs, id)
			}
		}
	}
	f.Programs = programs
	for i := range f.Degrees {
		f.Degrees[i] = DegreeValue(f.Degrees[i])
	}
	return f, nil
}
