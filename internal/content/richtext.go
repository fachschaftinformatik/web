package content

import (
	"bytes"
	"html"
	"io"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	markdownhtml "github.com/yuin/goldmark/renderer/html"
	xhtml "golang.org/x/net/html"
)

var richPolicy = func() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "h2", "h3", "br", "strong", "b", "em", "i", "ul", "ol", "li", "blockquote")
	p.AllowAttrs("href").OnElements("a")
	p.AllowURLSchemes("https", "http", "mailto")
	p.RequireParseableURLs(true)
	p.RequireNoFollowOnLinks(true)
	return p
}()

// SanitizeRichText is the only boundary for user-authored rich HTML.
func SanitizeRichText(source string) string { return richPolicy.Sanitize(source) }

// RichTextPlain produces an editable plain-text fallback and validation text.
func RichTextPlain(source string) string {
	return richTextPlain(source, true)
}

// RichTextVisible excludes link destinations when checking visible text length.
func RichTextVisible(source string) string { return richTextPlain(source, false) }

func richTextPlain(source string, includeLinks bool) string {
	var out strings.Builder
	link, linkStart := "", 0
	tokens := xhtml.NewTokenizer(strings.NewReader(SanitizeRichText(source)))
	for {
		switch tokens.Next() {
		case xhtml.ErrorToken:
			if tokens.Err() == io.EOF {
				return strings.TrimSpace(out.String())
			}
			return strings.TrimSpace(out.String())
		case xhtml.TextToken:
			out.Write(tokens.Text())
		case xhtml.StartTagToken, xhtml.SelfClosingTagToken:
			tag := tokens.Token()
			name := tag.Data
			if name == "a" {
				for _, attr := range tag.Attr {
					if attr.Key == "href" {
						link, linkStart = attr.Val, out.Len()
					}
				}
			}
			switch name {
			case "br", "p", "h2", "h3", "blockquote", "ul", "ol", "li":
				out.WriteString("\n")
			}
		case xhtml.EndTagToken:
			name, _ := tokens.TagName()
			if string(name) == "a" && link != "" && includeLinks {
				if out.String()[linkStart:] != link {
					out.WriteString(" (" + link + ")")
				}
				link = ""
			}
			switch string(name) {
			case "p", "h2", "h3", "blockquote", "li":
				out.WriteString("\n")
			}
		}
	}
}

// PlainRichText treats literal HTML as text when enhancing a plain textarea.
func PlainRichText(source string) string {
	return "<p>" + strings.ReplaceAll(html.EscapeString(strings.ReplaceAll(source, "\r\n", "\n")), "\n", "<br>") + "</p>"
}

var markdown = goldmark.New(goldmark.WithRendererOptions(markdownhtml.WithHardWraps()))

// MarkdownRichText renders text-based event descriptions without changing their
// original source. Raw HTML is disabled; generated HTML uses the same allowlist
// as descriptions authored in the rich-text editor.
func MarkdownRichText(source string) string {
	var out bytes.Buffer
	if err := markdown.Convert([]byte(source), &out); err != nil {
		return PlainRichText(source)
	}
	return SanitizeRichText(out.String())
}
