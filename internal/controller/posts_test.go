package controller

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/view"
)

func TestPostsFiltering(t *testing.T) {
	handler, err := testHandler(t)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		query, contains, absent string
	}{
		{"", "3 Beiträge", "govuk-pagination"},
		{"?type=discussion", "3 Beiträge", "Projektvorstellungen, Projektstände"},
		{"?search=Neovim", "LaTeX Template", "Wie läuft das Bachelor-Kolloquium"},
		{"?search=%3Cscript%3E", "&lt;script&gt;", "<script>LaTeX"},
		{"?type=unknown", "3 Beiträge", "Keine passenden Beiträge gefunden"},
	} {
		t.Run(tc.query, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", "/posts"+tc.query, nil))
			body := response.Body.String()
			if response.Code != 200 || !htmlContains(body, tc.contains) || htmlContains(body, tc.absent) {
				t.Fatalf("unexpected response: %d %s", response.Code, body)
			}
			if !htmlContains(body, `href="/posts" aria-current="page"`) {
				t.Fatal("forum missing from active Community navigation")
			}
		})
	}
	for _, kind := range []string{"news", "event"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", "/posts?type="+kind, nil))
		if w.Code != 200 || w.Header().Get("Location") != "" {
			t.Fatalf("unsupported list filter %s: %d %s", kind, w.Code, w.Header().Get("Location"))
		}
	}
}

func TestDiscussion(t *testing.T) {
	handler, db := testApplication(t)
	seedPublicContent(t, db)
	latex := publicRef(t, db, "posts", "203823111852785665")
	quote := publicRef(t, db, "posts", "147381306706100225")
	news := publicRef(t, db, "posts", "140128508453060609")
	event := publicRef(t, db, "posts", "140127153315381249")
	for _, tc := range []struct {
		path, contains string
		code           int
	}{
		{"/posts/" + latex + "?search=latex", `href="/posts" aria-current="true"`, 200},
		{"/posts/" + latex, `href="https://github.com/LEMMIIX/latex-ba-template"`, 200},
		{"/posts/" + quote, `<blockquote`, 200},
		{"/posts/" + news, `govuk-list--bullet`, 200},
		{"/posts/" + event, `30.01.2026, 17:00 Uhr`, 200},
		{"/posts/203823111852785665", `Seite nicht gefunden`, 404},
		{"/posts/unknown", `Seite nicht gefunden`, 404},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", tc.path, nil))
			body := response.Body.String()
			if response.Code != tc.code || !htmlContains(body, tc.contains) {
				t.Fatalf("%d: %s", response.Code, body)
			}
			ancestor := "/posts"
			if strings.Contains(tc.path, news) {
				ancestor = view.NewsURL
			}
			if strings.Contains(tc.path, event) {
				ancestor = view.EventsURL
			}
			if tc.code == 200 && !htmlContains(body, `href="`+ancestor+`" aria-current="true"`) {
				t.Fatal("missing active forum ancestor")
			}
		})
	}
}
