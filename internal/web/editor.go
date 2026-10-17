package web

import (
	"errors"
	"html"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/content/posts"
)

type ForumEditor struct {
	Live                       bool
	Draft                      posts.Draft
	Errors                     []posts.FieldError
	Editing, Unavailable       bool
	Action, BackURL, SourceURL string
	RichBody                   string
	Event, Announcement        bool
	CoverURL, CoverAlt         string
}

func (e ForumEditor) RichHTML() template.HTML {
	return template.HTML(posts.SanitizeRichText(e.RichBody))
}

func (e ForumEditor) Error(field string) string {
	for _, err := range e.Errors {
		if err.Field == field {
			return err.Message
		}
	}
	return ""
}

func (c *Handler) EditDiscussion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	user := forumUser(w)
	if c.posts != nil {
		if user == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if err := c.posts.CanWrite(r.Context(), user, 0, false); err != nil {
			c.forumError(w, r, err)
			return
		}
	}
	editor := ForumEditor{Action: "/d/new", BackURL: "/discussions", SourceURL: "https://ik.fsv-wh.de/d/new", Draft: posts.Draft{Type: "discussion"}}
	editor.Live = c.posts != nil
	title := "Beitrag verfassen"
	if r.URL.Path == newsURL+"/new" {
		editor.Announcement = true
		editor.Draft.Type = "news"
		editor.Action = newsURL + "/new"
		editor.BackURL = newsURL
		title = "Neuigkeit veröffentlichen"
	}
	if id := r.PathValue("post"); id != "" {
		var thread posts.Thread
		var ok bool
		if c.posts != nil {
			var err error
			thread, err = c.posts.Lookup(r.Context(), id)
			if err != nil {
				c.forumError(w, r, err)
				return
			}
			ok = true
			if err = c.posts.CanWrite(r.Context(), user, thread.AuthorID, false); err != nil {
				c.forumError(w, r, err)
				return
			}
		} else {
			thread, ok = posts.Lookup(id)
		}
		if !ok {
			http.NotFound(w, r)
			return
		}
		editor.Editing = true
		if c.events != nil && thread.Type == "event" {
			if event, err := c.events.Lookup(r.Context(), id); err == nil {
				http.Redirect(w, r, event.URL()+"/edit", http.StatusSeeOther)
				return
			}
		}
		editor.Announcement = thread.Type == "news"
		editor.Event = thread.Type == "event"
		editor.Action, editor.BackURL, editor.SourceURL = thread.URL()+"/edit", thread.URL(), thread.SourceURL()+"/edit"
		body := thread.Body
		if c.posts == nil {
			body = html.UnescapeString(body)
		}
		editor.Draft = posts.Draft{Title: thread.Title, Body: body, Type: thread.Type, Tags: strings.Join(thread.Tags, ", "), Location: thread.Location}
		if thread.BodyFormat == "html" {
			editor.RichBody = posts.SanitizeRichText(thread.Body)
			editor.Draft.Body = posts.RichTextPlain(editor.RichBody)
		}
		if date, err := time.Parse("2006-01-02T15:04", thread.EventDate); err == nil {
			editor.Draft.Date, editor.Draft.Time = date.Format("02.01.2006"), date.Format("15:04")
		}
		title = "Beitrag bearbeiten"
		if editor.Event {
			title = "Veranstaltung bearbeiten"
		}
		if editor.Announcement {
			title = "Neuigkeit bearbeiten"
		}
	}
	if c.posts != nil && (editor.Event || editor.Announcement) {
		if err := c.posts.CanPublish(r.Context(), user); err != nil {
			c.forumError(w, r, err)
			return
		}
	}
	status := http.StatusOK
	if editor.RichBody == "" {
		editor.RichBody = posts.PlainRichText(editor.Draft.Body)
	}
	if id := r.PathValue("post"); c.posts == nil && id != "" {
		if thread, ok := posts.Lookup(id); ok && thread.BodyFormat == "html" {
			editor.RichBody = posts.SanitizeRichText(thread.Body)
			editor.Draft.Body = posts.RichTextPlain(editor.RichBody)
		}
	}
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 128<<10)
		parseErr := r.ParseForm()
		if parseErr != nil {
			http.Error(w, "Das Formular konnte nicht gelesen werden. Versuche es mit einem kürzeren Beitrag erneut.", http.StatusBadRequest)
			return
		}
		kind, tags := editor.Draft.Type, editor.Draft.Tags
		if c.posts != nil {
			tags = r.PostForm.Get("tags")
		}
		editor.Draft = posts.Draft{Title: r.PostForm.Get("title"), Body: r.PostForm.Get("body"), Type: kind, Tags: tags, Date: r.PostForm.Get("date"), Time: r.PostForm.Get("time"), Location: r.PostForm.Get("location")}
		if r.PostForm.Has("rich_body") {
			editor.RichBody = posts.SanitizeRichText(r.PostForm.Get("rich_body"))
			editor.Draft.Body = posts.RichTextPlain(editor.RichBody)
		} else {
			editor.RichBody = posts.PlainRichText(editor.Draft.Body)
		}
		validation := editor.Draft
		if r.PostForm.Has("rich_body") {
			validation.Body = posts.RichTextVisible(editor.RichBody)
		}
		editor.Errors = validation.Validate()
		if len(editor.Errors) > 0 {
			status = http.StatusUnprocessableEntity
		} else if c.posts != nil {
			post, err := c.posts.Save(r.Context(), user, r.PathValue("post"), editor.Draft, editor.RichBody)
			if err == nil {
				http.Redirect(w, r, "/d/"+post, http.StatusSeeOther)
				return
			}
			var validation *posts.ValidationError
			if errors.As(err, &validation) {
				editor.Errors = validation.Fields
				status = http.StatusUnprocessableEntity
			} else {
				c.forumError(w, r, err)
				return
			}
		} else {
			editor.Unavailable = true
			status = http.StatusServiceUnavailable
		}
	}
	data := Page{Title: title, Active: "forum", Community: true, Editor: &editor}
	if editor.Event {
		data.Active = "calendar"
	}
	if editor.Announcement {
		data.Active = "news"
	}
	if len(editor.Errors) > 0 {
		data.DocumentTitle = "Fehler: " + title
	}
	if editor.Unavailable {
		data.DocumentTitle = "Fehler: " + title
	}
	c.renderStatus(w, "editor", false, data, status)
}
