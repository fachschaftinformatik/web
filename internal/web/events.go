package web

import (
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/content/events"
	"github.com/fachschaftinformatik/web/internal/content/posts"
)

const eventsURL = "/community/veranstaltungen"
const eventsPageSize = 6

type EventsPage struct {
	Pagination     ListPagination
	Upcoming, Past []events.Event
	Detail         *events.Event
	CanCreate      bool
}

func (c *Handler) Events(w http.ResponseWriter, r *http.Request) {
	page := EventsPage{}
	if c.events != nil {
		page.CanCreate = c.events.CanPublish(r.Context(), forumUser(w)) == nil
		if id := r.PathValue("event"); id != "" {
			event, err := c.events.Lookup(r.Context(), id)
			if err != nil {
				c.eventError(w, r, err)
				return
			}
			page.Detail = &event
			c.render(w, "events", false, Page{Title: event.Title, Active: "calendar", Community: true, Events: page})
			return
		}
		records, total, err := c.events.FindPage(r.Context(), time.Now(), requestedPage(r), eventsPageSize, false)
		if err != nil {
			c.serverError(w, err)
			return
		}
		page.Upcoming, page.Past = groupEvents(records, time.Now())
		page.Pagination = listPagination(r, total, eventsPageSize)
	} else if r.PathValue("event") != "" {
		http.NotFound(w, r)
		return
	}
	c.render(w, "events", false, Page{Title: "Veranstaltungen", Active: "calendar", Community: true, Events: page})
}

func groupEvents(records []events.Event, now time.Time) (upcoming, past []events.Event) {
	if zone, err := time.LoadLocation("Europe/Berlin"); err == nil {
		now = now.In(zone)
	}
	for _, event := range records {
		date := event.SortDate()
		if event.SourceStatus == "cancelled" || len(date) >= 10 && date[:10] < now.Format("2006-01-02") {
			past = append(past, event)
		} else {
			upcoming = append(upcoming, event)
		}
	}
	sort.SliceStable(upcoming, func(i, j int) bool {
		a, b := upcoming[i].SortDate(), upcoming[j].SortDate()
		if a == "" {
			return false
		}
		if b == "" {
			return true
		}
		return a < b
	})
	sort.SliceStable(past, func(i, j int) bool { return past[i].SortDate() > past[j].SortDate() })
	return upcoming, past
}

func (c *Handler) EditEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if c.events == nil {
		http.NotFound(w, r)
		return
	}
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
	editor := ForumEditor{Live: true, Event: true, Action: eventsURL + "/new", BackURL: eventsURL, Draft: posts.Draft{Type: "event"}}
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
		editor.Draft = posts.Draft{Title: event.Title, Body: event.Body, Type: "event", Location: event.Location}
		editor.RichBody = event.RichBody()
		if d, err := time.Parse("2006-01-02T15:04", event.EventDate); err == nil {
			editor.Draft.Date, editor.Draft.Time = d.Format("02.01.2006"), d.Format("15:04")
		} else if d, err := time.Parse("2006-01-02", event.EventDate); err == nil {
			editor.Draft.Date = d.Format("02.01.2006")
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
			http.Error(w, "Das Formular konnte nicht gelesen werden. Das Titelbild darf höchstens 6 MB groß sein.", http.StatusBadRequest)
			return
		}
		editor.Draft = posts.Draft{Type: "event", Title: r.PostForm.Get("title"), Body: r.PostForm.Get("body"), Date: r.PostForm.Get("date"), Time: r.PostForm.Get("time"), Location: r.PostForm.Get("location")}
		editor.RichBody = posts.PlainRichText(editor.Draft.Body)
		if r.PostForm.Has("rich_body") {
			editor.RichBody = posts.SanitizeRichText(r.PostForm.Get("rich_body"))
			editor.Draft.Body = posts.RichTextVisible(editor.RichBody)
		}
		editor.CoverAlt = r.PostForm.Get("cover_alt")
		cover := events.CoverUpload{Alt: editor.CoverAlt, Remove: r.PostForm.Get("remove_cover") == "1"}
		file, _, err := r.FormFile("cover")
		if err == nil {
			cover.Data, err = io.ReadAll(io.LimitReader(file, events.MaxCoverBytes+1))
			file.Close()
		}
		if err != nil && !errors.Is(err, http.ErrMissingFile) && !errors.Is(err, http.ErrNotMultipart) {
			editor.Errors = append(editor.Errors, posts.FieldError{Field: "cover", Message: "Das Titelbild konnte nicht gelesen werden."})
		}
		if len(editor.Errors) == 0 {
			event, err := c.events.Save(r.Context(), user, id, events.Draft{Title: editor.Draft.Title, Body: editor.Draft.Body, Date: editor.Draft.Date, Time: editor.Draft.Time, Location: editor.Draft.Location}, editor.RichBody, &cover)
			if err == nil {
				http.Redirect(w, r, eventsURL+"/"+event, http.StatusSeeOther)
				return
			}
			var validation *events.ValidationError
			if errors.As(err, &validation) {
				editor.Errors = validation.Fields
			} else if errors.Is(err, events.ErrCoverStorage) {
				editor.Unavailable = true
				status = http.StatusServiceUnavailable
			} else {
				c.eventError(w, r, err)
				return
			}
		}
		if len(editor.Errors) > 0 {
			status = http.StatusUnprocessableEntity
		}
	}
	c.renderStatus(w, "editor", false, Page{Title: title, Active: "calendar", Community: true, Editor: &editor}, status)
}

func (c *Handler) CalendarCover(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if c.events == nil {
		http.NotFound(w, r)
		return
	}
	object, err := c.events.OpenCover(r.Context(), r.PathValue("event"))
	if err != nil {
		c.eventError(w, r, err)
		return
	}
	c.serveGalleryObject(w, r, object)
}

func (c *Handler) eventError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, events.ErrNotFound) {
		http.NotFound(w, r)
	} else if errors.Is(err, events.ErrCoverStorage) {
		http.Error(w, "Das Titelbild konnte nicht gespeichert oder geladen werden.", http.StatusServiceUnavailable)
	} else {
		c.forumError(w, r, err)
	}
}
