package web

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/people"
)

func TestLegalChairFollowsMemberAssignments(t *testing.T) {
	views, err := newRenderer()
	if err != nil {
		t.Fatal(err)
	}
	store := people.NewStore()
	if errors := store.SavePerson(people.Person{ID: "neu", Name: "Neue Vorsitzperson", Email: "private@example.test"}); len(errors) != 0 {
		t.Fatal(errors)
	}
	persons := store.People()
	person := persons[len(persons)-1]
	for _, role := range store.Responsibilities("board") {
		if role.Title == "Vorsitz" {
			role.MemberIDs = []string{person.ID}
			if errors := store.SaveResponsibility(role); len(errors) != 0 {
				t.Fatal(errors)
			}
		}
	}
	c := &Handler{views: views, people: store}
	w := httptest.NewRecorder()
	c.Legal(w, httptest.NewRequest("GET", "/impressum", nil))
	body := w.Body.String()
	if w.Code != 200 || !strings.Contains(body, "Neue Vorsitzperson") || strings.Contains(body, "private@example.test") {
		t.Fatalf("imprint did not use public current chair data: %d %s", w.Code, body)
	}
}

func TestDocumentErrorsKeepStatusAndRecoveryWithoutDetails(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/missing", "/community/veranstaltungen/999", "/community/events/missing"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		body := w.Body.String()
		if w.Code != 404 || !strings.Contains(body, `<html lang="de"`) || !strings.Contains(body, `<h1 class="govuk-heading-xl">Seite nicht gefunden</h1>`) || !strings.Contains(body, `href="/sitemap"`) {
			t.Fatalf("missing document recovery for %s: %d %s", path, w.Code, body)
		}
	}
}
