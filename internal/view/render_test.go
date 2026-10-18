package view

import (
	"strings"
	"testing"
)

func TestPostTextEscapesSourceHTML(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	err = renderer.pages["thread"].ExecuteTemplate(&out, "post-body", rawPostText(`<script>alert(1)</script> https://example.org/?a=1&b=2. javascript:alert(1)`))
	if err != nil {
		t.Fatal(err)
	}
	body := out.String()
	if strings.Contains(body, "<script>") || strings.Contains(body, `href="javascript:`) || !strings.Contains(body, `href="https://example.org/?a=1&amp;b=2"`) {
		t.Fatal(body)
	}
}
