package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/fachschaftinformatik/web/internal/users"
)

// LoginPage never retains a submitted password.
type LoginPage struct {
	Email, EmailError, PasswordError string
}

func (c *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	data := Page{Title: "Anmelden", Active: "login"}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Das Formular konnte nicht gelesen werden. Lade die Seite neu und versuche es noch einmal.", http.StatusBadRequest)
			return
		}
		data.Login.Email = strings.TrimSpace(r.PostForm.Get("email"))
		if c.identity == nil && data.Login.Email == "admin" {
			response := w.(*sessionResponse)
			if !validCSRF(response.session, r.PostForm.Get("csrf")) {
				http.Error(w, "Lade die Seite neu und versuche es noch einmal.", http.StatusForbidden)
				return
			}
			if r.PostForm.Get("password") == "admin" {
				c.sessions.remove(response.session.ID)
				s := c.sessions.create(true)
				c.setSessionCookie(w, r, s)
				http.Redirect(w, r, "/verwaltung", http.StatusSeeOther)
				return
			}
			data.Login.PasswordError = "Benutzername oder Passwort ist falsch."
			data.DocumentTitle = "Fehler: Anmelden"
			c.renderStatus(w, "login", false, data, http.StatusUnprocessableEntity)
			return
		}
		data.Login.EmailError = users.EmailError(data.Login.Email)
		if r.PostForm.Get("password") == "" {
			data.Login.PasswordError = "Gib dein Passwort ein."
		}
		if data.Login.EmailError != "" || data.Login.PasswordError != "" {
			data.DocumentTitle = "Fehler: Anmelden"
			status = http.StatusUnprocessableEntity
		} else if c.identity != nil {
			userID, err := c.identity.Login(r.Context(), data.Login.Email, r.PostForm.Get("password"))
			if err != nil {
				if !errors.Is(err, users.ErrCredentials) && !errors.Is(err, users.ErrInactive) && !errors.Is(err, users.ErrUnverified) {
					c.serverError(w, err)
					return
				}
				data.Login.PasswordError = err.Error()
				data.DocumentTitle = "Fehler: Anmelden"
				status = http.StatusUnprocessableEntity
			} else {
				response := w.(*sessionResponse)
				if err := c.identity.DeleteSession(r.Context(), response.session.ID); err != nil {
					c.serverError(w, err)
					return
				}
				s, err := c.identity.CreateSession(r.Context(), userID)
				if err != nil {
					c.serverError(w, err)
					return
				}
				c.setSessionCookie(w, r, s)
				target := "/"
				if s.CanAdmin() {
					target = "/verwaltung"
				} else if s.CanReview() {
					target = "/verwaltung"
				}
				http.Redirect(w, r, target, http.StatusSeeOther)
				return
			}
		} else {
			// No credentials are checked, persisted or forwarded by this demo.
			data.Login.PasswordError = users.ErrCredentials.Error()
			data.DocumentTitle = "Fehler: Anmelden"
			status = http.StatusUnprocessableEntity
		}
	}
	c.renderStatus(w, "login", false, data, status)
}
