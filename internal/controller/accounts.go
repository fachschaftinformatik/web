package controller

import (
	"database/sql"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/fachschaftinformatik/web/internal/model/users"
	"github.com/fachschaftinformatik/web/internal/urls"
	"github.com/fachschaftinformatik/web/internal/view"
)

func (c *Handler) ResendUserConfirmation(w http.ResponseWriter, r *http.Request) {

	userID, err := c.identity.ResolvePublicID(r.Context(), r.PathValue("id"))
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	queued, err := c.identity.ResendConfirmation(r.Context(), userID)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	target := urls.AdminUser(r.PathValue("id"))
	if queued {
		flash(w, "Die E-Mail zur Bestätigung wurde zum Versand vorgemerkt.")
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (c *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	data := view.Page{Title: "Passwort vergessen", Active: "forgot"}
	if r.Method != http.MethodPost && takeFlash(w) == "password-reset-requested" {
		data.Forgot.Submitted = true
		data.Title = "Prüfe dein E-Mail-Postfach"
	}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		if err := r.ParseForm(); err != nil {
			c.pageError(w, "Das Formular konnte nicht gelesen werden. Lade die Seite neu und versuche es noch einmal.", http.StatusBadRequest)
			return
		}
		data.Forgot.Email = strings.TrimSpace(r.PostForm.Get("email"))
		data.Forgot.EmailError = c.identity.EmailError(data.Forgot.Email)
		if data.Forgot.EmailError != "" {
			data.DocumentTitle = "Fehler: Passwort vergessen"
			status = http.StatusUnprocessableEntity
		} else {
			if err := c.identity.RequestReset(r.Context(), data.Forgot.Email); err != nil {
				c.serverError(w, err)
				return
			}
			flash(w, "password-reset-requested")
			http.Redirect(w, r, urls.ForgotPasswordPath, http.StatusSeeOther)
			return
		}

	}
	c.renderStatus(w, "forgot", false, data, status)
}

func (c *Handler) ResendInvitation(w http.ResponseWriter, r *http.Request) {
	internal, err := c.members.ResolvePublicID(r.Context(), r.PathValue("id"))
	if errors.Is(err, sql.ErrNoRows) {
		c.notFound(w, r)
		return
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	queued, err := c.members.Resend(r.Context(), c.identity, internal)
	if errors.Is(err, sql.ErrNoRows) {
		c.notFound(w, r)
		return
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	target := urls.AdminMembersPath
	if queued {
		flash(w, "Die Einladung wurde zum Versand vorgemerkt.")
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (c *Handler) Invitation(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	page := view.Page{Title: "Einladung annehmen", Active: "register", Invite: view.InvitationPage{Value: r.URL.Query().Get("token")}}
	form := &page.Invite
	if r.Method != http.MethodPost && takeFlash(w) == "invitation-accepted" {
		form.Done = true
		c.render(w, "invite", false, page)
		return
	}
	if r.Method == http.MethodPost {
		form.Value = r.PostForm.Get("token")
	}
	invitation, err := c.identity.Invitation(r.Context(), form.Value)
	status := http.StatusOK
	if errors.Is(err, users.ErrToken) {
		form.Error = err.Error()
		status = http.StatusUnprocessableEntity
	} else {
		if err != nil {
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
					} else {
						if err != nil {
							c.serverError(w, err)
							return
						} else {
							flash(w, "invitation-accepted")
							http.Redirect(w, r, urls.AcceptInvitationPath, http.StatusSeeOther)
							return
						}
					}
				}
				if form.Error != "" || form.PasswordError != "" {
					status = http.StatusUnprocessableEntity
				}
			}
		}
	}
	if status != http.StatusOK {
		page.DocumentTitle = "Fehler: Einladung annehmen"
	}
	c.renderStatus(w, "invite", false, page, status)
}

func (c *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	data := view.Page{Title: "Anmelden", Active: "login"}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		if err := r.ParseForm(); err != nil {
			c.pageError(w, "Das Formular konnte nicht gelesen werden. Lade die Seite neu und versuche es noch einmal.", http.StatusBadRequest)
			return
		}
		data.Login.Email = strings.TrimSpace(r.PostForm.Get("email"))

		data.Login.EmailError = c.identity.EmailError(data.Login.Email)
		if r.PostForm.Get("password") == "" {
			data.Login.PasswordError = "Gib dein Passwort ein."
		}
		if data.Login.EmailError != "" || data.Login.PasswordError != "" {
			data.DocumentTitle = "Fehler: Anmelden"
			status = http.StatusUnprocessableEntity
		} else {
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
				target := urls.HomePath
				if s.CanAdmin() {
					target = urls.AdminPath
				} else {
					if s.CanReview() {
						target = urls.AdminPath
					}
				}
				http.Redirect(w, r, target, http.StatusSeeOther)
				return
			}
		}

	}
	c.renderStatus(w, "login", false, data, status)
}

