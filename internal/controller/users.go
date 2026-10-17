package controller

import (
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/fachschaftinformatik/web/internal/model/forum"
	"github.com/fachschaftinformatik/web/internal/model/users"
	"github.com/fachschaftinformatik/web/internal/view"
)

func (c *Handler) Admin(w http.ResponseWriter, r *http.Request) {
	if err := c.people.Refresh(r.Context()); err != nil {
		c.serverError(w, err)
		return
	}
	section, id := r.PathValue("section"), r.PathValue("id")
	a := &view.AdminPage{Section: section, View: "index"}
	if section == "referate" || section == "projektstellen" || section == "aufgaben" {
		if id == "" {
			target := "/verwaltung/funktionen"
			if r.URL.Query().Get("gespeichert") == "1" {
				target += "?gespeichert=1"
			}
			http.Redirect(w, r, target, http.StatusSeeOther)
			return
		}
		a.Section, a.SectionTitle = "funktionen", "Funktionen"
		c.EditResponsibility(w, r, a)
		return
	}
	if section == "" {
		c.adminRender(w, "Verwaltung", a, http.StatusOK)
		return
	}
	if section == "antraege" {
		target := "/verwaltung/einreichungen"
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusTemporaryRedirect)
		return
	}
	for _, link := range view.AdminLinks {
		if link.ID == section {
			a.SectionTitle = link.Label
		}
	}
	if a.SectionTitle == "" {
		c.notFound(w, r)
		return
	}
	if r.URL.Query().Get("gespeichert") == "1" {
		a.Success = "Die Änderungen wurden gespeichert."
	}
	if r.URL.Query().Get("eingeladen") == "1" {
		a.Success = "Das Mitglied wurde hinzugefügt."
	}
	if r.URL.Query().Get("einladung") == "1" {
		a.Success = "Die Einladung wurde zum Versand vorgemerkt."
	}
	switch section {
	case "einreichungen":
		if c.archive == nil {
			c.notFound(w, r)
			return
		}
		requests, err := c.archive.Requests(r.Context(), w.(*sessionResponse).session.UserID, true)
		if err != nil {
			c.archiveWriteError(w, r, err)
			return
		}
		a.View, a.Requests = "requests", requests
	case "benutzer":
		var directory *users.Directory
		if c.identity != nil {
			var err error
			directory, err = c.identity.Directory(r.Context())
			if err != nil {
				c.serverError(w, err)
				return
			}
		}
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if id != "" {
			if user, ok := directory.Get(id); ok {
				a.View = "user"
				if r.URL.Query().Get("zuruecksetzen") == "1" {
					a.Success = "Die E-Mail zum Zurücksetzen des Passworts wurde zum Versand vorgemerkt."
				}
				a.User = &user
				for _, person := range c.people.People() {
					if person.UserID != 0 && strconv.FormatInt(person.UserID, 10) == user.ID {
						a.Person = &person
						break
					}
				}
				if r.URL.Query().Get("bestaetigung") == "1" {
					a.Success = "Die E-Mail zur Bestätigung wurde zum Versand vorgemerkt."
					if a.Person != nil {
						a.Success = "Die Einladung wurde zum Versand vorgemerkt."
					}
				}
				c.adminRender(w, user.Name, a, http.StatusOK)
				return
			}
			c.notFound(w, r)
			return
		}
		a.View, a.Query = "users", strings.TrimSpace(r.URL.Query().Get("q"))
		a.Page = 1
		if value := r.URL.Query().Get("seite"); value != "" {
			page, err := strconv.Atoi(value)
			if err != nil {
				http.Redirect(w, r, a.UserPageURL(1), http.StatusSeeOther)
				return
			}
			a.Page = page
		}
		page, ok := directory.Search(a.Query, a.Page)
		if !ok {
			http.Redirect(w, r, a.UserPageURL(1), http.StatusSeeOther)
			return
		}
		a.Users, a.Total, a.Pages = page.Users, page.Total, page.Pages
		if a.Page > 1 {
			a.PreviousURL = a.UserPageURL(a.Page - 1)
		}
		if a.Page < a.Pages {
			a.NextURL = a.UserPageURL(a.Page + 1)
		}
	case "mitglieder", "funktionen":
		if id != "" {
			if section == "funktionen" {
				c.EditResponsibility(w, r, a)
			} else {
				c.EditMember(w, r, a)
			}
			return
		}
		a.View = "people"
		if section == "funktionen" {
			a.View = "functions"
		}
		members := c.people.People()
		responsibilities := append(c.people.Responsibilities("board"), c.people.Responsibilities("departments")...)
		responsibilities = append(responsibilities, c.people.Responsibilities("project")...)
		for _, person := range members {
			row := view.AdminMemberRow{Person: person}
			for _, task := range responsibilities {
				if slices.Contains(task.MemberIDs, person.ID) {
					row.Tasks = append(row.Tasks, task.Title+" ("+responsibilityKind(task.Group)+")")
				}
			}
			a.Members = append(a.Members, row)
		}
		for _, task := range responsibilities {
			var names []string
			for _, person := range members {
				if slices.Contains(task.MemberIDs, person.ID) {
					names = append(names, person.Name)
				}
			}
			a.Tasks = append(a.Tasks, view.AdminTaskRow{Title: task.Title, Kind: responsibilityKind(task.Group), Members: names, URL: "/verwaltung/funktionen/" + task.ID})
		}
	case "sitzungen":
		if id != "" {
			c.EditPublication(w, r, a)
			return
		}
		a.View = "meetings"
		var err error
		a.Meetings, err = c.meetings.AllWithError()
		if err != nil {
			c.serverError(w, err)
			return
		}
		if r.URL.Query().Get("veroeffentlicht") == "1" {
			a.Success = "Das Dokument wurde veröffentlicht."
		}
	}
	c.adminRender(w, a.SectionTitle, a, http.StatusOK)
}

// ResetUserPassword confirms the destination before queuing a reset email.
func (c *Handler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	var directory *users.Directory

	var err error
	directory, err = c.identity.Directory(r.Context())
	if err != nil {
		c.serverError(w, err)
		return
	}

	if user, ok := directory.Get(r.PathValue("id")); ok {
		if r.Method == http.MethodPost {

			if err := c.identity.RequestReset(r.Context(), user.Email); err != nil {
				c.serverError(w, err)
				return
			}
			http.Redirect(w, r, "/verwaltung/benutzer/"+url.PathEscape(user.ID)+"?zuruecksetzen=1", http.StatusSeeOther)
			return

		}
		c.adminRender(w, "Passwort zurücksetzen", &view.AdminPage{Section: "benutzer", SectionTitle: "Benutzer", View: "reset", User: &user}, http.StatusOK)
		return
	}
	c.notFound(w, r)
}

func (c *Handler) PublicProfile(w http.ResponseWriter, r *http.Request) {

	profile, err := c.identity.PublicProfile(r.Context(), r.PathValue("id"))
	if errors.Is(err, users.ErrProfileNotFound) {
		c.notFound(w, r)
		return
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	records, total, err := c.forum.FindPage(r.Context(), forum.Filter{AuthorID: profile.ID}, requestedPage(r), 20, forumUser(w))
	if err != nil {
		c.serverError(w, err)
		return
	}
	c.render(w, "profile", false, view.Page{Title: profile.Name, Active: "forum", Community: true, Forum: view.ForumPage{Posts: records, Pagination: listPagination(r, total, 20)}})
}
