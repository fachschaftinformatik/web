package exams

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"testing"
)

func TestResolveModulesUsesOnlyAvailableAssignments(t *testing.T) {
	available := []Assignment{
		{ModuleID: "10", Title: "Datenbanken", Program: "Informatik", Degree: "Bachelor", Version: "2023"},
		{ModuleID: "20", Title: "Datenbanken", Program: "Wirtschaftsinformatik", Degree: "Bachelor", Version: "2017"},
		{ModuleID: "30", Title: "Mathematik", Program: "Informatik", Degree: "Bachelor", Version: "2023"},
	}
	selected, fields := ResolveModules(available, []string{"Datenbanken", "Datenbanken"})
	if len(fields) != 0 || len(selected) != 2 || selected[0] != available[0] || selected[1] != available[1] {
		t.Fatalf("invalid expansion: %+v %v", selected, fields)
	}
	if _, fields = ResolveModules(available, []string{"Unbekannt"}); fields["module_id"] == "" {
		t.Fatal("unknown title accepted")
	}
}

func TestOnePDFMultipleAssignmentsAndSharedObjectPrivacy(t *testing.T) {
	s, db := reviewService(t)
	ctx := context.Background()
	if _, err := db.Exec(`INSERT INTO programs(id,name) VALUES(3,'Wirtschaftsinformatik (B. Sc.)');INSERT INTO program_versions(program_id,name) VALUES(3,'2016');INSERT INTO modules(id,program_id,name) VALUES(20,3,'Datenbanken')`); err != nil {
		t.Fatal(err)
	}
	storage := &failingStorage{writableStorage: s.storage.(writableStorage)}
	seedPublicIDs(t, db)
	s.storage = storage
	pdf := []byte("%PDF-1.7\nshared mathematics\n%%EOF")
	input := SubmissionDraft{Date: "2026-03-27", Documents: [][]byte{pdf}, Assignments: []Assignment{{ModuleID: testPublicID(t, db, "modules", "10"), Version: "2023"}, {ModuleID: testPublicID(t, db, "modules", "11"), Version: "2023"}, {ModuleID: testPublicID(t, db, "modules", "20"), Version: "2016"}}}
	saved, err := s.SaveDraft(ctx, 1, input)
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.IDs) != 1 || storage.calls != 1 {
		t.Fatal("identical files were stored more than once")
	}
	current, err := s.Submission(ctx, 1, saved.SubmissionID, false)
	if err != nil || len(current.Assignments) != 3 || current.Count != 1 {
		t.Fatalf("assignments or document lost: %+v %v", current, err)
	}
	other, err := s.SaveDraft(ctx, 3, input)
	if err != nil {
		t.Fatal(err)
	}
	if storage.calls != 1 {
		t.Fatal("known content was uploaded again")
	}
	if _, err := s.Open(ctx, 1, other.IDs[0]); !errors.Is(err, ErrNotFound) {
		t.Fatal("shared object bypassed draft authorization")
	}
	if err := s.Submit(ctx, 1, saved.SubmissionID); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, saved.SubmissionID, true); err != nil {
		t.Fatal(err)
	}
	catalog, err := s.Catalog(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.FindExams(Filter{Degrees: []string{"master"}, Versions: []string{"2016"}})) != 0 {
		t.Fatal("independent filters invented a nonexistent Master/PO combination")
	}
	for _, scope := range []Filter{{Programs: []string{testPublicID(t, db, "programs", "3")}}, {Degrees: []string{"master"}}} {
		found := catalog.FindExams(scope)
		matched := false
		for _, e := range found {
			if e.ID == saved.IDs[0] {
				matched = true
			}
		}
		if !matched {
			t.Fatal("one document was not discoverable through each assignment")
		}
	}
	modules := catalog.FindModules(Filter{})
	for _, m := range modules {
		if m.Title == "Datenbanken" {
			count := 0
			for _, e := range m.Exams {
				if e.ID == saved.IDs[0] {
					count++
				}
			}
			if count != 1 {
				t.Fatal("module repeats the same PDF for each placement")
			}
		}
	}
	// Replacing another owner's draft cannot remove the shared approved object.
	input.SubmissionID = other.SubmissionID
	input.Documents = [][]byte{[]byte("%PDF-1.7\nreplacement\n%%EOF")}
	if _, err := s.SaveDraft(ctx, 3, input); err != nil {
		t.Fatal(err)
	}
	object, err := s.Open(ctx, 3, saved.IDs[0])
	if err != nil {
		t.Fatal(err)
	}
	value, err := io.ReadAll(object)
	object.Close()
	if err != nil || string(value) != string(pdf) {
		t.Fatal("draft replacement damaged a shared public PDF")
	}
	// Invalid PO combinations must fail before touching blob storage.
	before := storage.calls
	input.SubmissionID = ""
	input.Assignments = []Assignment{{ModuleID: testPublicID(t, db, "modules", "11"), Version: "2016"}}
	if _, err := s.SaveDraft(ctx, 1, input); err == nil || storage.calls != before {
		t.Fatal("invalid module/PO combination contacted storage")
	}
}

