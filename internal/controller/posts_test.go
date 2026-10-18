package controller

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/view"
)

func TestAnnouncementAndForumSearchPagination(t *testing.T) {
	handler, db := testApplication(t)
	seedPublicContent(t, db)
	for kindIndex, kind := range []string{"news", "discussion"} {
		for i := 0; i < 22; i++ {
			_, err := db.Exec(`INSERT INTO discussion_posts(id,title,body,user_id,type,created_at) VALUES(?,?,?,900,?,'2026-01-01T00:00:00Z')`, 10000+kindIndex*100+i, fmt.Sprintf("Pagination %s %02d", kind, i), "Unique needle content", kind)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	seedPublicRefs(t, db)
	for _, tc := range []struct{ path, title, kind, other string }{
		{"/news", "Ankündigungen", "news", "discussion"},
		{"/posts", "Forum", "discussion", "news"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			for _, page := range []int{1, 2} {
				path := tc.path + "?search=Unique+needle"
				if page > 1 {
					path += "&page=2"
				}
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
				body := w.Body.String()
				for _, want := range []string{"govuk-pagination", fmt.Sprintf("Seite %d von 2", page), tc.title + " – Suchergebnisse", `value="Unique needle"`, "Pagination " + tc.kind, `search=Unique&#43;needle`, `aria-current="page"`} {
					if w.Code != 200 || !htmlContains(body, want) {
						t.Fatalf("%s missing %q: %d %s", path, want, w.Code, body)
					}
				}
				if strings.Contains(body, "Pagination "+tc.other) {
					t.Fatal("search mixed announcements and discussions")
				}
				canonical := tc.path + "?search=Unique&#43;needle"
				if page > 1 {
					canonical = tc.path + "?page=2&amp;search=Unique&#43;needle"
				}
				if !strings.Contains(body, `rel="canonical" href="`+canonical+`"`) {
					t.Fatal("canonical URL lost search or page context")
				}
				if page == 2 && strings.Count(body, ">Pagination "+tc.kind) != 2 {
					t.Fatal("second page did not contain the remaining two matching entries")
				}
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest("GET", tc.path+"?page=2", nil))
			if w.Code != 200 || !strings.Contains(w.Body.String(), "govuk-pagination") {
				t.Fatal("unfiltered list did not paginate")
			}
			w = httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest("GET", tc.path+"?search=nonexistentneedle", nil))
			if w.Code != 200 || strings.Contains(w.Body.String(), `class="govuk-pagination"`) || !strings.Contains(w.Body.String(), "Keine passenden") {
				t.Fatal("empty search did not show an unpaginated empty state")
			}
			for _, query := range []string{"?page=3&search=Unique+needle", "?page=2&search=nonexistentneedle"} {
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, httptest.NewRequest("GET", tc.path+query, nil))
				if w.Code != 404 {
					t.Fatalf("out-of-range page returned %d", w.Code)
				}
			}
		})
	}
}

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

func TestPublicProfilesAreUnavailable(t *testing.T) {
	handler, db := testApplication(t)
	seedPublicContent(t, db)
	id := publicRef(t, db, "users", "900")
	for _, path := range []string{"/users/" + id, "/users/" + id + "/", "/users/" + id + "?page=2"} {
		for _, method := range []string{"GET", "HEAD"} {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			if w.Code != 404 || w.Header().Get("Location") != "" {
				t.Fatalf("%s %s: expected unavailable profile, got %d", method, path, w.Code)
			}
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