func (c *Handler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	data := view.Page{Title: "Konto erstellen", Active: "register"}
	if r.Method != http.MethodPost && takeFlash(w) == "registration-submitted" {
		data.Register.Submitted = true
	}

	var err error
	data.Register.Programs, err = c.identity.Programs(r.Context())
	if err != nil {
		c.serverError(w, err)
		return
	}

	status := http.StatusOK
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		if err := r.ParseForm(); err != nil {
			c.pageError(w, "Das Formular konnte nicht gelesen werden. Lade die Seite neu und versuche es noch einmal.", http.StatusBadRequest)
			return
		}
		form := &data.Register
		form.Name = strings.TrimSpace(r.PostForm.Get("name"))
		form.Email = strings.TrimSpace(r.PostForm.Get("email"))
		form.ProgramID = r.PostForm.Get("program_id")
		if form.ProgramID != "" && !slices.ContainsFunc(form.Programs, func(p users.Program) bool { return p.ID == form.ProgramID }) {
			form.ProgramError = "Bitte wähle deinen Studiengang aus."
		}
		form.NameError, form.EmailError, form.PasswordError = c.identity.RegistrationErrors(form.Name, form.Email, r.PostForm.Get("password"))
		if form.NameError != "" || form.EmailError != "" || form.PasswordError != "" || form.ProgramError != "" {
			data.DocumentTitle = "Fehler: Konto erstellen"
			status = http.StatusUnprocessableEntity
		} else {
			err := c.identity.Register(r.Context(), form.Name, form.Email, r.PostForm.Get("password"), form.ProgramID)
			if errors.Is(err, users.ErrEmailExists) {
				form.EmailError = err.Error()
				data.DocumentTitle = "Fehler: Konto erstellen"
				status = http.StatusUnprocessableEntity
			} else {
				if err != nil {
					c.serverError(w, err)
					return
				} else {
					flash(w, "registration-submitted")
					http.Redirect(w, r, urls.RegisterPath, http.StatusSeeOther)
					return
				}
			}
		}

	}
	c.renderStatus(w, "register", false, data, status)
}

// Token displays a confirmation first so mail scanners cannot consume links.
func (c *Handler) Token(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	kind, title := "reset", "Passwort zurücksetzen"
	if r.PathValue("tokenKind") == "verify" {
		kind, title = "verify", "E-Mail-Adresse bestätigen"
	}
	formKind := kind
	if kind == "verify" {
		formKind = "confirm"
	}
	page := view.Page{Title: title, Active: "forgot", Token: view.TokenPage{Kind: formKind, Value: r.URL.Query().Get("token")}}
	if r.Method != http.MethodPost && takeFlash(w) == "account-token-"+kind {
		page.Token.Done = true
	}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		page.Token.Value = r.PostForm.Get("token")
		password := r.PostForm.Get("password")
		if kind == "reset" {
			page.Token.Error = users.PasswordError(password)
		}
		if page.Token.Error == "" {
			err := c.identity.ConsumeToken(r.Context(), page.Token.Value, kind, password)
			if errors.Is(err, users.ErrToken) {
				page.Token.Error = err.Error()
			} else {
				if err != nil {
					c.serverError(w, err)
					return
				} else {
					flash(w, "account-token-"+kind)
					http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
					return
				}
			}
		}
		if page.Token.Error != "" {
			status = http.StatusUnprocessableEntity
			page.DocumentTitle = "Fehler: " + title
		}
	}
	c.renderStatus(w, "token", false, page, status)
}
