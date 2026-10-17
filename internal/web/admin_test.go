package web

import (
	"bytes"
	"fmt"
	"github.com/fachschaftinformatik/web/internal/users"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

type adminBrowser struct {
	t      *testing.T
	server *httptest.Server
	client *http.Client
	csrf   string
}

func newAdminBrowser(t *testing.T, server *httptest.Server) *adminBrowser {
	t.Helper()
	jar, _ := cookiejar.New(nil)
	return &adminBrowser{t: t, server: server, client: &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func (b *adminBrowser) request(method, path, contentType string, body io.Reader) (int, string, http.Header) {
	b.t.Helper()
	r, err := http.NewRequest(method, b.server.URL+path, body)
	if err != nil {
		b.t.Fatal(err)
	}
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	response, err := b.client.Do(r)
	if err != nil {
		b.t.Fatal(err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		b.t.Fatal(err)
	}
	return response.StatusCode, string(data), response.Header
}
func (b *adminBrowser) get(path string) (int, string, http.Header) {
	return b.request("GET", path, "", nil)
}
func (b *adminBrowser) post(path string, form url.Values) (int, string, http.Header) {
	return b.request("POST", path, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
}
func csrfFrom(t *testing.T, body string) string {
	t.Helper()
	match := regexp.MustCompile(`name="csrf" value="([a-f0-9]+)"`).FindStringSubmatch(body)
	if match == nil {
		t.Fatal("missing CSRF token")
	}
	return match[1]
}
func (b *adminBrowser) login() {
	b.t.Helper()
	_, body, _ := b.get("/login")
	token := csrfFrom(b.t, body)
	status, _, headers := b.post("/login", url.Values{"email": {"admin"}, "password": {"admin"}, "csrf": {token}})
	if status != 303 || headers.Get("Location") != "/verwaltung" {
		b.t.Fatalf("login failed: %d", status)
	}
	_, body, _ = b.get("/verwaltung")
	b.csrf = csrfFrom(b.t, body)
	if b.csrf == token {
		b.t.Fatal("session not rotated at login")
	}
}

func TestAdminAuthenticationAndMutations(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	b := newAdminBrowser(t, server)
	for _, path := range []string{"/verwaltung", "/verwaltung/benutzer", "/verwaltung/mitglieder/1", "/verwaltung/sitzungen/neu"} {
		if status, _, headers := b.get(path); status != 303 || headers.Get("Location") != "/login" {
			t.Fatalf("unprotected route: %s", path)
		}
	}
	_, body, _ := b.get("/login")
	token := csrfFrom(t, body)
	if status, _, _ := b.post("/login", url.Values{"email": {"admin"}, "password": {"wrong"}, "csrf": {token}}); status != 422 {
		t.Fatal("wrong credentials accepted")
	}
	if status, _, _ := b.post("/login", url.Values{"email": {"admin"}, "password": {"admin"}}); status != 403 {
		t.Fatal("login CSRF accepted")
	}
	b.login()
	if status, body, headers := b.get("/verwaltung/benutzer?q=kim"); status != 200 || !strings.Contains(body, "Kim Muster") || strings.Contains(body, "Alex Beispiel") || headers.Get("Cache-Control") != "no-store" {
		t.Fatal("user search/cache control failed")
	}
	if status, _, _ := b.get("/verwaltung/benutzer/missing"); status != 404 {
		t.Fatal("unknown user accepted")
	}
	if status, _, _ := b.post("/verwaltung/mitglieder/1", url.Values{"name": {"Changed"}}); status != 403 {
		t.Fatal("write without CSRF accepted")
	}
	if status, body, _ := b.post("/verwaltung/mitglieder/1", url.Values{"csrf": {b.csrf}, "name": {""}, "email": {"not-email"}}); status != 422 || !strings.Contains(body, "govuk-error-summary") {
		t.Fatal("invalid member accepted")
	}
	if status, _, _ := b.post("/verwaltung/mitglieder/1", url.Values{"csrf": {b.csrf}, "name": {"Demo-Vorsitz"}, "role": {"Vorsitz"}, "email": {"demo@example.org"}}); status != 303 {
		t.Fatal("member edit failed")
	}
	if _, body, _ := b.get("/team"); !strings.Contains(body, "Demo-Vorsitz") {
		t.Fatal("public team did not reflect edit")
	}
	if status, _, _ := b.post("/verwaltung/referate/neu", url.Values{"csrf": {b.csrf}, "title": {"Demo-Referat"}, "members": {"1"}}); status != 303 {
		t.Fatal("responsibility add failed")
	}
	if _, body, _ := b.get("/team"); !strings.Contains(body, "Demo-Referat") {
		t.Fatal("public responsibility missing")
	}
	oldCookies := b.client.Jar.Cookies(mustURL(t, server.URL))
	if status, _, _ := b.post("/logout", url.Values{"csrf": {b.csrf}}); status != 303 {
		t.Fatal("logout failed")
	}
	b.client.Jar.SetCookies(mustURL(t, server.URL), oldCookies)
	if status, _, _ := b.get("/verwaltung"); status != 303 {
		t.Fatal("logged-out session replayed")
	}
}

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func (b *adminBrowser) upload(path string, fields map[string]string, data []byte) (int, string, http.Header) {
	b.t.Helper()
	var buffer bytes.Buffer
	w := multipart.NewWriter(&buffer)
	for key, value := range fields {
		if err := w.WriteField(key, value); err != nil {
			b.t.Fatal(err)
		}
	}
	if data != nil {
		file, err := w.CreateFormFile("file", "beispiel.pdf")
		if err != nil {
			b.t.Fatal(err)
		}
		file.Write(data)
	}
	w.Close()
	return b.request("POST", path, w.FormDataContentType(), &buffer)
}

func TestPublicationReviewAndVisibility(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	b := newAdminBrowser(t, server)
	b.login()
	other := newAdminBrowser(t, server)
	other.login()
	pdf, err := resources.ReadFile("assets/example.pdf")
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]string{"csrf": b.csrf, "day": "28", "month": "10", "year": "2027", "number": "45", "kind": "Einladung"}
	if status, _, _ := b.upload("/verwaltung/sitzungen/neu", fields, []byte("not a PDF")); status != 422 {
		t.Fatal("invalid file accepted")
	}
	fields["day"] = "32"
	status, body, _ := b.upload("/verwaltung/sitzungen/neu", fields, pdf)
	if status != 422 || !strings.Contains(body, "beibehalten") {
		t.Fatal("file was not retained through validation")
	}
	match := regexp.MustCompile(`id="admin-form" action="([^"]+)"`).FindStringSubmatch(body)
	if match == nil {
		t.Fatal("draft form missing")
	}
	editPath := match[1]
	fields["day"] = "28"
	status, _, headers := b.upload(editPath, fields, nil)
	if status != 303 {
		t.Fatalf("valid draft status %d", status)
	}
	review := headers.Get("Location")
	if status, body, _ := b.get(review); status != 200 || !strings.Contains(body, "Veröffentlichung prüfen") {
		t.Fatal("review unavailable")
	}
	document := "/documents/" + strings.TrimPrefix(editPath, "/verwaltung/sitzungen/")
	if status, _, _ := other.get(review); status != 404 {
		t.Fatal("another admin can access private draft")
	}
	if status, _, _ := other.get(document); status != 404 {
		t.Fatal("another admin can access draft PDF")
	}
	if status, _, _ := b.get(document); status != 200 {
		t.Fatal("owner cannot preview draft")
	}
	if _, body, _ := other.get("/fachschaft/sitzungen"); strings.Contains(body, document) {
		t.Fatal("draft published before confirmation")
	}
	if status, _, _ := b.post(review, url.Values{"csrf": {b.csrf}}); status != 303 {
		t.Fatal("publication failed")
	}
	if status, body, _ := other.get("/fachschaft/sitzungen?jahr=2027"); status != 200 || !strings.Contains(body, document) {
		t.Fatal("published document missing from public archive")
	}
	guest := newAdminBrowser(t, server)
	if status, body, headers := guest.get(document); status != 200 || body != string(pdf) || headers.Get("Content-Type") != "application/pdf" {
		t.Fatal("published PDF not publicly readable")
	}
	// Repeated publish is idempotent rather than adding another document.
	_, body, _ = guest.get("/fachschaft/sitzungen?jahr=2027")
	count := strings.Count(body, document)
	b.post(review, url.Values{"csrf": {b.csrf}})
	_, body, _ = guest.get("/fachschaft/sitzungen?jahr=2027")
	if count == 0 || strings.Count(body, document) != count {
		t.Fatal("duplicate publication")
	}
}

func TestPasswordResetConfirmationPreview(t *testing.T) {
	handler, err := NewWithUsers([]users.User{{ID: "42", Name: "Kim Beispiel", Email: "Kim@STUDMAIL.W-HS.DE", Permission: "FSV", Status: "Aktiv"}})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	b := newAdminBrowser(t, server)
	path := "/verwaltung/benutzer/42/passwort-zuruecksetzen"
	if status, _, _ := b.get(path); status != 303 {
		t.Fatal("unprotected password reset")
	}
	b.login()
	if status, body, _ := b.get("/verwaltung/benutzer/42"); status != 200 || !strings.Contains(body, "kim@studmail.w-hs.de") || strings.Contains(body, "Kim@STUDMAIL") || !strings.Contains(body, `<dialog`) {
		t.Fatal("user email normalization/dialog failed")
	}
	if status, body, _ := b.get(path); status != 200 || !strings.Contains(body, "Link zum Zurücksetzen senden") {
		t.Fatal("no-JS confirmation missing")
	}
	if status, _, _ := b.post(path, url.Values{}); status != 403 {
		t.Fatal("reset without CSRF accepted")
	}
	status, body, _ := b.post(path, url.Values{"csrf": {b.csrf}})
	if status != 503 || !strings.Contains(body, "Die Anfrage konnte nicht verarbeitet werden.") {
		t.Fatal("reset without a mail service must fail")
	}
	if status, _, _ := b.get("/verwaltung/benutzer/missing/passwort-zuruecksetzen"); status != 404 {
		t.Fatal("unknown reset user accepted")
	}
}

func TestUserPaginationKeepsSearchContext(t *testing.T) {
	var records []users.User
	for i := 1; i <= 45; i++ {
		records = append(records, users.User{ID: fmt.Sprint(i), Name: fmt.Sprintf("Beispiel %02d", i), Email: fmt.Sprintf("user%d@example.org", i), Verified: i%2 == 0, Status: "Aktiv"})
	}
	handler, err := NewWithUsers(records)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	b := newAdminBrowser(t, server)
	b.login()
	status, body, _ := b.get("/verwaltung/benutzer?q=Beispiel&seite=2")
	if status != 200 || !strings.Contains(body, "Benutzer – Seite 2 von 3") || !strings.Contains(body, `aria-current="page">2`) || !strings.Contains(body, `q=Beispiel&amp;seite=3`) || !strings.Contains(body, "Beispiel 21") || strings.Contains(body, "Beispiel 01</a>") {
		t.Fatal("numbered pagination lost page/search context")
	}
	for _, value := range []string{"0", "999", "invalid"} {
		status, _, headers := b.get("/verwaltung/benutzer?q=Beispiel&seite=" + value)
		if status != 303 || headers.Get("Location") != "/verwaltung/benutzer?q=Beispiel" {
			t.Fatal("invalid page did not redirect to first search page")
		}
	}
	_, body, _ = b.get("/verwaltung/benutzer?q=absent")
	if strings.Contains(body, `aria-label="Benutzerseiten"`) {
		t.Fatal("pagination shown for empty results")
	}
}
