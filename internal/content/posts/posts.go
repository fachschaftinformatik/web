// Package posts manages posts, comments, drafts and rich text.
package posts

import (
	_ "embed"
	"encoding/json"
	"slices"
	"strings"
	"time"
)

//go:embed snapshot.json
var snapshot []byte

type Post struct {
	AuthorID                                        int64
	ID, Title, Body, Summary, Author, Created, Type string
	BodyFormat                                      string
	Comments                                        int
	Votes, UserVote                                 int
	Tags                                            []string
}

func (p Post) URL() string { return "/d/" + p.ID }

func (p Post) SourceURL() string { return "https://ik.fsv-wh.de/d/" + p.ID }

func (p Post) DateISO() string {
	if len(p.Created) < 10 {
		return ""
	}
	return p.Created[:10]
}

func (p Post) DateLabel() string {
	return dateLabel(p.Created)
}

func dateLabel(value string) string {
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return ""
	}
	months := [...]string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"}
	return date.Format("2.") + " " + months[date.Month()-1] + " " + date.Format("2006")
}

func (p Post) TypeLabel() string {
	switch p.Type {
	case "news":
		return "Neuigkeit"
	case "event":
		return "Veranstaltung"
	default:
		return "Diskussion"
	}
}

type Filter struct {
	Query, Type string
}

// All returns a fresh copy, ordered by original publication date, newest first.
func All() []Post {
	var posts []Post
	if err := json.Unmarshal(snapshot, &posts); err != nil {
		panic("invalid embedded forum snapshot: " + err.Error())
	}
	slices.SortStableFunc(posts, func(a, b Post) int {
		return strings.Compare(b.Created, a.Created)
	})
	return posts
}

// Find applies all search terms and the type filter to the entire snapshot.
func Find(filter Filter) []Post {
	var result []Post
	terms := strings.Fields(strings.ToLower(filter.Query))
	for _, post := range All() {
		if filter.Type != "" && post.Type != filter.Type {
			continue
		}
		text := strings.ToLower(strings.Join([]string{post.Title, post.Body, post.Summary, strings.Join(post.Tags, " ")}, " "))
		if slices.ContainsFunc(terms, func(term string) bool { return !strings.Contains(text, term) }) {
			continue
		}
		result = append(result, post)
	}
	return result
}
