// Package web serves HTML pages using the archive and team domain packages.
package web

import (
	"bytes"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/content/files"
	"github.com/fachschaftinformatik/web/internal/content/media"
	"github.com/fachschaftinformatik/web/internal/people"
	"github.com/fachschaftinformatik/web/internal/users"
)

type Handler struct {
	views         *Renderer
	people        *people.Store
	content       *content.Store
	users         *users.Directory
	sessions      sessionStore
	identity      *users.Service
	secureCookies bool
}

func New() (http.Handler, error) {
	return NewWithUsers(users.Examples())
}

func NewWithUsers(records []users.User) (http.Handler, error) {
	return newHandler(records, nil, false)
}

// NewPersistent uses database-backed identities instead of demo credentials.
func NewPersistent(identity *users.Service, secureCookies bool) (http.Handler, error) {
	return newHandler(nil, identity, secureCookies)
}

func newHandler(records []users.User, identity *users.Service, secureCookies bool) (http.Handler, error) {
	views, err := newRenderer()
	if err != nil {
		return nil, err
	}
	c := &Handler{views: views, people: people.NewStore(), content: content.NewStore(), users: users.NewDirectory(records), sessions: newSessionStore()}
	c.identity = identity
	c.secureCookies = secureCookies
	router := http.NewServeMux()
	router.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	router.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		c.render(w, "home", false, Page{Title: "Startseite", Active: "home", Home: homeContent(), Images: ImagesPage{Albums: media.Albums()}})
	})
	router.HandleFunc("GET /team", c.Team)
	router.HandleFunc("GET /contact", func(w http.ResponseWriter, r *http.Request) {
		c.render(w, "contact", false, Page{Title: "Kontakt", Active: "contact"})
	})
	router.HandleFunc("GET /fachschaft", func(w http.ResponseWriter, r *http.Request) {
		c.render(w, "fachschaft", false, Page{Title: "Fachschaft", Active: "fachschaft", Fachschaft: true})
	})
	router.HandleFunc("GET /fachschaft/sitzungen", c.Meetings)
	router.HandleFunc("GET /fachschaft/wahlen", func(w http.ResponseWriter, r *http.Request) {
		c.render(w, "elections", false, Page{Title: "Wahlen", Active: "elections", Fachschaft: true})
	})
	router.HandleFunc("GET /fachschaft/ordnungen", func(w http.ResponseWriter, r *http.Request) {
		c.render(w, "regulations", false, Page{Title: "Satzung und Ordnungen", Active: "regulations", Fachschaft: true})
	})
	router.HandleFunc("GET /archive", c.Archive)
	router.HandleFunc("GET /archive/{module}", c.ArchiveModule)
	router.HandleFunc("GET /discussions", c.Forum)
	router.HandleFunc("GET /d/{post}", c.Discussion)
	router.HandleFunc("GET /d/new", c.EditDiscussion)
	router.HandleFunc("POST /d/new", c.EditDiscussion)
	router.HandleFunc("GET /d/{post}/edit", c.EditDiscussion)
	router.HandleFunc("POST /d/{post}/edit", c.EditDiscussion)
	router.HandleFunc("GET /login", c.Login)
	router.HandleFunc("POST /login", c.Login)
	router.HandleFunc("POST /logout", c.Logout)
	router.HandleFunc("GET /verwaltung", c.Admin)
	router.HandleFunc("GET /verwaltung/{section}", c.Admin)
	router.HandleFunc("GET /verwaltung/{section}/{id}", c.Admin)
	router.HandleFunc("POST /verwaltung/{section}/{id}", c.Admin)
	router.HandleFunc("GET /verwaltung/sitzungen/{id}/pruefen", c.ReviewPublication)
	router.HandleFunc("POST /verwaltung/sitzungen/{id}/pruefen", c.ReviewPublication)
	router.HandleFunc("GET /verwaltung/benutzer/{id}/passwort-zuruecksetzen", c.ResetUserPassword)
	router.HandleFunc("POST /verwaltung/benutzer/{id}/passwort-zuruecksetzen", c.ResetUserPassword)
	router.HandleFunc("GET /documents/{id}", c.Document)
	router.HandleFunc("GET /register", c.Register)
	router.HandleFunc("POST /register", c.Register)
	router.HandleFunc("GET /forgot", c.ForgotPassword)
	router.HandleFunc("POST /forgot", c.ForgotPassword)
	router.HandleFunc("GET /verify", c.Token)
	router.HandleFunc("POST /verify", c.Token)
	router.HandleFunc("GET /reset", c.Token)
	router.HandleFunc("POST /reset", c.Token)
	router.HandleFunc("GET /community", func(w http.ResponseWriter, r *http.Request) {
		c.render(w, "community", false, Page{Title: "Community", Active: "community", Community: true})
	})
	router.HandleFunc("GET /community/events", c.Images)
	router.HandleFunc("GET /community/events/{album}", c.Images)
	router.HandleFunc("GET /community/events/{album}/{image}", c.Images)
	legacyEvents := func(w http.ResponseWriter, r *http.Request) {
		target := "/community/events" + strings.TrimPrefix(r.URL.Path, "/events")
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusSeeOther)
	}
	router.HandleFunc("GET /events", legacyEvents)
	router.HandleFunc("GET /events/{rest...}", legacyEvents)
	router.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServerFS(Assets())))
	return c.withSession(router), nil
}

