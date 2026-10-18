package view

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedBrowserAssets(t *testing.T) {
	assets := Assets()
	for _, path := range []string{
		"govuk.css", "govuk.js", "htmx.js", "site.css", "site.js", "editor.css", "editor.js",
		"fonts/bold-b542beb274-v2.woff2", "fonts/light-94a07e06a1-v2.woff2",
		"tinymce/tinymce.min.js", "tinymce/langs8/de.js", "tinymce/themes/silver/theme.min.js",
		"tinymce/icons/default/icons.min.js", "tinymce/models/dom/model.min.js", "tinymce/plugins/lists/plugin.min.js",
		"tinymce/skins/ui/oxide/skin.min.css", "tinymce/skins/ui/oxide/content.min.css",
		"tinymce/LICENSE.txt", "tinymce/I18N-LICENSE.md",
		"icons/bullet-list.svg", "icons/ordered-list.svg", "icons/link.svg", "icons/undo.svg", "icons/redo.svg",
		"community/discuss-and-give-feedback.svg", "community/share-your-examples-and-research.svg",
		"community/propose-a-change-to-pages.svg", "community/research-interview.png",
		"community/catchup-groupshot.png", "community/workshop.png", "event-placeholder.svg",
	} {
		if data, err := fs.ReadFile(assets, path); err != nil || len(data) == 0 {
			t.Errorf("missing embedded browser asset %s: %v", path, err)
		}
	}
}

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
