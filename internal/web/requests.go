package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/fachschaftinformatik/web/internal/content/files"
)

type RequestListPage struct {
	Requests     []files.Request
	Reviewing    bool
	CanViewUsers bool
}

type RequestPage struct {
	Entry     files.Request
	Reviewing bool
	Error     string
	EditURL   string
}

func (c *Handler) ArchiveRequests(w http.ResponseWriter, r *http.Request) {
	if c.archive == nil {
		c.notFound(w, r)
		return
	}
	requests, err := c.archive.Requests(r.Context(), w.(*sessionResponse).session.UserID, false)
	if err != nil {
		c.archiveWriteError(w, r, err)
		return
	}
	c.render(w, "requests", false, Page{Title: "Meine Einreichungen", Active: "archive", RequestList: RequestListPage{Requests: requests}})
}

func (c *Handler) ArchiveRequest(w http.ResponseWriter, r *http.Request) {
	if c.archive == nil {
		c.notFound(w, r)
		return
	}
	userID := w.(*sessionResponse).session.UserID
	request, err := c.archive.Request(r.Context(), userID, r.PathValue("request"), false)
	if err != nil {
		c.archiveWriteError(w, r, err)
		return
	}
	if r.Method == http.MethodPost {
		if err := c.archive.SendRequest(r.Context(), userID, request.ID); err != nil {
			c.archiveWriteError(w, r, err)
			return
		}
		http.Redirect(w, r, "/archive/requests/"+request.ID, http.StatusSeeOther)
		return
	}
	title := "Deine Einreichung"
	if request.Status == files.Draft {
		title = "Prüfe deine Angaben"
	}
	if request.Status == files.Pending {
		title = "Einreichung eingegangen"
	}
	c.render(w, "request", false, Page{Title: title, Active: "archive", Request: &RequestPage{Entry: request, EditURL: "/archive/requests/" + request.ID + "/edit"}})
}

func (c *Handler) ReviewArchiveRequest(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/verwaltung/antraege/") {
		http.Redirect(w, r, "/verwaltung/einreichungen/"+r.PathValue("request"), http.StatusTemporaryRedirect)
		return
	}
	if c.archive == nil {
		c.notFound(w, r)
		return
	}
	session := w.(*sessionResponse).session
	request, err := c.archive.Request(r.Context(), session.UserID, r.PathValue("request"), true)
	if err != nil {
		c.archiveWriteError(w, r, err)
		return
	}
	page := Page{Title: "Einreichung prüfen", Active: "admin", Admin: &AdminPage{Section: "einreichungen", SectionTitle: "Einreichungen", View: "review", ReviewerOnly: !session.CanAdmin()}, Request: &RequestPage{Entry: request, Reviewing: true, EditURL: "/verwaltung/einreichungen/" + request.ID + "/edit"}}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		decision := r.PostForm.Get("decision")
		if decision != "approve" && decision != "reject" {
			c.pageError(w, "Wähle Bestätigen oder Ablehnen aus.", http.StatusBadRequest)
			return
		}
		err := c.archive.Review(r.Context(), session.UserID, request.ID, decision == "approve")
		if errors.Is(err, files.ErrConflict) {
			page.Request.Error = "Die Einreichung oder die bisherige Revision wurde bereits geändert. Lade die Seite neu und prüfe den aktuellen Stand."
			page.DocumentTitle = "Fehler: Einreichung prüfen"
			status = http.StatusConflict
		} else if err != nil {
			c.archiveWriteError(w, r, err)
			return
		} else {
			http.Redirect(w, r, "/verwaltung/einreichungen/"+request.ID, http.StatusSeeOther)
			return
		}
	}
	c.renderStatus(w, "request", false, page, status)
}
