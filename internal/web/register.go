package web

import (
	"net/http"
	"strings"

	"github.com/fachschaftinformatik/web/internal/account"
)

// RegisterPage never retains a submitted password.
type RegisterPage struct {
	Name, Email, NameError, EmailError, PasswordError string
	ProgramID, ProgramError                           string
	Programs                                          []account.Program
	Submitted                                         bool
}

func (c *Handler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	data := Page{Title: "Konto erstellen", Active: "register"}
	data.Register.Programs = account.Programs()
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
		form.ProgramError = account.ProgramError(form.ProgramID)
		form.NameError, form.EmailError, form.PasswordError = account.RegistrationErrors(form.Name, form.Email, r.PostForm.Get("password"))
		if form.NameError != "" || form.EmailError != "" || form.PasswordError != "" || form.ProgramError != "" {
			data.DocumentTitle = "Fehler: Konto erstellen"
			status = http.StatusUnprocessableEntity
		} else {
			// No account is created and no verification email is sent in this demo.
			form.Submitted = true
			status = http.StatusServiceUnavailable
		}
	}
	c.renderStatus(w, "register", false, data, status)
}
