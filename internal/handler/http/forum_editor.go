package web

import (
	"html"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/forum"
)

type ForumEditor struct {
	Draft                      forum.Draft
	Errors                     []forum.FieldError
	Editing, Unavailable       bool
	Action, BackURL, SourceURL string
	RichBody                   string
}

func (e ForumEditor) RichHTML() template.HTML {
	return template.HTML(forum.SanitizeRichText(e.RichBody))
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
	editor := ForumEditor{Action: "/d/new", BackURL: "/discussions", SourceURL: "https://ik.fsv-wh.de/d/new", Draft: forum.Draft{Type: "discussion"}}
	title := "Beitrag verfassen"
	if id := r.PathValue("post"); id != "" {
		thread, ok := forum.Lookup(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		editor.Editing = true
		editor.Action, editor.BackURL, editor.SourceURL = thread.URL()+"/edit", thread.URL(), thread.SourceURL()+"/edit"
		editor.Draft = forum.Draft{Title: thread.Title, Body: html.UnescapeString(thread.Body), Type: thread.Type, Tags: strings.Join(thread.Tags, ", "), Location: thread.Location}
		if date, err := time.Parse("2006-01-02T15:04", thread.EventDate); err == nil {
			editor.Draft.Date, editor.Draft.Time = date.Format("02.01.2006"), date.Format("15:04")
		}
		title = "Beitrag bearbeiten"
	}
	status := http.StatusOK
	editor.RichBody = forum.PlainRichText(editor.Draft.Body)
	if id := r.PathValue("post"); id != "" {
		if thread, ok := forum.Lookup(id); ok && thread.BodyFormat == "html" {
			editor.RichBody = forum.SanitizeRichText(thread.Body)
			editor.Draft.Body = forum.RichTextPlain(editor.RichBody)
		}
	}
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 128<<10)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Das Formular konnte nicht gelesen werden. Versuche es mit einem kürzeren Beitrag erneut.", http.StatusBadRequest)
			return
		}
		editor.Draft = forum.Draft{Title: r.PostForm.Get("title"), Body: r.PostForm.Get("body"), Type: editor.Draft.Type, Tags: editor.Draft.Tags, Date: r.PostForm.Get("date"), Time: r.PostForm.Get("time"), Location: r.PostForm.Get("location")}
		if r.PostForm.Has("rich_body") {
			editor.RichBody = forum.SanitizeRichText(r.PostForm.Get("rich_body"))
			editor.Draft.Body = forum.RichTextPlain(editor.RichBody)
		} else {
			editor.RichBody = forum.PlainRichText(editor.Draft.Body)
		}
		validation := editor.Draft
		if r.PostForm.Has("rich_body") {
			validation.Body = forum.RichTextVisible(editor.RichBody)
		}
		editor.Errors = validation.Validate()
		if len(editor.Errors) > 0 {
			status = http.StatusUnprocessableEntity
		} else {
			editor.Unavailable = true
			status = http.StatusServiceUnavailable
		}
	}
	data := Page{Title: title, Active: "forum", Community: true, Editor: &editor}
	if len(editor.Errors) > 0 {
		data.DocumentTitle = "Fehler: " + title
	}
	if editor.Unavailable {
		data.DocumentTitle = "Speichern derzeit nicht verfügbar – " + title
	}
	c.renderStatus(w, "forum-editor", false, data, status)
}
