package files

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
)

func (s *Service) Preferences(ctx context.Context, userID int64) (Filter, error) {
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
		return f, nil
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
	result := Filter{Programs: []string{program}}
	if degree != "" {
		result.Degrees = []string{degree}
	}
	return result, nil
}

func (s *Service) SavePreferences(ctx context.Context, userID int64, filter Filter) error {
	if _, err := accountRole(ctx, s.db, userID); err != nil {
		return err
	}
	catalog, err := s.Catalog(ctx)
	if err != nil {
		return err
	}
	for _, group := range []struct{ values, known []string }{{filter.Programs, catalog.Programs()}, {filter.Degrees, catalog.Degrees()}, {filter.Versions, catalog.Versions()}} {
		if len(group.values) > 100 || slices.ContainsFunc(group.values, func(value string) bool { return !slices.Contains(group.known, value) }) {
			return ValidationErrors{"preferences": "Wähle gültige Studiengänge, Abschlüsse und Prüfungsordnungen aus."}
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
