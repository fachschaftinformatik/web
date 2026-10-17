package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestForumFiltering(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		query, contains, absent string
	}{
		{"", "3 Beiträge", "govuk-pagination"},
		{"?type=discussion", "3 Beiträge", "Projektvorstellungen, Projektstände"},
		{"?q=Neovim", "LaTeX Template", "Wie läuft das Bachelor-Kolloquium"},
		{"?q=%3Cscript%3E", "&lt;script&gt;", "<script>LaTeX"},
		{"?type=unknown", "3 Beiträge", "Keine passenden Beiträge gefunden"},
	} {
		t.Run(tc.query, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", "/discussions"+tc.query, nil))
			body := response.Body.String()
			if response.Code != 200 || !strings.Contains(body, tc.contains) || strings.Contains(body, tc.absent) {
				t.Fatalf("unexpected response: %d %s", response.Code, body)
			}
			if !strings.Contains(body, `href="/discussions" aria-current="page"`) {
				t.Fatal("forum missing from active Community navigation")
			}
		})
	}
	for _, tc := range []struct{ kind, path string }{{"news", newsURL}, {"event", eventsURL}} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", "/discussions?type="+tc.kind, nil))
		if w.Code != 303 || w.Header().Get("Location") != tc.path {
			t.Fatalf("legacy filter %s: %d %s", tc.kind, w.Code, w.Header().Get("Location"))
		}
	}
}

func TestDiscussion(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path, contains string
		code           int
	}{
		{"/d/203823111852785665?q=latex&type=discussion", `href="/discussions" aria-current="true"`, 200},
		{"/d/203823111852785665", `href="https://github.com/LEMMIIX/latex-ba-template"`, 200},
		{"/d/147381306706100225", `<blockquote`, 200},
		{"/d/140128508453060609", `govuk-list--bullet`, 200},
		{"/d/140127153315381249", `30. Januar 2026, 17:00 Uhr`, 200},
		{"/d/unknown", `Seite nicht gefunden`, 404},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", tc.path, nil))
			body := response.Body.String()
			if response.Code != tc.code || !strings.Contains(body, tc.contains) {
				t.Fatalf("%d: %s", response.Code, body)
			}
			ancestor := "/discussions"
			if strings.Contains(tc.path, "140128508453060609") {
				ancestor = newsURL
			}
			if strings.Contains(tc.path, "140127153315381249") {
				ancestor = eventsURL
			}
			if tc.code == 200 && !strings.Contains(body, `href="`+ancestor+`" aria-current="true"`) {
				t.Fatal("missing active forum ancestor")
			}
		})
	}
}

func TestForumTextEscapesSourceHTML(t *testing.T) {
	renderer, err := newRenderer()
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	err = renderer.pages["thread"].ExecuteTemplate(&out, "forum-body", forumText(`&lt;script&gt;alert(1)&lt;/script&gt; https://example.org/?a=1&b=2. javascript:alert(1)`))
	if err != nil {
		t.Fatal(err)
	}
	body := out.String()
	if strings.Contains(body, "<script>") || strings.Contains(body, `href="javascript:`) || !strings.Contains(body, `href="https://example.org/?a=1&amp;b=2"`) {
		t.Fatal(body)
	}
}
