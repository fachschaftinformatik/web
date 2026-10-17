package web

import (
	"net/http"
	"strings"

	"github.com/fachschaftinformatik/web/internal/users"
)

type ForgotPage struct {
	Email, EmailError string
	Unavailable       bool
	Submitted         bool
}

func (c *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	data := Page{Title: "Passwort vergessen", Active: "forgot"}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Das Formular konnte nicht gelesen werden. Lade die Seite neu und versuche es noch einmal.", http.StatusBadRequest)
			return
		}
		data.Forgot.Email = strings.TrimSpace(r.PostForm.Get("email"))
		data.Forgot.EmailError = users.EmailError(data.Forgot.Email)
		if data.Forgot.EmailError != "" {
			data.DocumentTitle = "Fehler: Passwort vergessen"
			status = http.StatusUnprocessableEntity
		} else if c.identity != nil {
			if err := c.identity.RequestReset(r.Context(), data.Forgot.Email); err != nil {
				c.serverError(w, err)
				return
			}
			data.Forgot = ForgotPage{Submitted: true}
			data.Title = "Prüfe dein E-Mail-Postfach"
		} else {
			// The isolated demo has no mail delivery or account lookup.
			// Never imply a reset email was sent or an account was found.
			data.Forgot = ForgotPage{Unavailable: true}
			data.Title = "Passwort zurücksetzen"
			status = http.StatusServiceUnavailable
		}
	}
	c.renderStatus(w, "forgot", false, data, status)
}
