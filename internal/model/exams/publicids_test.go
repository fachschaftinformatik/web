package exams

import (
	"context"
	"errors"
	"testing"

	"github.com/fachschaftinformatik/web/internal/publicid"
)

func TestPublicIDsAndProgramFiltersSurviveRenames(t *testing.T) {
	s, db := examsService(t)
	ctx := context.Background()
	programID := testPublicID(t, db, "programs", "1")
	moduleID := testPublicID(t, db, "modules", "10")
	documentID := testPublicID(t, db, "documents", "102")
	filter := Filter{Programs: []string{programID}, Degrees: []string{"bachelor"}, Versions: []string{"2023"}}
	if _, err := db.Exec("UPDATE programs SET name='Umbenannt (B. Sc.)' WHERE id=1; UPDATE modules SET name='Neuer Modultitel' WHERE id=10"); err != nil {
		t.Fatal(err)
	}
	for _, paged := range []bool{false, true} {
		var catalog Catalog
		var err error
		if paged {
			catalog, _, err = s.CatalogPage(ctx, filter, 1, 20)
		} else {
			catalog, err = s.Catalog(ctx)
		}
		if err != nil {
			t.Fatal(err)
		}
		found := catalog.FindExams(filter)
		if len(found) != 1 || found[0].ID != documentID || found[0].ModuleID != moduleID || found[0].ProgramID != programID {
			t.Fatalf("stable filter or IDs lost after rename (paged=%v): %+v", paged, found)
		}
		if catalog.ProgramLabel(programID) != "Umbenannt (B. Sc.)" {
			t.Fatal("program label did not follow rename")
		}
	}
	if _, err := s.Open(ctx, 1, "102"); !errors.Is(err, ErrNotFound) {
		t.Fatal("internal document ID accepted at service boundary")
	}
	if _, _, err := s.ModulePage(ctx, "10", Filter{}, 1, 20); !errors.Is(err, ErrNotFound) {
		t.Fatal("internal module ID accepted at service boundary")
	}
	if _, err := s.Revision(ctx, 1, "102"); !errors.Is(err, ErrNotFound) {
		t.Fatal("internal revision ID accepted at service boundary")
	}
}

func TestAllProgramsSettingsAndLegacyPreferences(t *testing.T) {
	s, db := examsService(t)
	ctx := context.Background()
	if _, err := db.Exec(`INSERT INTO user_archive_preferences(user_id,programs,degrees,versions) VALUES(1,'["Informatik"]','["Bachelor"]','["2023"]')`); err != nil {
		t.Fatal(err)
	}
	f, err := s.Settings(ctx, 1)
	if err != nil || len(f.Programs) != 2 || f.Degrees[0] != "bachelor" {
		t.Fatalf("legacy preferences: %+v %v", f, err)
	}
	for _, id := range f.Programs {
		if !publicid.Valid(id) {
			t.Fatal("legacy display name leaked into filter")
		}
	}
	all := Filter{Programs: []string{"all"}, Degrees: []string{"bachelor"}}
	if err := s.SaveSettings(ctx, 1, all); err != nil {
		t.Fatal(err)
	}
	restored, err := s.Settings(ctx, 1)
	if err != nil || len(restored.Programs) != 1 || restored.Programs[0] != "all" {
		t.Fatalf("all sentinel lost: %+v %v", restored, err)
	}
	catalog, total, err := s.CatalogPage(ctx, restored, 1, 20)
	if err != nil || total != 1 || len(catalog.FindExams(restored)) != 2 {
		t.Fatalf("all-program SQL/in-memory mismatch: %d %v", total, err)
	}
}

func TestDraftReplacementNeverReusesPublishedIdentifier(t *testing.T) {
	s, db := examsService(t)
	ctx := context.Background()
	input := submission(t, db)
	first, err := s.SaveDraft(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	input.SubmissionID = first.SubmissionID
	input.Documents = [][]byte{[]byte("%PDF-1.7\nreplacement\n%%EOF")}
	second, err := s.SaveDraft(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	if first.SubmissionID != second.SubmissionID || first.IDs[0] == second.IDs[0] {
		t.Fatal("draft identity changed or document identity reused")
	}
	if _, err := s.Open(ctx, 1, first.IDs[0]); !errors.Is(err, ErrNotFound) {
		t.Fatal("replaced draft document remained accessible")
	}
	object, err := s.Open(ctx, 1, second.IDs[0])
	if err != nil {
		t.Fatal(err)
	}
	object.Close()
}

func TestNewDocumentAfterHardDeletionNeverReusesInternalID(t *testing.T) {
	s, db := examsService(t)
	ctx := context.Background()
	first, err := s.SaveDraft(ctx, 1, submission(t, db))
	if err != nil {
		t.Fatal(err)
	}
	internalID := testInternalID(t, db, "documents", first.IDs[0])
	if _, err := db.Exec("DELETE FROM archive WHERE id=?", internalID); err != nil {
		t.Fatal(err)
	}
	second, err := s.SaveDraft(ctx, 1, submission(t, db))
	if err != nil {
		t.Fatal(err)
	}
	if second.IDs[0] == first.IDs[0] || testInternalID(t, db, "documents", second.IDs[0]) == internalID {
		t.Fatal("new document reused a permanently reserved identity")
	}
}
