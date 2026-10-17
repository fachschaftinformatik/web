package web

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fachschaftinformatik/web/internal/content/files"
	"github.com/fachschaftinformatik/web/internal/users"
)

const sessionCookie = "web-session"

type session = users.Session

type sessionStore struct {
	mu    sync.Mutex
	items map[string]session
}

func newSessionStore() sessionStore { return sessionStore{items: make(map[string]session)} }

func randomToken() string {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(value[:])
}

func (s *sessionStore) get(id string) *session {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return nil
	}
	if time.Now().After(item.Expires) {
		delete(s.items, id)
		return nil
	}
	return &item
}

func (s *sessionStore) create(authenticated bool) *session {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, item := range s.items {
		if time.Now().After(item.Expires) {
			delete(s.items, id)
		}
	}
	item := session{ID: randomToken(), CSRF: randomToken(), Authenticated: authenticated, Expires: time.Now().Add(8 * time.Hour)}
	if authenticated {
		item.Role = "admin"
	}
	s.items[item.ID] = item
	return &item
}

func (s *sessionStore) remove(id string) { s.mu.Lock(); defer s.mu.Unlock(); delete(s.items, id) }

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
			if c.identity != nil {
				var err error
				s, err = c.identity.Session(r.Context(), cookie.Value)
				if err != nil {
					c.serverError(w, err)
					return
				}
			} else {
				s = c.sessions.get(cookie.Value)
			}
		}
		account := r.URL.Path == "/login" || r.URL.Path == "/register" || r.URL.Path == "/forgot" || r.URL.Path == "/reset" || r.URL.Path == "/verify"
		archiveWrite := r.URL.Path == "/archive/new" || strings.HasPrefix(r.URL.Path, "/archive/requests") || strings.HasPrefix(r.URL.Path, "/archive/files/") && strings.HasSuffix(r.URL.Path, "/revision")
		if s == nil && (r.URL.Path == "/login" || c.identity != nil && account) && r.Method == http.MethodGet {
			if c.identity != nil {
				var err error
				s, err = c.identity.CreateSession(r.Context(), 0)
				if err != nil {
					c.serverError(w, err)
					return
				}
			} else {
				s = c.sessions.create(false)
			}
			c.setSessionCookie(w, r, s)
		}
		response := &sessionResponse{ResponseWriter: w, session: s}
		admin := r.URL.Path == "/verwaltung" || strings.HasPrefix(r.URL.Path, "/verwaltung/")
		review := r.URL.Path == "/verwaltung/antraege" || strings.HasPrefix(r.URL.Path, "/verwaltung/antraege/")
		if admin || r.URL.Path == "/login" || r.URL.Path == "/logout" || s != nil {
			w.Header().Set("Cache-Control", "no-store")
		}
		w.Header().Add("Vary", "Cookie")
		if (admin || archiveWrite) && (s == nil || !s.Authenticated) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if admin && !(s.CanAdmin() || review && s.CanReview()) {
			http.Error(w, "Du hast keine Berechtigung für die Verwaltung.", http.StatusForbidden)
			return
		}
		if r.Method == http.MethodPost && (admin || archiveWrite || r.URL.Path == "/logout" || c.identity != nil && account) {
			limit := int64(uploadBodyLimit)
			if account {
				limit = 64 << 10
			}
			if archiveWrite {
				limit = files.MaxUploadBytes + 1<<20
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
				if err := r.ParseMultipartForm(uploadBodyLimit); err != nil {
					http.Error(w, "Das Formular ist zu groß oder konnte nicht gelesen werden.", http.StatusBadRequest)
					return
				}
				defer r.MultipartForm.RemoveAll()
			} else if err := r.ParseForm(); err != nil {
				http.Error(w, "Das Formular konnte nicht gelesen werden.", http.StatusBadRequest)
				return
			}
			if !validCSRF(s, r.PostForm.Get("csrf")) {
				http.Error(w, "Lade die Seite neu und versuche es noch einmal.", http.StatusForbidden)
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
		if c.identity != nil {
			if err := c.identity.DeleteSession(r.Context(), response.session.ID); err != nil {
				c.serverError(w, err)
				return
			}
		} else {
			c.sessions.remove(response.session.ID)
		}
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Path: "/", MaxAge: -1, HttpOnly: true, Secure: c.secureCookies || r.TLS != nil, SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
