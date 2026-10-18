package exams

import (
	"context"
	"reflect"
	"testing"

	"github.com/fachschaftinformatik/web/internal/publicid"
)

func TestCanonicalProgramsAcrossCatalogAssignmentsAndPreferences(t *testing.T) {
	s, db := examsService(t)
	ctx := context.Background()
	if _, err := db.Exec(`INSERT INTO study_programs(id,name) VALUES(1,'Informatik');
		INSERT INTO program_qualifications(legacy_program_id,study_program_id,degree) VALUES(1,1,'Bachelor'),(2,1,'Master');
		UPDATE programs SET name='Informatik';
		UPDATE users SET program_id=2 WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	program, err := publicid.Ensure(ctx, db, "study_programs", "1")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := s.Catalog(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(catalog.Programs(), []string{program}) || catalog.ProgramLabel(program) != "Informatik" {
		t.Fatalf("duplicate or incorrect upload choices: %v", catalog.Programs())
	}
	for _, tc := range []struct {
		degree, version string
		want            int
	}{
		{"", "", 3}, {"bachelor", "", 2}, {"master", "", 1},
		{"master", "2017", 0}, {"bachelor", "2023", 1},
	} {
		filter := Filter{Programs: []string{program}}
		if tc.degree != "" {
			filter.Degrees = []string{tc.degree}
		}
		if tc.version != "" {
			filter.Versions = []string{tc.version}
		}
		page, _, err := s.CatalogPage(ctx, filter, 1, 20)
		if err != nil || len(page.FindExams(filter)) != tc.want || len(catalog.FindExams(filter)) != tc.want {
			t.Fatalf("filter %+v: SQL=%d memory=%d want=%d err=%v", filter, len(page.FindExams(filter)), len(catalog.FindExams(filter)), tc.want, err)
		}
		if !reflect.DeepEqual(page.Programs(), []string{program}) {
			t.Fatalf("duplicate listing choices: %v", page.Programs())
		}
	}
	// 2017 belongs only to Bachelor, even though both degrees share a programme.
	if got := catalog.Assignments(Filter{Programs: []string{program}, Degrees: []string{"master"}, Versions: []string{"2017"}}); len(got) != 0 {
		t.Fatalf("invented Master/2017 assignments: %+v", got)
	}
	available := catalog.Assignments(Filter{Programs: []string{program}, Degrees: []string{"bachelor", "master"}, Versions: []string{"2017", "2023"}})
	if len(available) != 5 {
		t.Fatalf("lost qualification-specific module choices: %+v", available)
	}
	before, err := assignments(ctx, db, 103)
	if err != nil || len(before) != 1 || before[0].ProgramID != program {
		t.Fatalf("existing assignment: %+v %v", before, err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err := writeAssignments(ctx, tx, 103, before); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	after, err := assignments(ctx, db, 103)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("assignment changed on save: before=%+v after=%+v err=%v", before, after, err)
	}
	legacy := testPublicID(t, db, "programs", "2")
	conflict := Filter{Programs: []string{legacy}, Degrees: []string{"bachelor"}}
	page, _, err := s.CatalogPage(ctx, conflict, 1, 20)
	if err != nil || len(page.FindExams(Filter{})) != 0 || len(catalog.FindExams(conflict)) != 0 {
		t.Fatalf("legacy Master reference must not select Bachelor: %v", err)
	}
	for _, ref := range []string{legacy, "Informatik (M. Sc.)"} {
		page, _, err := s.CatalogPage(ctx, Filter{Programs: []string{ref}}, 1, 20)
		if err != nil || len(page.FindExams(Filter{})) != 1 {
			t.Fatalf("legacy link %s: %+v %v", ref, page.FindExams(Filter{}), err)
		}
		if _, err := db.Exec(`INSERT OR REPLACE INTO user_archive_preferences(user_id,programs,degrees,versions) VALUES(1,json_array(?),'[]','[]')`, ref); err != nil {
			t.Fatal(err)
		}
		settings, err := s.Settings(ctx, 1)
		if err != nil || !reflect.DeepEqual(settings.Programs, []string{program}) || !reflect.DeepEqual(settings.Degrees, []string{"master"}) {
			t.Fatalf("legacy preferences: %+v %v", settings, err)
		}
		if err := s.SaveSettings(ctx, 1, settings); err != nil {
			t.Fatal(err)
		}
		var stored string
		if err := db.QueryRow(`SELECT json_extract(programs,'$[0]') FROM user_archive_preferences WHERE user_id=1`).Scan(&stored); err != nil || stored != program {
			t.Fatalf("noncanonical preferences %s: %v", stored, err)
		}
	}
	if _, err := db.Exec(`DELETE FROM user_archive_preferences WHERE user_id=1`); err != nil {
		t.Fatal(err)
	}
	settings, err := s.Settings(ctx, 1)
	if err != nil || !reflect.DeepEqual(settings.Programs, []string{program}) || !reflect.DeepEqual(settings.Degrees, []string{"master"}) || len(settings.Versions) != 0 {
		t.Fatalf("account defaults: %+v %v", settings, err)
	}
}
