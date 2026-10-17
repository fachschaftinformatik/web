package controller

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/fachschaftinformatik/web/internal/model/archive"
	"github.com/fachschaftinformatik/web/internal/model/users"
)

const sessionCookie = "web-session"

type session = users.Session

func randomToken() string {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(value[:])
}

type sessionResponse struct {
	http.ResponseWriter
	session *session
}

func (c *Handler) setSessionCookie(w http.ResponseWriter, r *http.Request, s *session) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: s.ID, Path: "/", HttpOnly: true, Secure: c.secureCookies || r.TLS != nil, SameSite: http.SameSiteLaxMode, MaxAge: 8 * 60 * 60})
}

func (c *Handler) withSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var s *session
		if cookie, err := r.Cookie(sessionCookie); err == nil {

			var err error
			s, err = c.identity.Session(r.Context(), cookie.Value)
			if err != nil {
				c.serverError(w, err)
				return
			}

		}
		account := r.URL.Path == "/login" || r.URL.Path == "/register" || r.URL.Path == "/forgot" || r.URL.Path == "/reset" || r.URL.Path == "/verify" || r.URL.Path == "/invite"
		archiveWrite := r.URL.Path == "/archive/new" || r.URL.Path == "/archive/preferences" || strings.HasPrefix(r.URL.Path, "/archive/requests") || strings.HasPrefix(r.URL.Path, "/archive/files/") && strings.HasSuffix(r.URL.Path, "/revision")
		forumWrite := r.URL.Path == "/community/neuigkeiten/new" || r.URL.Path == "/d/new" || strings.HasPrefix(r.URL.Path, "/d/") && (strings.HasSuffix(r.URL.Path, "/edit") || r.Method == http.MethodPost) || strings.HasPrefix(r.URL.Path, "/community/veranstaltungen/") && (strings.HasSuffix(r.URL.Path, "/new") || strings.HasSuffix(r.URL.Path, "/edit"))
		if s == nil && (r.URL.Path == "/login" || account) && r.Method == http.MethodGet {

			var err error
			s, err = c.identity.CreateSession(r.Context(), 0)
			if err != nil {
				c.serverError(w, err)
				return
			}

			c.setSessionCookie(w, r, s)
		}
		response := &sessionResponse{ResponseWriter: w, session: s}
		admin := r.URL.Path == "/verwaltung" || strings.HasPrefix(r.URL.Path, "/verwaltung/")
		review := r.URL.Path == "/verwaltung/einreichungen" || strings.HasPrefix(r.URL.Path, "/verwaltung/einreichungen/") || r.URL.Path == "/verwaltung/antraege" || strings.HasPrefix(r.URL.Path, "/verwaltung/antraege/")
		if admin || r.URL.Path == "/login" || r.URL.Path == "/logout" || s != nil {
			w.Header().Set("Cache-Control", "no-store")
		}
		w.Header().Add("Vary", "Cookie")
		if (admin || archiveWrite || forumWrite) && (s == nil || !s.Authenticated) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if admin && !(s.CanAdmin() || (review || r.URL.Path == "/verwaltung") && s.CanReview()) {
			c.pageError(response, "Du hast keine Berechtigung für die Verwaltung.", http.StatusForbidden)
			return
		}
		if r.Method == http.MethodPost && (admin || archiveWrite || forumWrite || r.URL.Path == "/logout" || account) {
			limit := int64(uploadBodyLimit)
			if account {
				limit = 64 << 10
			}
			if archiveWrite {
				limit = archive.MaxUploadBytes + 1<<20
			}
			if forumWrite {
				limit = 8 << 20
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
				if err := r.ParseMultipartForm(uploadBodyLimit); err != nil {
					c.pageError(response, "Das Formular ist zu groß oder konnte nicht gelesen werden.", http.StatusBadRequest)
					return
				}
				defer r.MultipartForm.RemoveAll()
			} else {
				if err := r.ParseForm(); err != nil {
					c.pageError(response, "Das Formular konnte nicht gelesen werden.", http.StatusBadRequest)
					return
				}
			}
			if !validCSRF(s, r.PostForm.Get("csrf")) {
				c.pageError(response, "Lade die Seite neu und versuche es noch einmal.", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(response, r)
	})
}

func validCSRF(s *session, value string) bool {
	return s != nil && subtle.ConstantTimeCompare([]byte(s.CSRF), []byte(value)) == 1
}

func (c *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if response := w.(*sessionResponse); response.session != nil {

		if err := c.identity.DeleteSession(r.Context(), response.session.ID); err != nil {
			c.serverError(w, err)
			return
		}

	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Path: "/", MaxAge: -1, HttpOnly: true, Secure: c.secureCookies || r.TLS != nil, SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
