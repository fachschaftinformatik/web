package site

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestContentEditor(t *testing.T) {
	handler, db := testApplication(t)
	server := httptest.NewServer(handler)
	defer server.Close()
	b := newAdminBrowser(t, server)
	b.login()
	status, body, _ := b.get("/posts/create")
	if status != 200 || strings.Contains(body, "Beitragsart") || strings.Contains(body, `name="type"`) {
		t.Fatal("editor exposes a post-type selector")
	}
	form := url.Values{"csrf": {b.csrf}, "title": {""}, "body": {"Ein vollständiger Beitragstext."}, "type": {"event"}, "location": {"Raum A4"}}
	status, body, _ = b.post("/posts/create", form)
	if status != 422 || !strings.Contains(body, "Ein vollständiger Beitragstext.") {
		t.Fatalf("invalid form lost submitted text: %d %s", status, body)
	}
	form.Set("title", "Mein Beitrag")
	status, _, headers := b.post("/posts/create", form)
	if status != 303 {
		t.Fatalf("valid post failed: %d", status)
	}
	var kind string
	if err := db.QueryRow("SELECT type FROM discussion_posts WHERE title='Mein Beitrag'").Scan(&kind); err != nil || kind != "discussion" {
		t.Fatalf("client changed post kind: %s %v", kind, err)
	}
	path := headers.Get("Location")
	if status, body, _ = b.get(path + "/edit"); status != 200 || !strings.Contains(body, "Mein Beitrag") {
		t.Fatal("saved post cannot be edited")
	}
	if status, _, _ = b.get("/posts/unknown/edit"); status != 404 {
		t.Fatal("unknown post editable")
	}
	if status, _, _ = b.request("POST", "/posts/create", "application/x-www-form-urlencoded", strings.NewReader(strings.Repeat("x", 9<<20))); status != 400 {
		t.Fatalf("oversized request: %d", status)
	}
}

func TestRichEditorSubmission(t *testing.T) {
	handler, _ := testApplication(t)
	server := httptest.NewServer(handler)
	defer server.Close()
	b := newAdminBrowser(t, server)
	b.login()
	for _, tc := range []struct {
		html   string
		status int
	}{
		{`<p><strong>Hallo Studierende</strong></p><ul><li>Ein Punkt</li></ul><a href="https://example.org">Quelle</a><script>bad()</script>`, 303},
		{`<p>&nbsp;</p><ul><li><br></li></ul>`, 422},
		{`<p><strong>kurz</strong></p>`, 422},
	} {
		status, body, headers := b.post("/posts/create", url.Values{"csrf": {b.csrf}, "title": {"Mein Beitrag"}, "rich_body": {tc.html}})
		if status != tc.status || strings.Contains(body, "bad()") {
			t.Fatalf("%d %s", status, body)
		}
		if status == 303 {
			_, body, _ = b.get(headers.Get("Location"))
			if !strings.Contains(body, "<strong>Hallo Studierende</strong>") || strings.Contains(body, "bad()") {
				t.Fatal("stored rich text was not sanitized")
			}
		}
	}
}
