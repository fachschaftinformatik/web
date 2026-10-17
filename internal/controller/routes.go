package controller

import (
	"net/http"
	"strings"

	"github.com/fachschaftinformatik/web/internal/view"
)

func (c *Handler) routes() http.Handler {
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
		albums, _, err := c.gallery.AlbumPage(r.Context(), 1, 6)
		if err != nil {
			c.serverError(w, err)
			return
		}

		c.render(w, "home", false, view.Page{Title: "Startseite", Active: "home", Home: home, Images: view.ImagesPage{Albums: albums}})
	})
	router.HandleFunc("GET /team", c.Team)
	router.HandleFunc("GET /sitemap", c.Sitemap)
	router.HandleFunc("GET /impressum", c.Legal)
	router.HandleFunc("GET /datenschutz", c.Legal)
	router.HandleFunc("GET /barrierefreiheit", c.Legal)
	router.HandleFunc("GET /contact", func(w http.ResponseWriter, r *http.Request) {
		c.render(w, "contact", false, view.Page{Title: "Kontakt", Active: "contact"})
	})
	router.HandleFunc("GET /fachschaft", func(w http.ResponseWriter, r *http.Request) {
		c.render(w, "fachschaft", false, view.Page{Title: "Fachschaft", Active: "fachschaft", Fachschaft: true})
	})
	router.HandleFunc("GET /fachschaft/sitzungen", c.Meetings)
	router.HandleFunc("GET /fachschaft/wahlen", func(w http.ResponseWriter, r *http.Request) {
		c.render(w, "elections", false, view.Page{Title: "Wahlen", Active: "elections", Fachschaft: true})
	})
	router.HandleFunc("GET /fachschaft/ordnungen", func(w http.ResponseWriter, r *http.Request) {
		c.render(w, "regulations", false, view.Page{Title: "Satzung und Ordnungen", Active: "regulations", Fachschaft: true})
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
		c.render(w, "community", false, view.Page{Title: "Community", Active: "community", Community: true})
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
	router.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServerFS(view.Assets())))
	router.HandleFunc("/", c.notFound)
	return c.withSession(router)
}
