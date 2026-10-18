package content

import (
	"strings"
	"testing"
)

func TestRichTextBoundary(t *testing.T) {
	input := `<h2>Abschnitt</h2><h3>Unterabschnitt</h3><p onclick="bad()">Hallo <strong>Studierende</strong><script>alert(1)</script><img src=x onerror=bad()></p><ul><li>Ein Punkt</li></ul><blockquote>Ein Zitat</blockquote><a href="javascript:alert(1)">Unsicher</a><a href="https://example.org">Quelle</a>`
	clean := SanitizeRichText(input)
	for _, forbidden := range []string{"<script", "onclick", "onerror", "<img", "javascript:"} {
		if strings.Contains(clean, forbidden) {
			t.Fatal(clean)
		}
	}
	for _, allowed := range []string{"<h2>Abschnitt</h2><h3>Unterabschnitt</h3>", "<strong>Studierende</strong>", "<ul><li>Ein Punkt</li></ul>", "<blockquote>Ein Zitat</blockquote>", `href="https://example.org"`} {
		if !strings.Contains(clean, allowed) {
			t.Fatal(clean)
		}
	}
	if plain := RichTextPlain(clean); !strings.Contains(plain, "Quelle (https://example.org)") || strings.Contains(plain, "<strong>") {
		t.Fatal(plain)
	}
	if plain := RichTextPlain(`<p>&nbsp;</p><ul><li><br></li></ul>`); strings.TrimSpace(plain) != "" {
		t.Fatal(plain)
	}
	if safe := PlainRichText(`<script>literal text</script>`); strings.Contains(safe, "<script>") || !strings.Contains(safe, "&lt;script&gt;") {
		t.Fatal(safe)
	}
	if got := RichTextVisible(`<h2>Abschnitt</h2><h3>Unterabschnitt</h3><p>Text</p>`); !strings.Contains(got, "Abschnitt\n") || !strings.Contains(got, "Unterabschnitt\n") {
		t.Fatal(got)
	}
	if got := SanitizeRichText(`<h1>Kein zweiter Seitentitel</h1><h2 style="color:red" onclick="bad()">Abschnitt</h2>`); strings.Contains(got, "<h1") || strings.Contains(got, "style=") || strings.Contains(got, "onclick=") {
		t.Fatal(got)
	}
}

func TestEventMarkdownFormattingAndBoundary(t *testing.T) {
	source := "Am **22.04.2026 ab 16:00 Uhr** laden wir euch ein.\n\n**16:00 Uhr:** Vortrag von Prof. Beispiel\n**16:30 Uhr:** Zweiter Vortrag\n\n- Austausch\n- *Gemeinsames Grillen*\n\n[Weitere Informationen](https://example.org/event)"
	rendered := MarkdownRichText(source)
	for _, want := range []string{"<strong>22.04.2026 ab 16:00 Uhr</strong>", "<strong>16:00 Uhr:</strong>", "<strong>16:30 Uhr:</strong>", "<br>", "<ul>", "<li>Austausch</li>", "<em>Gemeinsames Grillen</em>", `href="https://example.org/event"`} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("event formatting missing %q: %s", want, rendered)
		}
	}
	unsafe := MarkdownRichText("[Link](javascript:alert%281%29)\n\n<script>alert(1)</script>\n\n<img src=x onerror=alert(1)>")
	for _, forbidden := range []string{"<script", "<img", "onerror=", "javascript:"} {
		if strings.Contains(unsafe, forbidden) {
			t.Fatal("markdown bypassed rich-text boundary", unsafe)
		}
	}
}
