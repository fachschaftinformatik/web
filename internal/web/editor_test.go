package web

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/forum"
)

func TestForumEditor(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/d/new", "/d/203823111852785665/edit", "/d/140127153315381249/edit"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
			if response.Code != 200 || response.Header().Get("Cache-Control") != "no-store" {
				t.Fatal(response)
			}
			if strings.Contains(response.Body.String(), "Beitragsart") || strings.Contains(response.Body.String(), `name="type"`) {
				t.Fatal("post type selector remains in editor")
			}
			if strings.Contains(path, "203823") && !strings.Contains(response.Body.String(), `&#34;Beispiel.pdf&#34;`) {
				t.Fatal("body not decoded before editing")
			}
			if strings.Contains(path, "140127") && !strings.Contains(response.Body.String(), `value="30.01.2026"`) {
				t.Fatal("event date missing")
			}
			values := url.Values{"title": {"Mein Titel"}, "body": {"<script>Kein HTML</script>"}, "type": {"event"}, "date": {"30.01.2027"}, "time": {"17:00"}, "tags": {"LaTeX, Studium"}, "location": {"Raum A4"}}
			for _, valid := range []bool{false, true} {
				if !valid {
					values.Set("title", "")
				} else {
					values.Set("title", "Mein Titel")
				}
				request := httptest.NewRequest("POST", path, strings.NewReader(values.Encode()))
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				response = httptest.NewRecorder()
				handler.ServeHTTP(response, request)
				want := 422
				if valid {
					want = 503
				}
				body := response.Body.String()
				if response.Code != want || !strings.Contains(body, "&lt;script&gt;Kein HTML&lt;/script&gt;") {
					t.Fatalf("%d %s", response.Code, body)
				}
				if strings.Contains(path, "140127") && !strings.Contains(body, `value="Raum A4"`) {
					t.Fatal("event location was not retained")
				}
				if valid && !strings.Contains(body, "Dein Beitrag wurde nicht gespeichert.") {
					t.Fatal("missing unavailable explanation")
				}
			}
		})
	}
	thread, _ := forum.Lookup("203823111852785665")
	if thread.Title != "LaTeX Template für Bachelor-/Masterarbeit" {
		t.Fatal("source was mutated")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/d/unknown/edit", nil))
	if response.Code != 404 {
		t.Fatal(response.Code)
	}
	request := httptest.NewRequest("POST", "/d/new", strings.NewReader(strings.Repeat("x", 129<<10)))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatal(response.Code)
	}
}

func TestRichEditorSubmission(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		html   string
		status int
	}{
		{`<p><strong>Hallo Studierende</strong></p><ul><li>Ein Punkt</li></ul><a href="https://example.org">Quelle</a><script>bad()</script>`, 503},
		{`<p>&nbsp;</p><ul><li><br></li></ul>`, 422},
		{`<p><strong>kurz</strong></p>`, 422},
	} {
		values := url.Values{"title": {"Mein Beitrag"}, "rich_body": {tc.html}}
		request := httptest.NewRequest("POST", "/d/new", strings.NewReader(values.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		body := response.Body.String()
		if response.Code != tc.status || strings.Contains(body, "bad()") {
			t.Fatalf("%d %s", response.Code, body)
		}
		if tc.status == 503 && !strings.Contains(body, "<strong>Hallo Studierende</strong>") {
			t.Fatal("formatted preview missing")
		}
	}
}
