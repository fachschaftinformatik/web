package controller

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/content"

	"github.com/fachschaftinformatik/web/internal/model/events"
	"github.com/fachschaftinformatik/web/internal/view"
)

const eventsPageSize = 6

func (c *Handler) Events(w http.ResponseWriter, r *http.Request) {
	page := view.EventsPage{}

	page.CanCreate = c.events.CanPublish(r.Context(), forumUser(w)) == nil
	if id := r.PathValue("event"); id != "" {
		event, err := c.events.Lookup(r.Context(), id)
		if err != nil {
			c.eventError(w, r, err)
			return
		}
		page.Detail = &event
		c.render(w, "events", false, view.Page{Title: event.Title, Active: "calendar", Community: true, Events: page})
		return
	}
	records, total, err := c.events.FindPage(r.Context(), time.Now(), requestedPage(r), eventsPageSize, false)
	if err != nil {
		c.serverError(w, err)
		return
	}
	page.Upcoming, page.Past = events.Group(records, time.Now())
	page.Pagination = listPagination(r, total, eventsPageSize)

	c.render(w, "events", false, view.Page{Title: "Veranstaltungen", Active: "calendar", Community: true, Events: page})
}

func (c *Handler) EditEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	user := forumUser(w)
	if user == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := c.events.CanPublish(r.Context(), user); err != nil {
		c.eventError(w, r, err)
		return
	}
	id := r.PathValue("event")
	editor := view.ContentEditor{Event: true, Action: view.EventsURL + "/new", BackURL: view.EventsURL, Draft: view.EditorDraft{Type: "event"}}
	title := "Veranstaltung erstellen"
	if id != "" {
		event, err := c.events.Lookup(r.Context(), id)
		if err != nil {
			c.eventError(w, r, err)
			return
		}
		editor.Editing = true
		editor.Action = event.URL() + "/edit"
		editor.BackURL = event.URL()
		editor.CoverURL, editor.CoverAlt = event.CoverURL(), event.CoverAlt
		editor.Draft = view.EditorDraft{Title: event.Title, Body: event.Body, Type: "event", Location: event.Location}
		editor.RichBody = event.RichBody()
		if d, err := time.Parse("2006-01-02T15:04", event.EventDate); err == nil {
			editor.Draft.Date, editor.Draft.Time = d.Format("02.01.2006"), d.Format("15:04")
		} else {
			if d, err := time.Parse("2006-01-02", event.EventDate); err == nil {
				editor.Draft.Date = d.Format("02.01.2006")
			}
		}
		title = "Veranstaltung bearbeiten"
	}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
		var err error
		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			err = r.ParseMultipartForm(1 << 20)
			if r.MultipartForm != nil {
				defer r.MultipartForm.RemoveAll()
			}
		} else {
			err = r.ParseForm()
		}
		if err != nil {
			c.pageError(w, "Das Formular konnte nicht gelesen werden. Das Titelbild darf höchstens 6 MB groß sein.", http.StatusBadRequest)
			return
		}
		editor.Draft = view.EditorDraft{Type: "event", Title: r.PostForm.Get("title"), Body: r.PostForm.Get("body"), Date: r.PostForm.Get("date"), Time: r.PostForm.Get("time"), Location: r.PostForm.Get("location")}
		editor.RichBody = content.PlainRichText(editor.Draft.Body)
		if r.PostForm.Has("rich_body") {
			editor.RichBody = content.SanitizeRichText(r.PostForm.Get("rich_body"))
			editor.Draft.Body = content.RichTextVisible(editor.RichBody)
		}
		editor.CoverAlt = r.PostForm.Get("cover_alt")
		cover := events.CoverUpload{Alt: editor.CoverAlt, Remove: r.PostForm.Get("remove_cover") == "1"}
		file, _, err := r.FormFile("cover")
		if err == nil {
			cover.Data, err = io.ReadAll(io.LimitReader(file, events.MaxCoverBytes+1))
			file.Close()
		}
		if err != nil && !errors.Is(err, http.ErrMissingFile) && !errors.Is(err, http.ErrNotMultipart) {
			editor.Errors = append(editor.Errors, content.FieldError{Field: "cover", Message: "Das Titelbild konnte nicht gelesen werden."})
		}
		if len(editor.Errors) == 0 {
			event, err := c.events.Save(r.Context(), user, id, events.Draft{Title: editor.Draft.Title, Body: editor.Draft.Body, Date: editor.Draft.Date, Time: editor.Draft.Time, Location: editor.Draft.Location}, editor.RichBody, &cover)
			if err == nil {
				http.Redirect(w, r, view.EventsURL+"/"+event, http.StatusSeeOther)
				return
			}
			var validation *events.ValidationError
			if errors.As(err, &validation) {
				editor.Errors = validation.Fields
			} else {
				if errors.Is(err, events.ErrCoverStorage) {
					editor.Unavailable = true
					status = http.StatusServiceUnavailable
				} else {
					c.eventError(w, r, err)
					return
				}
			}
		}
		if len(editor.Errors) > 0 {
			status = http.StatusUnprocessableEntity
		}
	}
	c.renderStatus(w, "editor", false, view.Page{Title: title, Active: "calendar", Community: true, Editor: &editor}, status)
}

func (c *Handler) eventError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, events.ErrNotFound) {
		c.notFound(w, r)
	} else {
		if errors.Is(err, events.ErrCoverStorage) {
			c.pageError(w, "Das Titelbild konnte nicht gespeichert oder geladen werden.", http.StatusServiceUnavailable)
		} else {
			c.forumError(w, r, err)
		}
	}
}
