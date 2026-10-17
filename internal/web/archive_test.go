package web

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/data/blob/disk"
	"github.com/fachschaftinformatik/web/internal/users"
)

func TestArchiveDownloadRequiresLiveAuthenticatedSession(t *testing.T) {
	storage, err := disk.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	b, db := identityBrowserWithStorage(t, storage)
	_, err = db.Exec(`
INSERT INTO programs (id,name) VALUES (1,'Informatik (B. Sc.)');
INSERT INTO program_versions (program_id,name) VALUES (1,'2023');
INSERT INTO modules (id,program_id,name) VALUES (10,1,'Datenbanken');
INSERT INTO users (id,email,name,password,active,verified) VALUES (1,'reader@fsv-wh.de','Person','unused',1,1);
INSERT INTO archive (id,group_id,user_id,module_id,version,exam_date,access_key,mime_type,nbytes,checksum) VALUES (100,100,1,10,'2023','2026-01-01','key','application/pdf',9,'checksum');`)
	if err != nil {
		t.Fatal(err)
	}
	if err := storage.Upload(context.Background(), "archive/100/source.pdf", strings.NewReader("%PDF-test"), 9, "application/pdf"); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/archive/files/100", "/api/v1/archive/100/file"} {
		if status, _, headers := b.get(path); status != 303 || headers.Get("Location") != "/login" || headers.Get("Cache-Control") != "private, no-store" {
			t.Fatal("anonymous PDF download allowed or cacheable")
		}
	}
	session, err := users.NewService(db, "http://example.org").CreateSession(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	base, _ := url.Parse(b.server.URL)
	b.client.Jar.SetCookies(base, []*http.Cookie{{Name: sessionCookie, Value: session.ID, Path: "/"}})
	request, _ := http.NewRequest("GET", b.server.URL+"/archive/files/100", nil)
	request.Header.Set("Range", "bytes=0-4")
	response, err := b.client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	value, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil || response.StatusCode != 206 || string(value) != "%PDF-" || response.Header.Get("Content-Type") != "application/pdf" {
		t.Fatal("authenticated PDF range download failed")
	}
	if status, body, _ := b.request("HEAD", "/archive/files/100", "", nil); status != 200 || body != "" {
		t.Fatal("PDF HEAD failed")
	}
	if status, body, _ := b.get("/archive"); status != 200 || !strings.Contains(body, `href="/archive/files/100"`) || strings.Contains(body, "example.pdf") {
		t.Fatal("archive still uses example downloads")
	}
	if _, err := db.Exec(`UPDATE users SET active=0 WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	if status, _, _ := b.get("/archive/files/100"); status != 303 {
		t.Fatal("deactivated user retained PDF access")
	}
}

func TestAuthenticatedMultiplePDFSubmission(t *testing.T) {
	storage, err := disk.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	b, db := identityBrowserWithStorage(t, storage)
	if _, err := db.Exec(`
INSERT INTO programs(id,name) VALUES (1,'Informatik (B. Sc.)');
INSERT INTO program_versions(program_id,name) VALUES (1,'2023');
INSERT INTO modules(id,program_id,name) VALUES (10,1,'Datenbanken');
INSERT INTO users(id,email,name,password,active,verified) VALUES (1,'author@fsv-wh.de','Person','unused',1,1);`); err != nil {
		t.Fatal(err)
	}
	if status, _, _ := b.get("/archive/new"); status != 303 {
		t.Fatal("anonymous upload form allowed")
	}
	session, err := users.NewService(db, "http://example.org").CreateSession(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	base, _ := url.Parse(b.server.URL)
	b.client.Jar.SetCookies(base, []*http.Cookie{{Name: sessionCookie, Value: session.ID, Path: "/"}})
	status, body, _ := b.get("/archive/new")
	if status != 200 {
		t.Fatal("upload form unavailable")
	}
	csrf := csrfFrom(t, body)
	var reviewPath string
	for _, value := range []string{"", csrf} {
		var buffer bytes.Buffer
		form := multipart.NewWriter(&buffer)
		for name, field := range map[string]string{"csrf": value, "module_id": "10", "version": "2023", "date": "12.02.2026"} {
			if err := form.WriteField(name, field); err != nil {
				t.Fatal(err)
			}
		}
		for _, filename := range []string{"eins.pdf", "zwei.pdf"} {
			part, err := form.CreateFormFile("files", filename)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.WriteString(part, "%PDF-1.7\n"+filename+"\n%%EOF"); err != nil {
				t.Fatal(err)
			}
		}
		if err := form.Close(); err != nil {
			t.Fatal(err)
		}
		status, body, headers := b.request("POST", "/archive/new", form.FormDataContentType(), &buffer)
		if value == "" {
			if status != 403 {
				t.Fatal("upload accepted without CSRF")
			}
		} else {
			reviewPath = headers.Get("Location")
			if status != 303 || !strings.HasPrefix(reviewPath, "/archive/requests/") || !strings.HasSuffix(reviewPath, "/review") {
				t.Fatalf("multiple upload: %d %s", status, body)
			}
		}
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM archive WHERE is_latest=1`).Scan(&count); err != nil || count != 0 {
		t.Fatal("draft PDFs were immediately published")
	}
	if status, body, _ := b.get(reviewPath); status != 200 || !strings.Contains(body, "Prüfe deine Angaben") || !strings.Contains(body, "Reko 2") {
		t.Fatal("check-answers review missing")
	}
	if status, _, _ := b.post(reviewPath, url.Values{"csrf": {csrf}}); status != 303 {
		t.Fatal("user confirmation failed")
	}
	requestID := strings.TrimSuffix(strings.TrimPrefix(reviewPath, "/archive/requests/"), "/review")
	adminPath := "/verwaltung/einreichungen/" + requestID
	if status, _, _ := b.get(adminPath); status != 403 {
		t.Fatal("ordinary user accessed administration review")
	}
	if _, err := db.Exec(`UPDATE users SET role='editor' WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	if status, body, _ := b.get("/verwaltung/einreichungen"); status != 200 || !strings.Contains(body, requestID) {
		t.Fatal("editor cannot find submitted request")
	}
	if status, _, _ := b.get("/verwaltung/benutzer"); status != 403 {
		t.Fatal("review permission granted user administration")
	}
	if status, _, _ := b.post(adminPath, url.Values{"decision": {"approve"}}); status != 403 {
		t.Fatal("approval accepted without CSRF")
	}
	if status, _, _ := b.post(adminPath, url.Values{"csrf": {csrf}, "decision": {"approve"}}); status != 303 {
		t.Fatal("editor approval failed")
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM archive WHERE is_latest=1`).Scan(&count); err != nil || count != 2 {
		t.Fatal("confirmed PDFs were not published")
	}
	if status, body, _ := b.get("/archive/10"); status != 200 || !strings.Contains(body, "2 Rekos") || !strings.Contains(body, "Neue Revision einreichen") {
		t.Fatal("uploaded PDFs not available in module view")
	}
}
