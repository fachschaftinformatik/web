package controller

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/model/members"
)

func TestLegalChairFollowsMemberAssignments(t *testing.T) {
	c, _ := testApplication(t)
	store := c.members
	if errors := store.SaveMember(members.Member{ID: "create", Name: "Neue Vorsitzperson", Email: "private@example.test"}); len(errors) != 0 {
		t.Fatal(errors)
	}
	persons := store.Members()
	person := persons[len(persons)-1]
	for _, role := range store.Roles("board") {
		if role.Title == "Vorsitz" {
			role.Title = "Sprecherkreis"
			role.MemberIDs = []string{person.ID}
			if errors := store.SaveRole(role); len(errors) != 0 {
				t.Fatal(errors)
			}
		}
	}
	w := httptest.NewRecorder()
	c.Legal(w, httptest.NewRequest("GET", "/legal", nil))
	body := w.Body.String()
	if w.Code != 200 || !strings.Contains(body, "Neue Vorsitzperson") || strings.Contains(body, "private@example.test") {
		t.Fatalf("imprint did not use public current chair data: %d %s", w.Code, body)
	}
}

func TestDocumentErrorsKeepStatusAndRecoveryWithoutDetails(t *testing.T) {
	handler, err := testHandler(t)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/missing", "/events/999", "/media/albums/missing"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		body := w.Body.String()
		if w.Code != 404 || !strings.Contains(body, `<html lang="de"`) || !strings.Contains(body, `<h1 class="govuk-heading-xl">Seite nicht gefunden</h1>`) || !strings.Contains(body, `href="/sitemap"`) {
			t.Fatalf("missing document recovery for %s: %d %s", path, w.Code, body)
		}
	}
}
