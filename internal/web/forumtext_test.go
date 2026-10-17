package web

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/content/posts"
	xhtml "golang.org/x/net/html"
)

func TestLegacyForumTextDecoding(t *testing.T) {
	renderer, err := newRenderer()
	if err != nil {
		t.Fatal(err)
	}
	const source = "Beispiel &lt;script&gt; und &amp; https://example.org/?a=1&amp;b=2"
	const decoded = "Beispiel <script> und & https://example.org/?a=1&b=2"
	thread := posts.Thread{Post: posts.Post{ID: "1", Body: source}, Replies: []posts.Comment{{ID: "2", Text: source}}}
	var output bytes.Buffer
	if err := renderer.Render(&output, "thread", false, Page{Forum: ForumPage{Thread: &thread}}); err != nil {
		t.Fatal(err)
	}
	doc, err := xhtml.Parse(strings.NewReader(output.String()))
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range []*xhtml.Node{
		forumTestElement(t, doc, "article", "aria-label", "Beitrag"),
		forumTestElement(t, doc, "li", "id", "comment-2"),
	} {
		paragraph := forumTestElement(t, node, "p", "class", "govuk-body")
		if got := forumTestText(paragraph); got != decoded {
			t.Fatalf("legacy text = %q, want %q", got, decoded)
		}
		forumTestElement(t, paragraph, "a", "href", "https://example.org/?a=1&b=2")
	}
	if strings.Contains(output.String(), "<script> und") {
		t.Fatal("legacy text became executable HTML")
	}
}
