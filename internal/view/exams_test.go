package view

import (
	"net/url"
	"testing"

	"github.com/fachschaftinformatik/web/internal/model/exams"
)

func TestExamsLinksPreserveEffectiveFilters(t *testing.T) {
	p := ExamsPage{Filter: exams.Filter{Query: "Math & logic", Programs: []string{"12", "34"}, Degrees: []string{"bachelor"}, Versions: []string{"2025"}}}
	for _, raw := range []string{p.SearchURL(), p.ModuleURL(Module{Module: exams.Module{ID: "42"}}), p.PageURL(2)} {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		q := u.Query()
		if q.Get("search") != "Math & logic" || len(q["program"]) != 2 || q.Get("degree") != "bachelor" || q.Get("version") != "2025" {
			t.Fatalf("filters lost in %s", raw)
		}
		for _, key := range []string{"q", "seite", "applied", "scope", "choose_scope", "edit_token"} {
			if q.Has(key) {
				t.Fatalf("retired query %s in %s", key, raw)
			}
		}
	}
	if p.PageURL(1) != p.SearchURL() {
		t.Fatal("first page must omit the default page query")
	}
	if (ExamsPage{}).PageURL(1) != "/exams?program=all" {
		t.Fatal("cleared filters must override saved preferences")
	}
	if (ExamsPage{}).ModuleURL(Module{Module: exams.Module{ID: "42"}}) != "/exams/modules/42?program=all" {
		t.Fatal("incorrect module route")
	}
}

func TestSubmissionStatusVocabulary(t *testing.T) {
	for status, label := range map[exams.Status]string{exams.Draft: "Entwurf", exams.Pending: "In Prüfung", exams.Approved: "Freigegeben", exams.Rejected: "Abgelehnt"} {
		if got := SubmissionStatus(status).Label(); got != label {
			t.Fatalf("%s: got %q, want %q", status, got, label)
		}
	}
}
