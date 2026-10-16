package web

import (
	"html"
	"regexp"
	"strings"
)

type textPart struct{ Text, URL string }
type textBlock struct {
	Kind  string
	Lines [][]textPart
}

var publicURL = regexp.MustCompile(`https?://[^\s<>"']+`)

// forumText preserves source paragraphs, lists and quotes as escaped text.
// Only explicit HTTP(S) URLs become links; source HTML is never trusted.
func forumText(source string) []textBlock {
	var blocks []textBlock
	for _, line := range strings.Split(strings.ReplaceAll(html.UnescapeString(source), "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			blocks = append(blocks, textBlock{})
			continue
		}
		kind := "paragraph"
		if strings.HasPrefix(line, "- ") {
			kind, line = "list", strings.TrimPrefix(line, "- ")
		}
		if strings.HasPrefix(line, "> ") {
			kind, line = "quote", strings.TrimPrefix(line, "> ")
		}
		if len(blocks) == 0 || blocks[len(blocks)-1].Kind != kind {
			blocks = append(blocks, textBlock{Kind: kind})
		}
		index := len(blocks) - 1
		blocks[index].Lines = append(blocks[index].Lines, forumParts(line))
	}
	return blocks
}

func forumParts(line string) []textPart {
	var parts []textPart
	for {
		match := publicURL.FindStringIndex(line)
		if match == nil {
			break
		}
		link := strings.TrimRight(line[match[0]:match[1]], ".,;:!?)]}")
		parts = append(parts, textPart{Text: line[:match[0]]}, textPart{Text: link, URL: link})
		line = line[match[0]+len(link):]
	}
	return append(parts, textPart{Text: line})
}
