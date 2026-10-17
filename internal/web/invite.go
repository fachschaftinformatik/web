package web

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/fachschaftinformatik/web/internal/users"
)

type InvitationPage struct {
	Value, Name, Email, Error, PasswordError string
	Done                                     bool
}

func (c *Handler) ResendInvitation(w http.ResponseWriter, r *http.Request) {
	if c.identity == nil {
		http.NotFound(w, r)
		return
	}
	queued, err := c.people.Resend(r.Context(), c.identity, r.PathValue("id"))
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	target := "/verwaltung/mitglieder?gespeichert=1"
	if queued {
		target = "/verwaltung/mitglieder?einladung=1"
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (c *Handler) Invitation(w http.ResponseWriter, r *http.Request) {
	if c.identity == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	page := Page{Title: "Einladung annehmen", Active: "register", Invite: InvitationPage{Value: r.URL.Query().Get("token")}}
	form := &page.Invite
	if r.Method == http.MethodPost {
		form.Value = r.PostForm.Get("token")
	}
	invitation, err := c.identity.Invitation(r.Context(), form.Value)
	status := http.StatusOK
	if errors.Is(err, users.ErrToken) {
		form.Error = err.Error()
		status = http.StatusUnprocessableEntity
	} else if err != nil {
		c.serverError(w, err)
		return
	} else {
		form.Name, form.Email = invitation.Name, invitation.Email
		if r.Method == http.MethodPost {
			password := r.PostForm.Get("password")
			form.PasswordError = users.PasswordError(password)
			if form.PasswordError == "" {
				err := c.identity.AcceptInvitation(r.Context(), form.Value, password)
				if errors.Is(err, users.ErrToken) {
					form.Error = err.Error()
				} else if err != nil {
					c.serverError(w, err)
					return
				} else {
					form.Done = true
				}
			}
			if form.Error != "" || form.PasswordError != "" {
				status = http.StatusUnprocessableEntity
			}
		}
	}
	if status != http.StatusOK {
		page.DocumentTitle = "Fehler: Einladung annehmen"
	}
	c.renderStatus(w, "invite", false, page, status)
}
