package view

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/model/forum"
	xhtml "golang.org/x/net/html"
)

func TestForumLiteralTextEscaping(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	const source = "Beispiel <script> und & https://example.org/?a=1&b=2"
	const decoded = "Beispiel <script> und & https://example.org/?a=1&b=2"
	thread := forum.Thread{Post: forum.Post{ID: "1", Body: source}, Replies: []forum.Comment{{ID: "2", Text: source}}}
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
