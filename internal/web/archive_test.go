package web

import (
	"context"
	"io"
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