func TestLegacyCompactionVerifiesBytesAndPreservesOldLinks(t *testing.T) {
	for _, damaged := range []string{"target", "copy"} {
		t.Run(damaged, func(t *testing.T) {
			s, db := examsService(t)
			ctx := context.Background()
			storage := s.storage.(writableStorage)
			checksum := fmt.Sprintf("%x", sha256.Sum256([]byte("%PDF-")))
			if _, err := db.Exec("UPDATE archive SET checksum=?", checksum); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec("INSERT INTO archive_objects(checksum,blob_key,nbytes) VALUES(?,'archive/100/source.pdf',5)", checksum); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec("INSERT INTO archive_legacy_copies(blob_key,checksum) VALUES('archive/101/source.pdf',?)", checksum); err != nil {
				t.Fatal(err)
			}
			key := "archive/100/source.pdf"
			if damaged == "copy" {
				key = "archive/101/source.pdf"
			}
			if err := storage.Upload(ctx, key, bytes.NewReader([]byte("wrong")), 5, "application/pdf"); err != nil {
				t.Fatal(err)
			}
			if err := s.CleanupBlobs(ctx); !errors.Is(err, ErrStorage) {
				t.Fatal("compaction ignored a checksum mismatch")
			}
			object, err := storage.GetObject(ctx, "archive/101/source.pdf")
			if err != nil {
				t.Fatal("unverified legacy copy was deleted")
			}
			object.Close()
			if err := storage.Upload(ctx, key, bytes.NewReader([]byte("%PDF-")), 5, "application/pdf"); err != nil {
				t.Fatal(err)
			}
			if err := s.CleanupBlobs(ctx); err != nil {
				t.Fatal(err)
			}
			if object, err := storage.GetObject(ctx, "archive/101/source.pdf"); err == nil {
				object.Close()
				t.Fatal("verified redundant object was retained")
			}
			object, err = s.Open(ctx, 1, testPublicID(t, db, "documents", "101"))
			if err != nil {
				t.Fatal("compaction broke a legacy download link")
			}
			value, err := io.ReadAll(object)
			object.Close()
			if err != nil || string(value) != "%PDF-" {
				t.Fatal("legacy download changed bytes")
			}
		})
	}
}

func TestRepeatedApprovedUploadAddsAssignmentsWithoutDuplicateDocument(t *testing.T) {
	s, db := reviewService(t)
	ctx := context.Background()
	first := submission(t, db)
	first.Documents = first.Documents[:1]
	saved, err := s.SaveDraft(ctx, 1, first)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(ctx, 1, saved.SubmissionID); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, saved.SubmissionID, true); err != nil {
		t.Fatal(err)
	}
	next := first
	next.Assignments = []Assignment{{ModuleID: testPublicID(t, db, "modules", "11"), Version: "2023"}}
	repeated, err := s.SaveDraft(ctx, 1, next)
	if err != nil {
		t.Fatal(err)
	}
	before, err := s.Catalog(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(ctx, 1, repeated.SubmissionID); err != nil {
		t.Fatal(err)
	}
	if err := s.Review(ctx, 2, repeated.SubmissionID, true); err != nil {
		t.Fatal(err)
	}
	after, err := s.Catalog(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.FindExams(Filter{})) != len(after.FindExams(Filter{})) {
		t.Fatal("review created a duplicate logical document")
	}
	current, err := s.Submission(ctx, 1, repeated.SubmissionID, false)
	if err != nil || len(current.Assignments) != 2 {
		t.Fatal("review did not preserve existing applicability")
	}
	var oldLatest int
	if err := db.QueryRow("SELECT is_latest FROM archive WHERE id=?", testInternalID(t, db, "documents", saved.IDs[0])).Scan(&oldLatest); err != nil || oldLatest != 0 {
		t.Fatal("applicability change did not advance the revision head")
	}
}

func TestExamSettingsPersistAndUseAccountQualification(t *testing.T) {
	s, db := examsService(t)
	ctx := context.Background()
	if _, err := db.Exec("UPDATE users SET program_id=1 WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	defaults, err := s.Settings(ctx, 1)
	if err != nil || len(defaults.Programs) != 1 || defaults.Programs[0] != testPublicID(t, db, "programs", "1") || defaults.Degrees[0] != "bachelor" || len(defaults.Versions) != 0 {
		t.Fatalf("account defaults: %+v %v", defaults, err)
	}
	chosen := Filter{Programs: []string{testPublicID(t, db, "programs", "1")}, Degrees: []string{"bachelor", "master"}, Versions: []string{"2023"}}
	if err := s.SaveSettings(ctx, 1, chosen); err != nil {
		t.Fatal(err)
	}
	restored, err := NewService(db, s.storage).Settings(ctx, 1)
	if err != nil || len(restored.Degrees) != 2 || restored.Versions[0] != "2023" {
		t.Fatal("settings did not survive service restart")
	}
	if err := s.SaveSettings(ctx, 1, Filter{}); err != nil {
		t.Fatal(err)
	}
	empty, err := s.Settings(ctx, 1)
	if err != nil || len(empty.Programs) != 0 {
		t.Fatal("explicitly cleared settings fell back to account program")
	}
	if err := s.SaveSettings(ctx, 1, Filter{Degrees: []string{"invented"}}); err == nil {
		t.Fatal("invalid preference accepted")
	}
}
