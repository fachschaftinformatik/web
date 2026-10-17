package controller

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestArchiveResponses(t *testing.T) {
	handler, err := testHandler(t)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, path              string
		htmx, restore, fullPage bool
		contains                string
	}{
		{"ordinary form", "/archive?q=SQL", false, false, true, "Datenbanken"},
		{"partial update", "/archive?q=SQL", true, false, false, "Datenbanken"},
		{"history restore", "/archive?q=SQL", true, true, true, "Datenbanken"},
		{"empty result", "/archive?q=nonexistent", true, false, false, "Keine Module"},
		{"escaped input", "/archive?q=%3Cscript%3E", false, false, true, "&lt;script&gt;"},
		{"multiple PO filters", "/archive?program=Informatik&version=2023&version=2017", false, false, true, "3 Module gefunden"},
		{"legacy all PO", "/archive?version=", false, false, true, "6 Module gefunden"},
		{"degree filter", "/archive?degree=Bachelor&program=Informatik&version=2023", false, false, true, "2 Module gefunden"},
		{"master fixture empty", "/archive?degree=Master", false, false, true, "Keine Module gefunden"},
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
			if tc.path == "/archive?q=SQL" && strings.Contains(body, "Mathematik I") {
				t.Fatal("unmatched exam leaked into results")
			}
		})
	}
}
