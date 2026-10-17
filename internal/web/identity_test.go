package web

import (
	"context"
	"database/sql"
	"io"
	"log"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/content/files"
	"github.com/fachschaftinformatik/web/internal/content/media"
	"github.com/fachschaftinformatik/web/internal/content/posts"
	"github.com/fachschaftinformatik/web/internal/data"
	"github.com/fachschaftinformatik/web/internal/data/blob"
	"github.com/fachschaftinformatik/web/internal/data/blob/disk"
	"github.com/fachschaftinformatik/web/internal/people"
	"github.com/fachschaftinformatik/web/internal/users"
	"golang.org/x/crypto/bcrypt"
)

func identityBrowser(t *testing.T) (*adminBrowser, *sql.DB) {
	return identityBrowserWithStorage(t, nil)
}

func identityBrowserWithStorage(t *testing.T, storage files.Storage) (*adminBrowser, *sql.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sqlite.db")
	if err := data.Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	db, err := data.NewConnection(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	handler, err := NewPersistent(persistentServices(t, db, storage), false)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return newAdminBrowser(t, server), db
}

func persistentServices(t *testing.T, db *sql.DB, storage files.Storage) Services {
	t.Helper()
	team, err := people.Open(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	provider, ok := storage.(blob.Provider)
	if !ok {
		local, err := disk.New(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { local.Close() })
		provider = local
	}
	meetings, err := content.NewPersistentStore(db, provider)
	if err != nil {
		t.Fatal(err)
	}
	return Services{Users: users.NewService(db, "https://example.org"), Archive: files.NewService(db, storage), People: team, Meetings: meetings, Posts: posts.NewService(db), Media: media.NewService(db, provider)}
}

func TestPersistentLoginAuthorizationAndLogout(t *testing.T) {
	b, db := identityBrowser(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("existing password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users (id,email,name,password,role,active,verified) VALUES (42,'PERSON@fsv-wh.de','Person',?,'user',1,1)`, string(hash)); err != nil {
		t.Fatal(err)
	}
	_, body, _ := b.get("/login")
	if strings.Contains(body, "Demozugang") {
		t.Fatal("persistent login advertises demo credentials")
	}
	csrf := csrfFrom(t, body)
	if status, _, _ := b.post("/login", url.Values{"email": {"person@fsv-wh.de"}, "password": {"existing password"}}); status != 403 {
		t.Fatal("login accepted missing CSRF")
	}
	if status, _, _ := b.post("/login", url.Values{"email": {"admin"}, "password": {"admin"}, "csrf": {csrf}}); status != 422 {
		t.Fatal("demo credentials accepted")
	}
	status, _, headers := b.post("/login", url.Values{"email": {"person@fsv-wh.de"}, "password": {"existing password"}, "csrf": {csrf}})
	if status != 303 || headers.Get("Location") != "/" {
		t.Fatalf("login: %d %v", status, headers)
	}
	_, body, _ = b.get("/archive")
	rotated := csrfFrom(t, body)
	if rotated == csrf {
		t.Fatal("login did not rotate session")
	}
	if strings.Contains(body, `href="/verwaltung"`) {
		t.Fatal("normal user sees administration link")
	}
	if status, _, _ := b.get("/verwaltung"); status != 403 {
		t.Fatal("normal user accessed administration")
	}
	if _, err := db.Exec(`UPDATE users SET role='admin' WHERE id=42`); err != nil {
		t.Fatal(err)
	}
	if status, body, _ := b.get("/verwaltung/benutzer"); status != 200 || !strings.Contains(body, "person@fsv-wh.de") {
		t.Fatal("live directory or updated role not used")
	}
	// Build a fresh handler and service, keeping only the database and browser cookie.
	handler, err := NewPersistent(persistentServices(t, db, nil), false)
	if err != nil {
		t.Fatal(err)
	}
	restarted := httptest.NewServer(handler)
	defer restarted.Close()
	b.server = restarted
	if status, _, _ := b.get("/verwaltung"); status != 200 {
		t.Fatal("session lost after handler restart")
	}
	if status, _, _ := b.post("/logout", url.Values{"csrf": {rotated}}); status != 303 {
		t.Fatal("logout failed")
	}
	if status, _, headers := b.get("/verwaltung"); status != 303 || headers.Get("Location") != "/login" {
		t.Fatal("logout retained administration access")
	}
}

func mailTokenPath(t *testing.T, db *sql.DB, kind string) string {
	t.Helper()
	var link string
	if err := db.QueryRow(`SELECT link FROM mail_outbox WHERE kind=?`, kind).Scan(&link); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	return u.RequestURI()
}

func TestPersistentRegistrationAndRecoveryJourney(t *testing.T) {
	b, db := identityBrowser(t)
	_, body, _ := b.get("/register")
	csrf := csrfFrom(t, body)
	form := url.Values{"name": {"Kim Beispiel"}, "email": {"KIM@studmail.w-hs.de"}, "password": {"eight 🔐🔐🔐"}, "csrf": {csrf}}
	if status, _, _ := b.post("/register", url.Values{}); status != 403 {
		t.Fatal("registration accepted missing CSRF")
	}
	if status, body, _ := b.post("/register", form); status != 200 || !strings.Contains(body, "Dein Konto wurde erstellt.") {
		t.Fatalf("registration: %d %s", status, body)
	}
	verifyPath := mailTokenPath(t, db, "verify")
	_, body, _ = b.get(verifyPath)
	var verified int
	if err := db.QueryRow(`SELECT verified FROM users WHERE email='kim@studmail.w-hs.de'`).Scan(&verified); err != nil || verified != 0 {
		t.Fatal("GET consumed verification token")
	}
	u, _ := url.Parse(verifyPath)
	if status, _, _ := b.post("/verify", url.Values{"csrf": {csrfFrom(t, body)}, "token": {u.Query().Get("token")}}); status != 200 {
		t.Fatal("verification failed")
	}
	_, body, _ = b.get("/forgot")
	if status, body, _ := b.post("/forgot", url.Values{"csrf": {csrfFrom(t, body)}, "email": {"kim@studmail.w-hs.de"}}); status != 200 || !strings.Contains(body, "Wenn ein Konto") {
		t.Fatal("reset request failed")
	}
	resetPath := mailTokenPath(t, db, "reset")
	_, body, _ = b.get(resetPath)
	u, _ = url.Parse(resetPath)
	reset := url.Values{"csrf": {csrfFrom(t, body)}, "token": {u.Query().Get("token")}, "password": {"replacement 🔐"}}
	if status, _, _ := b.post("/reset", reset); status != 200 {
		t.Fatal("reset failed")
	}
	if status, _, _ := b.post("/reset", reset); status != 422 {
		t.Fatal("reset token reused")
	}
	s := users.NewService(db, "https://example.org")
	if _, err := s.Login(context.Background(), "kim@studmail.w-hs.de", "replacement 🔐"); err != nil {
		t.Fatal(err)
	}
}
