// Package web serves HTML pages using the archive and team domain packages.
package web

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/content/events"
	"github.com/fachschaftinformatik/web/internal/content/files"
	"github.com/fachschaftinformatik/web/internal/content/media"
	"github.com/fachschaftinformatik/web/internal/content/posts"
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
	archive       *files.Service
	posts         *posts.Service
	media         *media.Service
	events        *events.Service
}

type Services struct {
	Users    *users.Service
	Archive  *files.Service
	People   *people.Store
	Meetings *content.Store
	Posts    *posts.Service
	Media    *media.Service
	Events   *events.Service
}

func New() (http.Handler, error) {
	return NewWithUsers(users.Examples())
}

func NewWithUsers(records []users.User) (http.Handler, error) {
	return newHandler(records, nil, false)
}

// NewPersistent uses database-backed identities instead of demo credentials.
func NewPersistent(services Services, secureCookies bool) (http.Handler, error) {
	if services.Users == nil || services.Archive == nil || services.People == nil || services.Meetings == nil || services.Posts == nil || services.Media == nil {
		return nil, errors.New("persistent web services are incomplete")
	}
	return newHandler(nil, &services, secureCookies)
}

func newHandler(records []users.User, services *Services, secureCookies bool) (http.Handler, error) {
	views, err := newRenderer()
	if err != nil {
		return nil, err
	}
	c := &Handler{views: views, users: users.NewDirectory(records), sessions: newSessionStore(), secureCookies: secureCookies}
	if services == nil {
		c.people, c.content = people.NewStore(), content.NewStore()
	} else {
		c.identity, c.archive, c.people, c.content, c.posts, c.media = services.Users, services.Archive, services.People, services.Meetings, services.Posts, services.Media
		c.events = services.Events
	}
	router := http.NewServeMux()
	router.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	router.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		home, err := c.home(r.Context(), forumUser(w))
		if err != nil {
			c.serverError(w, err)
			return
		}
		albums := media.Albums()
		if c.media != nil {
			var err error
			albums, _, err = c.media.AlbumPage(r.Context(), 1, 6)
			if err != nil {
				c.serverError(w, err)
				return
			}
		}
		c.render(w, "home", false, Page{Title: "Startseite", Active: "home", Home: home, Images: ImagesPage{Albums: albums}})
	})
	router.HandleFunc("GET /team", c.Team)
	router.HandleFunc("GET /sitemap", c.Sitemap)
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
	router.HandleFunc("GET /archive/new", c.UploadArchive)
	router.HandleFunc("POST /archive/new", c.UploadArchive)
	router.HandleFunc("GET /archive/preferences", c.ArchivePreferences)
	router.HandleFunc("POST /archive/preferences", c.ArchivePreferences)
	router.HandleFunc("GET /archive/files/{id}/revision", c.UploadArchive)
	router.HandleFunc("POST /archive/files/{id}/revision", c.UploadArchive)
	router.HandleFunc("GET /archive/requests", c.ArchiveRequests)
	router.HandleFunc("GET /archive/requests/{request}", c.ArchiveRequest)
	router.HandleFunc("GET /archive/requests/{request}/review", c.ArchiveRequest)
	router.HandleFunc("POST /archive/requests/{request}/review", c.ArchiveRequest)
	router.HandleFunc("GET /archive/requests/{request}/edit", c.UploadArchive)
	router.HandleFunc("POST /archive/requests/{request}/edit", c.UploadArchive)
	router.HandleFunc("GET /archive/files/{id}", c.ArchiveFile)
	router.HandleFunc("GET /api/v1/archive/{id}/file", c.ArchiveFile)
	router.HandleFunc("GET /archive/{module}/{entry}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/archive/"+r.PathValue("module"), http.StatusSeeOther)
	})
	router.HandleFunc("GET /discussions", c.Forum)
	router.HandleFunc("GET /community/veranstaltungen", c.Events)
	router.HandleFunc("GET /community/veranstaltungen/new", c.EditEvent)
	router.HandleFunc("POST /community/veranstaltungen/new", c.EditEvent)
	router.HandleFunc("GET /community/veranstaltungen/{event}", c.Events)
	router.HandleFunc("GET /community/veranstaltungen/{event}/edit", c.EditEvent)
	router.HandleFunc("POST /community/veranstaltungen/{event}/edit", c.EditEvent)
	router.HandleFunc("GET /community/veranstaltungen/{event}/cover", c.CalendarCover)
	router.HandleFunc("GET /community/neuigkeiten", c.News)
	router.HandleFunc("GET /community/neuigkeiten/new", c.EditDiscussion)
	router.HandleFunc("POST /community/neuigkeiten/new", c.EditDiscussion)
	router.HandleFunc("GET /d/{post}", c.Discussion)
	router.HandleFunc("GET /users/{id}", c.PublicProfile)
	router.HandleFunc("GET /d/new", c.EditDiscussion)
	router.HandleFunc("POST /d/new", c.EditDiscussion)
	router.HandleFunc("GET /d/{post}/edit", c.EditDiscussion)
	router.HandleFunc("POST /d/{post}/edit", c.EditDiscussion)
	router.HandleFunc("POST /d/{post}/comments", c.Comment)
	router.HandleFunc("POST /d/{post}/comments/{comment}/edit", c.Comment)
	router.HandleFunc("POST /d/{post}/vote", c.Vote)
	router.HandleFunc("POST /d/{post}/comments/{comment}/vote", c.Vote)
	router.HandleFunc("GET /media/{id}/{size}", c.MediaFile)
	router.HandleFunc("GET /media/events/{id}/cover/{size}", c.EventCover)
	router.HandleFunc("GET /login", c.Login)
	router.HandleFunc("POST /login", c.Login)
	router.HandleFunc("POST /logout", c.Logout)
	router.HandleFunc("GET /verwaltung", c.Admin)
	router.HandleFunc("GET /verwaltung/{section}", c.Admin)
	router.HandleFunc("GET /verwaltung/{section}/{id}", c.Admin)
	router.HandleFunc("POST /verwaltung/{section}/{id}", c.Admin)
	router.HandleFunc("POST /verwaltung/mitglieder/{id}/einladen", c.ResendInvitation)
	router.HandleFunc("GET /verwaltung/antraege/{request}", c.ReviewArchiveRequest)
	router.HandleFunc("POST /verwaltung/antraege/{request}", c.ReviewArchiveRequest)
	router.HandleFunc("GET /verwaltung/einreichungen/{request}", c.ReviewArchiveRequest)
	router.HandleFunc("POST /verwaltung/einreichungen/{request}", c.ReviewArchiveRequest)
	router.HandleFunc("GET /verwaltung/einreichungen/{request}/edit", c.UploadArchive)
	router.HandleFunc("POST /verwaltung/einreichungen/{request}/edit", c.UploadArchive)
	router.HandleFunc("GET /verwaltung/sitzungen/{id}/pruefen", c.ReviewPublication)
	router.HandleFunc("POST /verwaltung/sitzungen/{id}/pruefen", c.ReviewPublication)
	router.HandleFunc("GET /verwaltung/benutzer/{id}/passwort-zuruecksetzen", c.ResetUserPassword)
	router.HandleFunc("POST /verwaltung/benutzer/{id}/passwort-zuruecksetzen", c.ResetUserPassword)
	router.HandleFunc("POST /verwaltung/benutzer/{id}/bestaetigung-senden", c.ResendUserConfirmation)
	router.HandleFunc("GET /documents/{id}", c.Document)
	router.HandleFunc("GET /register", c.Register)
	router.HandleFunc("POST /register", c.Register)
	router.HandleFunc("GET /forgot", c.ForgotPassword)
	router.HandleFunc("POST /forgot", c.ForgotPassword)
	router.HandleFunc("GET /verify", c.Token)
	router.HandleFunc("POST /verify", c.Token)
	router.HandleFunc("GET /reset", c.Token)
	router.HandleFunc("POST /reset", c.Token)
	router.HandleFunc("GET /invite", c.Invitation)
	router.HandleFunc("POST /invite", c.Invitation)
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
	if err := c.people.Refresh(r.Context()); err != nil {
		c.serverError(w, err)
		return
	}
	c.render(w, "team", false, Page{Title: "Mitglieder", Active: "team", Fachschaft: true, Team: c.people.Groups()})
}