func (c *Handler) Team(w http.ResponseWriter, r *http.Request) {
	c.render(w, "team", false, Page{Title: "Mitglieder", Active: "team", Fachschaft: true, Team: c.people.Groups()})
}

func archiveFilter(r *http.Request) files.Filter {
	q := r.URL.Query()
	return files.Filter{Query: q.Get("q"), Programs: q["program"], Versions: slices.DeleteFunc(q["version"], func(version string) bool { return version == "" }), Degrees: q["degree"]}
}

func (c *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	filter := archiveFilter(r)
	programs := make([]FilterOption, 0)
	for _, program := range files.Programs() {
		programs = append(programs, FilterOption{Name: program, Selected: slices.Contains(filter.Programs, program)})
	}
	versions := make([]FilterOption, 0)
	for _, version := range files.Versions() {
		versions = append(versions, FilterOption{Name: version, Selected: slices.Contains(filter.Versions, version)})
	}
	degrees := make([]FilterOption, 0)
	for _, degree := range files.Degrees() {
		degrees = append(degrees, FilterOption{Name: degree, Selected: slices.Contains(filter.Degrees, degree)})
	}
	data := Page{
		Title: "Altklausuren", Active: "archive",
		Archive: ArchivePage{Filter: filter, Programs: programs, Versions: versions, Degrees: degrees, Modules: files.FindModules(filter)},
	}
	w.Header().Set("Vary", "HX-Request, HX-History-Restore-Request")
	fragment := r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-History-Restore-Request") != "true"
	c.render(w, "archive", fragment, data)
}

func (c *Handler) ArchiveModule(w http.ResponseWriter, r *http.Request) {
	filter := archiveFilter(r)
	module, ok := files.FindModule(r.PathValue("module"), filter)
	if !ok {
		http.NotFound(w, r)
		return
	}
	c.render(w, "module", false, Page{
		Title: module.Title, Active: "archive",
		Archive: ArchivePage{Module: &module, Filter: filter},
	})
}

func (c *Handler) render(w http.ResponseWriter, page string, fragment bool, data Page) {
	c.renderStatus(w, page, fragment, data, http.StatusOK)
}

func (c *Handler) serverError(w http.ResponseWriter, err error) {
	log.Printf("Request failed: %v", err)
	http.Error(w, "Die Anfrage konnte nicht verarbeitet werden. Versuche es später erneut.", http.StatusInternalServerError)
}

func (c *Handler) renderStatus(w http.ResponseWriter, page string, fragment bool, data Page, status int) {
	if response, ok := w.(*sessionResponse); ok && response.session != nil {
		data.SignedIn, data.CSRF = response.session.Authenticated, response.session.CSRF
		data.CanAdmin = response.session.CanAdmin()
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
