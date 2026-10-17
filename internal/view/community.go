package view

import (
	"html/template"

	"github.com/fachschaftinformatik/web/internal/content"

	"github.com/fachschaftinformatik/web/internal/model/forum"
	"github.com/fachschaftinformatik/web/internal/model/gallery"
)

type MeetingsPage struct {
	Year     int
	Years    []int
	Meetings []Meeting
	Next     *Meeting
}

type ContentEditor struct {
	Draft                EditorDraft
	Errors               []content.FieldError
	Editing, Unavailable bool
	Action, BackURL      string
	RichBody             string
	Event, Announcement  bool
	CoverURL, CoverAlt   string
}

func (e ContentEditor) RichHTML() template.HTML {
	return template.HTML(content.SanitizeRichText(e.RichBody))
}

func (e ContentEditor) Error(field string) string {
	for _, err := range e.Errors {
		if err.Field == field {
			return err.Message
		}
	}
	return ""
}

type EventsPage struct {
	Pagination     ListPagination
	Upcoming, Past []Event
	Detail         *Event
	CanCreate      bool
}

type ForumPage struct {
	Pagination        ListPagination
	CanWrite, CanEdit bool
	Filter            forum.Filter
	Posts             []Post
	Thread            *Thread
	BackURL           string
	Comment           CommentForm
}

type CommentForm struct {
	ID, Text, Error string
}

func (f CommentForm) Anchor() string {
	if f.ID != "" {
		return "edit-comment-" + f.ID
	}
	return "comment-text"
}

func (p ForumPage) CommentURL(id string) string {
	base := p.Thread.URL()
	if p.Pagination.Page > 1 {
		base = p.Pagination.PageURL(p.Pagination.Page)
	}
	return base + "#comment-" + id
}

type HomePost struct {
	Title, Summary, URL, Date, DateISO string
	Votes, UserVote                    int
}

func HomeItem(post forum.Post) HomePost {
	p := NewPost(post)
	return HomePost{Title: p.Title, Summary: p.Summary(), URL: p.URL(), Date: p.DateLabel(), DateISO: p.DateISO(), Votes: p.Votes, UserVote: p.UserVote}
}

type HomePage struct {
	News   HomePost
	Forum  []HomePost
	Events []Event
}

type ImagesPage struct {
	Pagination                 ListPagination
	Albums                     []Album
	Album                      Album
	Images                     []Image
	Image                      Image
	Detail, Viewer             bool
	Number, Total, Page, Pages int
	PageNumbers                []int
	Previous, Next, Return     string
}

func (g ImagesPage) ImageNumber(index int) int {
	return (g.Page-1)*gallery.PageSize + index + 1
}

func (g ImagesPage) CarouselAlbums() []Album {
	var albums []Album
	for _, album := range g.Albums {
		if album.CoverImage().ID != "" {
			albums = append(albums, album)
		}
	}
	return albums
}

type VoteControl struct {
	Page                  Page
	Action, Return, Label string
	Score, Selected       int
	Up, Down              int
}

func voteControl(page Page, action, back, label string, score, selected int) VoteControl {
	v := VoteControl{Page: page, Action: action, Return: back, Label: label, Score: score, Selected: selected, Up: 1, Down: -1}
	if selected == 1 {
		v.Up = 0
	}
	if selected == -1 {
		v.Down = 0
	}
	return v
}

type EditorDraft struct{ Title, Body, Type, Tags, Date, Time, Location string }