func archiveFilter(r *http.Request) files.Filter {
	q := r.URL.Query()
	return files.Filter{Query: q.Get("q"), Programs: q["program"], Versions: slices.DeleteFunc(q["version"], func(version string) bool { return version == "" }), Degrees: q["degree"]}
}

func (c *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	filter, err := c.preferredArchiveFilter(w, r)
	if err != nil {
		c.serverError(w, err)
		return
	}
	var catalog files.Catalog
	var total int
	pageNumber, _ := strconv.Atoi(r.URL.Query().Get("seite"))
	pageNumber = max(1, pageNumber)
	if c.archive != nil {
		catalog, total, err = c.archive.CatalogPage(r.Context(), filter, pageNumber, archivePageSize)
	} else {
		catalog = files.Examples()
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	programs := make([]FilterOption, 0)
	for _, program := range catalog.Programs() {
		programs = append(programs, FilterOption{Name: program, Selected: slices.Contains(filter.Programs, program)})
	}
	versions := make([]FilterOption, 0)
	for _, version := range catalog.Versions() {
		versions = append(versions, FilterOption{Name: version, Selected: slices.Contains(filter.Versions, version)})
	}
	degrees := make([]FilterOption, 0)
	for _, degree := range catalog.Degrees() {
		degrees = append(degrees, FilterOption{Name: degree, Selected: slices.Contains(filter.Degrees, degree)})
	}
	data := Page{
		Title: "Altklausuren", Active: "archive",
		Archive: ArchivePage{Filter: filter, Programs: programs, Versions: versions, Degrees: degrees, Modules: catalog.FindModules(filter)},
	}
	if c.archive != nil {
		data.Archive.Total = total
		data.Archive.Pages = max(1, (total+archivePageSize-1)/archivePageSize)
		data.Archive.Page = min(pageNumber, data.Archive.Pages)
	} else {
		data.Archive.paginate(r.URL.Query().Get("seite"))
	}
	if raw := r.URL.Query().Get("seite"); raw != "" && raw != fmt.Sprint(data.Archive.Page) {
		http.Redirect(w, r, data.Archive.PageURL(1), http.StatusSeeOther)
		return
	}
	if data.Archive.Pages > 1 {
		data.DocumentTitle = fmt.Sprintf("Altklausuren (Seite %d von %d)", data.Archive.Page, data.Archive.Pages)
	}
	w.Header().Set("Vary", "HX-Request, HX-History-Restore-Request")
	fragment := r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-History-Restore-Request") != "true"
	c.render(w, "archive", fragment, data)
}

func (c *Handler) ArchiveModule(w http.ResponseWriter, r *http.Request) {
	filter, err := c.preferredArchiveFilter(w, r)
	if err != nil {
		c.serverError(w, err)
		return
	}
	if c.archive != nil {
		module, total, err := c.archive.ModulePage(r.Context(), r.PathValue("module"), filter, requestedPage(r), 20)
		if errors.Is(err, files.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			c.serverError(w, err)
			return
		}
		c.render(w, "module", false, Page{Title: module.Title, Active: "archive", Archive: ArchivePage{Module: &module, Filter: filter, ListPagination: listPagination(r, total, 20)}})
		return
	}
	catalog := files.Examples()
	module, ok := catalog.FindModule(r.PathValue("module"), filter)
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
	data.Voting = c.posts != nil
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
