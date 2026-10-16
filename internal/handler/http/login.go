package web

import (
	"net/http"
	"strings"

	"github.com/fachschaftinformatik/web/internal/account"
)

// LoginPage supports an explicit local demo credential; normal account login remains a preview.
type LoginPage struct {
	Email, EmailError, PasswordError string
	Submitted                        bool
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
		if data.Login.Email == "admin" {
			response := w.(*sessionResponse)
			if !validCSRF(response.session, r.PostForm.Get("csrf")) {
				http.Error(w, "Lade die Seite neu und versuche es noch einmal.", http.StatusForbidden)
				return
			}
			if r.PostForm.Get("password") == "admin" {
				c.sessions.remove(response.session.ID)
				s := c.sessions.create(true)
				setSessionCookie(w, r, s)
				http.Redirect(w, r, "/verwaltung", http.StatusSeeOther)
				return
			}
			data.Login.PasswordError = "Benutzername oder Passwort ist falsch."
			data.DocumentTitle = "Fehler: Anmelden"
			c.renderStatus(w, "login", false, data, http.StatusUnprocessableEntity)
			return
		}
		data.Login.EmailError = account.EmailError(data.Login.Email)
		if r.PostForm.Get("password") == "" {
			data.Login.PasswordError = "Gib dein Passwort ein."
		}
		if data.Login.EmailError != "" || data.Login.PasswordError != "" {
			data.DocumentTitle = "Fehler: Anmelden"
			status = http.StatusUnprocessableEntity
		} else {
			// No credentials are checked, persisted or forwarded by this demo.
			data.Login.Submitted = true
			data.Login.Email = ""
		}
	}
	c.renderStatus(w, "login", false, data, status)
}
