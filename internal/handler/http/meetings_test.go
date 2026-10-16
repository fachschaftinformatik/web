package web

import (
	"crypto/sha256"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/meetings"
)

func TestMeetingDocuments(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	files, sources := map[string]bool{}, map[string]bool{}
	for _, meeting := range meetings.All() {
		for _, document := range meeting.Documents {
			if files[document.File] {
				t.Fatalf("duplicate document URL: %s", document.File)
			}
			files[document.File] = true
			for _, source := range document.Sources {
				if sources[source] {
					t.Fatalf("duplicate source mapping: %s", source)
				}
				sources[source] = true
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", document.URL(), nil))
			content := response.Body.Bytes()
			if response.Code != 200 || !strings.HasPrefix(response.Header().Get("Content-Type"), "application/pdf") {
				t.Fatalf("PDF not served: %s (%d)", document.File, response.Code)
			}
			if len(content) != document.Bytes || fmt.Sprintf("%x", sha256.Sum256(content)) != document.SHA256 {
				t.Fatalf("source integrity mismatch: %s", document.File)
			}
		}
	}
	if len(files) != 48 || len(sources) != 49 {
		t.Fatalf("import coverage: %d files, %d sources", len(files), len(sources))
	}
}

func TestMeetingYears(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		query, want string
		status      int
	}{
		{"", "24. September 2026", 200},
		{"?jahr=2024", "10. Dezember 2024", 200},
		{"?jahr=2025", "Datum uneindeutig", 200},
		{"?jahr=2022", "2022", 200},
		{"?jahr=2020", "6. März 2020", 200},
		{"?jahr=2021", "22. Januar 2021", 200},
		{"?jahr=2023", "18. Januar 2023", 200},
		{"?jahr=1999", "404", 404},
		{"?jahr=invalid", "404", 404},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/fachschaft/sitzungen"+tc.query, nil))
		if response.Code != tc.status || !strings.Contains(response.Body.String(), tc.want) {
			t.Fatalf("%s: %d %s", tc.query, response.Code, response.Body.String())
		}
	}
}
