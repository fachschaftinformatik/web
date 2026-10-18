package controller

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExamsResponses(t *testing.T) {
	handler, db := testApplication(t)
	seedPublicContent(t, db)
	program := publicRef(t, db, "programs", "1")
	for _, tc := range []struct {
		name, path              string
		htmx, restore, fullPage bool
		contains                string
	}{
		{"ordinary form", "/exams?search=SQL", false, false, true, "Datenbanken"},
		{"partial update", "/exams?search=SQL", true, false, false, "Datenbanken"},
		{"history restore", "/exams?search=SQL", true, true, true, "Datenbanken"},
		{"empty result", "/exams?search=nonexistent", true, false, false, "Keine Module"},
		{"escaped input", "/exams?search=%3Cscript%3E", false, false, true, "&lt;script&gt;"},
		{"multiple PO filters", "/exams?program=" + program + "&version=2023&version=2017", false, false, true, "3 Module gefunden"},
		{"all PO", "/exams", false, false, true, "6 Module gefunden"},
		{"degree filter", "/exams?degree=bachelor&program=" + program + "&version=2023", false, false, true, "2 Module gefunden"},
		{"master fixture empty", "/exams?degree=master", false, false, true, "Keine Module gefunden"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			if tc.htmx {
				req.Header.Set("HX-Request", "true")
			}
			if tc.restore {
				req.Header.Set("HX-History-Restore-Request", "true")
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			body := response.Body.String()
			if response.Code != 200 || !strings.Contains(body, tc.contains) {
				t.Fatalf("unexpected response: status %d, body %s", response.Code, body)
			}
			if got := strings.Contains(body, "<!doctype html>"); got != tc.fullPage {
				t.Fatalf("full page = %v, want %v", got, tc.fullPage)
			}
			if tc.path == "/exams?search=SQL" && strings.Contains(body, "Mathematik I") {
				t.Fatal("unmatched exam leaked into results")
			}
		})
	}
}
