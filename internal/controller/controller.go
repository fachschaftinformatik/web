package controller

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/fachschaftinformatik/web/internal/model/archive"
	"github.com/fachschaftinformatik/web/internal/model/events"
	"github.com/fachschaftinformatik/web/internal/model/forum"
	"github.com/fachschaftinformatik/web/internal/model/gallery"
	"github.com/fachschaftinformatik/web/internal/model/meetings"
	"github.com/fachschaftinformatik/web/internal/model/people"
	"github.com/fachschaftinformatik/web/internal/model/users"
	"github.com/fachschaftinformatik/web/internal/view"
)

func (c *Handler) adminRender(w http.ResponseWriter, title string, a *view.AdminPage, status int) {
	a.CanResendConfirmation = a.User != nil && !a.User.Verified
	a.ReviewerOnly = !w.(*sessionResponse).session.CanAdmin()
	p := view.Page{Title: title, Active: "admin", Admin: a}
	if a.View == "users" && a.Pages > 1 {
		p.DocumentTitle = fmt.Sprintf("Benutzerkonten – Seite %d von %d", a.Page, a.Pages)
	}
	if a.HasErrors() {
		p.DocumentTitle = "Fehler: " + title
	}
	c.renderStatus(w, "admin", false, p, status)
}

func (c *Handler) notFound(w http.ResponseWriter, _ *http.Request) {
	c.pageError(w, "Prüfe die Adresse. Wenn du einem Link auf unserer Website gefolgt bist, gib uns bitte Bescheid.", http.StatusNotFound)
}

func (c *Handler) pageError(w http.ResponseWriter, message string, status int) {
	title := "Die Anfrage konnte nicht verarbeitet werden"
	switch status {
	case http.StatusNotFound:
		title = "Seite nicht gefunden"
	case http.StatusForbidden:
		title = "Zugriff nicht möglich"
	case http.StatusInternalServerError, http.StatusServiceUnavailable:
		title = "Der Dienst ist gerade nicht verfügbar"
	}
	w.Header().Set("Cache-Control", "no-store")
	c.renderStatus(w, "error", false, view.Page{Title: title, Active: "error", Message: message}, status)
}

type Handler struct {
	router        http.Handler
	views         *view.Renderer
	people        *people.Store
	meetings      *meetings.Store
	identity      *users.Service
	secureCookies bool
	archive       *archive.Service
	forum         *forum.Service
	gallery       *gallery.Service
	events        *events.Service
}

type Services struct {
	Users    *users.Service
	Archive  *archive.Service
	People   *people.Store
	Meetings *meetings.Store
	Forum    *forum.Service
	Gallery  *gallery.Service
	Events   *events.Service
}

// New constructs the HTML application with its required dependencies.
func New(services Services, secureCookies bool) (*Handler, error) {
	if services.Users == nil || services.Archive == nil || services.People == nil || services.Meetings == nil || services.Forum == nil || services.Gallery == nil || services.Events == nil {
		return nil, errors.New("application services are incomplete")
	}
	views, err := view.NewRenderer()
	if err != nil {
		return nil, err
	}
	c := &Handler{views: views, secureCookies: secureCookies,
		identity: services.Users, archive: services.Archive, people: services.People,
		meetings: services.Meetings, forum: services.Forum, gallery: services.Gallery, events: services.Events}
	c.router = c.routes()
	return c, nil
}

func (c *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.router.ServeHTTP(w, r)
}

func (c *Handler) render(w http.ResponseWriter, page string, fragment bool, data view.Page) {
	c.renderStatus(w, page, fragment, data, http.StatusOK)
}

func (c *Handler) serverError(w http.ResponseWriter, err error) {
	log.Printf("Request failed: %v", err)
	c.pageError(w, "Versuche es später erneut. Deine Anfrage konnte nicht abgeschlossen werden.", http.StatusInternalServerError)
}

func (c *Handler) renderStatus(w http.ResponseWriter, page string, fragment bool, data view.Page, status int) {
	if status == http.StatusUnprocessableEntity {
		if data.DocumentTitle == "" {
			data.DocumentTitle = data.Title
		}
		if !strings.HasPrefix(data.DocumentTitle, "Fehler: ") {
			data.DocumentTitle = "Fehler: " + data.DocumentTitle
		}
	}
	if response, ok := w.(*sessionResponse); ok && response.session != nil {
		data.SignedIn, data.CSRF = response.session.Authenticated, response.session.CSRF
		data.CanAdmin = response.session.CanReview()
		data.UserID, data.Role = response.session.UserID, response.session.Role
	}
	var html bytes.Buffer
	if err := c.views.Render(&html, page, fragment, data); err != nil {
		log.Printf("Render %s: %v", page, err)
		http.Error(w, "Die Seite konnte nicht geladen werden.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = html.WriteTo(w)
}

func ptr[T any](value T) *T { return &value }
