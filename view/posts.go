package view

import (
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/urls"
)

type Post struct {
	AuthorPublicID                                     string
	AuthorID                                           int64
	AuthorPublic                                       bool
	ID, Title, Body, Author, Created, Type, BodyFormat string
	Comments, Votes, UserVote                          int
	Tags                                               []string
}
type Comment struct {
	AuthorPublicID                     string
	AuthorPublic                       bool
	Votes, UserVote                    int
	AuthorID                           int64
	CanEdit                            bool
	ID, Author, Text, Created, Updated string
}
type Thread struct {
	Post
	Updated, EventDate, Location string
	Replies                      []Comment
}
type PostsFilter struct{ Query string }

func (p Post) URL() string                            { return urls.Post(p.ID) }
func (p Post) EditURL() string                        { return urls.EditPost(p.ID) }
func (p Post) VoteURL() string                        { return urls.PostVote(p.ID) }
func (p Post) CommentsURL() string                    { return urls.PostComments(p.ID) }
func (t Thread) CommentVoteURL(comment string) string { return urls.CommentVote(t.ID, comment) }
func (t Thread) EditCommentURL(comment string) string { return urls.EditComment(t.ID, comment) }

// Public profiles are currently disabled.
func (p Post) AuthorURL() string    { return "" }
func (c Comment) AuthorURL() string { return "" }
func (p Post) PostedLabel() string  { return postedLabel(p.Created) }
func (p Post) DateISO() string {
	if len(p.Created) < 10 {
		return ""
	}
	return p.Created[:10]
}
func (p Post) DateLabel() string { return dateLabel(p.Created) }
func (p Post) TypeLabel() string {
	switch p.Type {
	case "news":
		return "Ankündigung"
	case "event":
		return "Veranstaltung"
	default:
		return "Diskussion"
	}
}
func (c Comment) DateLabel() string    { return dateLabel(c.Created) }
func (c Comment) UpdatedLabel() string { return dateLabel(c.Updated) }
func (t Thread) UpdatedLabel() string  { return dateLabel(t.Updated) }
func (t Thread) EventLabel() string    { return eventLabel(t.EventDate) }
func (p Post) RelativeTime() string    { return relativeTime(p.Created, time.Now()) }
func (c Comment) RelativeTime() string { return relativeTime(c.Created, time.Now()) }
func (c Comment) PostedLabel() string  { return postedLabel(c.Created) }
func (t Thread) EditedLabel() string   { return postedLabel(t.Updated) }
func (c Comment) EditedLabel() string  { return postedLabel(c.Updated) }
func (t Thread) Edited() bool          { return t.Updated != "" && t.Updated != t.Created }
func (c Comment) Edited() bool         { return c.Updated != "" && c.Updated != c.Created }
func (p Post) Summary() string         { return summary(p.Body, p.BodyFormat) }
func summary(body, format string) string {
	if format == "html" {
		body = content.RichTextVisible(body)
	}
	body = strings.Join(strings.Fields(body), " ")
	if r := []rune(body); len(r) > 240 {
		end := 240
		if r[end] != ' ' {
			for i := end - 1; i > 0; i-- {
				if r[i] == ' ' {
					end = i
					break
				}
			}
		}
		return strings.TrimSpace(string(r[:end])) + "…"
	}
	return body
}
