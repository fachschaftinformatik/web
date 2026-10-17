package web

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/content/posts"
	"github.com/fachschaftinformatik/web/internal/data"
	xhtml "golang.org/x/net/html"
)

func TestPersistentForumRequests(t *testing.T) {
	path := filepath.Join(t.TempDir(), "forum.db")
	if err := data.Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	db, err := data.NewConnection(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`INSERT INTO users(id,email,name,password,active,verified) VALUES(1,'author@fsv-wh.de','Echter Autor','unused',1,1),(2,'other@fsv-wh.de','Andere Person','unused',1,1)`)
	if err != nil {
		t.Fatal(err)
	}
	views, err := newRenderer()
	if err != nil {
		t.Fatal(err)
	}
	c := &Handler{views: views, posts: posts.NewService(db)}
	request := func(method, path string, form url.Values, user int64, handle http.HandlerFunc, post, comment string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.SetPathValue("post", post)
		r.SetPathValue("comment", comment)
		w := httptest.NewRecorder()
		response := &sessionResponse{ResponseWriter: w}
		if user != 0 {
			response.session = &session{Authenticated: true, UserID: user, CSRF: "forum-token"}
		}
		handle(response, r)
		return w
	}
	if w := request("POST", "/d/new", nil, 0, c.EditDiscussion, "", ""); w.Code != 303 || w.Header().Get("Location") != "/login" {
		t.Fatal(w)
	}
	form := url.Values{"title": {"Ein echter Beitrag"}, "body": {"Ein ausreichender Text."}, "type": {"discussion"}, "rich_body": {`<p><strong>Ein ausreichender Text.</strong><script>bad()</script></p>`}}
	w := request("POST", "/d/new", form, 1, c.EditDiscussion, "", "")
	if w.Code != 303 || !strings.HasPrefix(w.Header().Get("Location"), "/d/") {
		t.Fatal(w)
	}
	post := strings.TrimPrefix(w.Header().Get("Location"), "/d/")
	thread, err := c.posts.Lookup(context.Background(), post)
	if err != nil || thread.Author != "Echter Autor" || strings.Contains(thread.Body, "bad()") {
		t.Fatalf("%+v %v", thread, err)
	}
	if w = request("GET", "/d/"+post+"/edit", nil, 2, c.EditDiscussion, post, ""); w.Code != 403 {
		t.Fatal(w)
	}
	if w = request("POST", "/d/"+post+"/edit", form, 2, c.EditDiscussion, post, ""); w.Code != 403 {
		t.Fatal(w)
	}
	if w = request("GET", "/d/"+post+"/edit", nil, 1, c.EditDiscussion, post, ""); w.Code != 200 || !strings.Contains(w.Body.String(), `name="csrf" value="forum-token"`) {
		t.Fatal(w)
	}
	form.Set("title", "")
	if w = request("POST", "/d/"+post+"/edit", form, 1, c.EditDiscussion, post, ""); w.Code != 422 {
		t.Fatal(w)
	}
	if w = request("POST", "/d/"+post+"/comments", url.Values{"text": {""}}, 1, c.Comment, post, ""); w.Code != 422 {
		t.Fatal(w)
	}
	w = request("POST", "/d/"+post+"/comments", url.Values{"text": {"<script>Kein ausführbarer Kommentar</script>"}}, 2, c.Comment, post, "")
	if w.Code != 303 || !strings.Contains(w.Header().Get("Location"), "#comment-") {
		t.Fatal(w)
	}
	w = request("GET", "/d/"+post, nil, 2, c.Discussion, post, "")
	if w.Code != 200 || strings.Contains(w.Body.String(), "<script>Kein") || !strings.Contains(w.Body.String(), "&lt;script&gt;Kein") || !strings.Contains(w.Body.String(), "Kommentar bearbeiten") {
		t.Fatal(w)
	}
	list, err := c.posts.Find(context.Background(), posts.Filter{})
	if err != nil || len(list) != 1 || list[0].Comments != 1 {
		t.Fatalf("%+v %v", list, err)
	}
	for _, text := range []string{
		"Literal &lt;script&gt; und &amp; bleiben erhalten.",
		"https://example.org/?value=&lt;script&gt;&amp;next=1",
		"<script>Kein HTML</script> javascript:alert(1)",
	} {
		t.Run(text, func(t *testing.T) {
			post, err := c.posts.Save(context.Background(), 1, "", posts.Draft{Title: "Klartext prüfen", Body: text, Type: "discussion"}, "")
			if err != nil {
				t.Fatal(err)
			}
			comment, err := c.posts.SaveComment(context.Background(), 1, post, "", text)
			if err != nil {
				t.Fatal(err)
			}
			w := request("GET", "/d/"+post, nil, 1, c.Discussion, post, "")
			if w.Code != 200 {
				t.Fatal(w)
			}
			doc, err := xhtml.Parse(strings.NewReader(w.Body.String()))
			if err != nil {
				t.Fatal(err)
			}
			commentNode := forumTestElement(t, doc, "li", "id", "comment-"+comment)
			shownComment := forumTestElement(t, commentNode, "p", "class", "govuk-body")
			editComment := forumTestElement(t, commentNode, "textarea", "name", "text")
			postNode := forumTestElement(t, doc, "article", "aria-label", "Beitrag")
			shownPost := forumTestElement(t, postNode, "p", "class", "govuk-body")
			for _, node := range []*xhtml.Node{shownComment, editComment, shownPost} {
				if got := forumTestText(node); got != text {
					t.Fatalf("%s text = %q, want %q", node.Data, got, text)
				}
			}
			if strings.Contains(w.Body.String(), "<script>Kein HTML") {
				t.Fatal("literal text became executable HTML")
			}
			if strings.HasPrefix(text, "https://") {
				for _, node := range []*xhtml.Node{shownComment, shownPost} {
					forumTestElement(t, node, "a", "href", text)
				}
			}
			w = request("GET", "/d/"+post+"/edit", nil, 1, c.EditDiscussion, post, "")
			if w.Code != 200 {
				t.Fatal(w)
			}
			doc, err = xhtml.Parse(strings.NewReader(w.Body.String()))
			if err != nil {
				t.Fatal(err)
			}
			if got := forumTestText(forumTestElement(t, doc, "textarea", "name", "body")); got != text {
				t.Fatalf("post edit text = %q, want %q", got, text)
			}
		})
	}
}

func forumTestElement(t *testing.T, root *xhtml.Node, tag, key, value string) *xhtml.Node {
	t.Helper()
	var find func(*xhtml.Node) *xhtml.Node
	find = func(node *xhtml.Node) *xhtml.Node {
		if node.Type == xhtml.ElementNode && node.Data == tag {
			for _, attr := range node.Attr {
				if attr.Key == key && attr.Val == value {
					return node
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if found := find(child); found != nil {
				return found
			}
		}
		return nil
	}
	if found := find(root); found != nil {
		return found
	}
	t.Fatalf("missing %s with %s=%q", tag, key, value)
	return nil
}

func forumTestText(node *xhtml.Node) string {
	if node.Type == xhtml.TextNode {
		return node.Data
	}
	var text strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		text.WriteString(forumTestText(child))
	}
	return text.String()
}
