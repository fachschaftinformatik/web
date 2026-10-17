package web

import (
	"errors"
	"net/http"

	"github.com/fachschaftinformatik/web/internal/users"
)

type TokenPage struct {
	Value, Kind, Error string
	Done               bool
}

// Token displays a confirmation first so mail scanners cannot consume links.
func (c *Handler) Token(w http.ResponseWriter, r *http.Request) {
	if c.identity == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	kind, title := "reset", "Passwort zurücksetzen"
	if r.URL.Path == "/verify" {
		kind, title = "verify", "E-Mail-Adresse bestätigen"
	}
	page := Page{Title: title, Active: "forgot", Token: TokenPage{Kind: kind, Value: r.URL.Query().Get("token")}}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		page.Token.Value = r.PostForm.Get("token")
		password := r.PostForm.Get("password")
		if kind == "reset" {
			_, _, page.Token.Error = users.RegistrationErrors("Valid Name", "valid@studmail.w-hs.de", password)
		}
		if page.Token.Error == "" {
			err := c.identity.ConsumeToken(r.Context(), page.Token.Value, kind, password)
			if errors.Is(err, users.ErrToken) {
				page.Token.Error = err.Error()
			} else if err != nil {
				c.serverError(w, err)
				return
			} else {
				page.Token.Done = true
			}
		}
		if page.Token.Error != "" {
			status = http.StatusUnprocessableEntity
			page.DocumentTitle = "Fehler: " + title
		}
	}
	c.renderStatus(w, "token", false, page, status)
}
