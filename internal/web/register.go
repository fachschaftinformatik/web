package web

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/fachschaftinformatik/web/internal/users"
)

// RegisterPage never retains a submitted password.
type RegisterPage struct {
	Name, Email, NameError, EmailError, PasswordError string
	ProgramID, ProgramError                           string
	Programs                                          []users.Program
	Submitted                                         bool
}

func (c *Handler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	data := Page{Title: "Konto erstellen", Active: "register"}
	data.Register.Programs = users.Programs()
	if c.identity != nil {
		var err error
		data.Register.Programs, err = c.identity.Programs(r.Context())
		if err != nil {
			c.serverError(w, err)
			return
		}
	}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Das Formular konnte nicht gelesen werden. Lade die Seite neu und versuche es noch einmal.", http.StatusBadRequest)
			return
		}
		form := &data.Register
		form.Name = strings.TrimSpace(r.PostForm.Get("name"))
		form.Email = strings.TrimSpace(r.PostForm.Get("email"))
		form.ProgramID = r.PostForm.Get("program_id")
		if form.ProgramID != "" && !slices.ContainsFunc(form.Programs, func(p users.Program) bool { return p.ID == form.ProgramID }) {
			form.ProgramError = "Bitte wähle deinen Studiengang aus."
		}
		form.NameError, form.EmailError, form.PasswordError = users.RegistrationErrors(form.Name, form.Email, r.PostForm.Get("password"))
		if form.NameError != "" || form.EmailError != "" || form.PasswordError != "" || form.ProgramError != "" {
			data.DocumentTitle = "Fehler: Konto erstellen"
			status = http.StatusUnprocessableEntity
		} else if c.identity != nil {
			err := c.identity.Register(r.Context(), form.Name, form.Email, r.PostForm.Get("password"), form.ProgramID)
			if errors.Is(err, users.ErrEmailExists) {
				form.EmailError = err.Error()
				data.DocumentTitle = "Fehler: Konto erstellen"
				status = http.StatusUnprocessableEntity
			} else if err != nil {
				c.serverError(w, err)
				return
			} else {
				form.Submitted = true
			}
		} else {
			// No account is created and no verification email is sent in this demo.
			form.EmailError = "Dein Konto konnte nicht erstellt werden."
			data.DocumentTitle = "Fehler: Konto erstellen"
			status = http.StatusServiceUnavailable
		}
	}
	c.renderStatus(w, "register", false, data, status)
}
