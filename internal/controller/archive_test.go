package controller

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/model/users"
	"github.com/fachschaftinformatik/web/internal/storage/disk"
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
	session, err := users.NewService(db, "http://example.org", users.DomainPolicy{Registration: []string{"studmail.w-hs.de"}, Accounts: []string{"studmail.w-hs.de", "fsv-wh.de"}}).CreateSession(context.Background(), 1)
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

func TestAuthenticatedSinglePDFSubmission(t *testing.T) {
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
INSERT INTO programs(id,name) VALUES (2,'Medieninformatik (B. Sc.)');
INSERT INTO program_versions(program_id,name) VALUES (2,'2025');
INSERT INTO modules(id,program_id,name) VALUES (20,2,'Fremdes Modul');
INSERT INTO users(id,email,name,password,active,verified) VALUES (1,'author@fsv-wh.de','Person','unused',1,1);`); err != nil {
		t.Fatal(err)
	}
	if status, _, _ := b.get("/archive/new"); status != 303 {
		t.Fatal("anonymous upload form allowed")
	}
	session, err := users.NewService(db, "http://example.org", users.DomainPolicy{Registration: []string{"studmail.w-hs.de"}, Accounts: []string{"studmail.w-hs.de", "fsv-wh.de"}}).CreateSession(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	base, _ := url.Parse(b.server.URL)
	b.client.Jar.SetCookies(base, []*http.Cookie{{Name: sessionCookie, Value: session.ID, Path: "/"}})
	if status, body, _ := b.get("/archive/new"); status != 200 || !strings.Contains(body, "Studiengänge") {
		t.Fatal("program selection is missing")
	}
	status, body, _ := b.get("/archive/new?scope=1&program=Informatik&degree=Bachelor&version=2023")
	if status != 200 {
		t.Fatal("upload form unavailable")
	}
	csrf := csrfFrom(t, body)
	if strings.Contains(body, "Fremdes Modul") || strings.Contains(body, `value="2025"`) {
		t.Fatal("upload catalog is not scoped to the selected program")
	}
	if status, body, _ := b.post("/archive/new", url.Values{"csrf": {csrf}, "scope": {"1"}, "program": {"Informatik"}, "degree": {"Bachelor"}, "version": {"2023"}}); status != 422 || strings.Index(body, `href="#module_id"`) > strings.Index(body, `href="#date-day"`) {
		t.Fatal("upload errors do not follow field order")
	}
	var reviewPath string
	for _, value := range []string{"", csrf + "-multiple", csrf} {
		var buffer bytes.Buffer
		form := multipart.NewWriter(&buffer)
		for name, field := range map[string]string{"csrf": strings.TrimSuffix(value, "-multiple"), "module_id": "10", "version": "2023", "date": "12.02.2026"} {
			if err := form.WriteField(name, field); err != nil {
				t.Fatal(err)
			}
		}
		filenames := []string{"eins.pdf"}
		if strings.HasSuffix(value, "-multiple") {
			filenames = append(filenames, "zwei.pdf")
		}
		for _, filename := range filenames {
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
		} else if strings.HasSuffix(value, "-multiple") {
			if status != 422 || !strings.Contains(body, "Wähle genau eine Datei aus.") {
				t.Fatal("multiple PDF upload accepted")
			}
			var count int
			if err := db.QueryRow("SELECT count(*) FROM archive").Scan(&count); err != nil || count != 0 {
				t.Fatal("rejected multi-file upload wrote records")
			}
		} else {
			reviewPath = headers.Get("Location")
			if status != 303 || !strings.HasPrefix(reviewPath, "/archive/requests/") || !strings.HasSuffix(reviewPath, "/review") {
				t.Fatalf("single upload: %d %s", status, body)
			}
		}
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM archive WHERE is_latest=1`).Scan(&count); err != nil || count != 0 {
		t.Fatal("draft PDFs were immediately published")
	}
	if status, body, _ := b.get(reviewPath); status != 200 || !strings.Contains(body, "Prüfe deine Angaben") || strings.Contains(body, "Reko 2") {
		t.Fatal("check-answers review missing")
	}
	// Enhanced uploads must redirect with a 200 response: browsers follow a 303
	// before HTMX can inspect its HX-Redirect header.
	var enhanced bytes.Buffer
	writer := multipart.NewWriter(&enhanced)
	for key, value := range map[string]string{"csrf": csrf, "module_id": "10", "version": "2023", "date": "27.03.2026"} {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	part, err := writer.CreateFormFile("files", "enhanced.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(part, "%PDF-1.7\n%%EOF"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", b.server.URL+"/archive/new", &enhanced)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("HX-Request", "true")
	response, err := b.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 200 || !strings.HasSuffix(response.Header.Get("HX-Redirect"), "/review") {
		t.Fatal("enhanced upload does not redirect to the saved draft")
	}
	if status, _, _ := b.post(reviewPath, url.Values{"csrf": {csrf}}); status != 303 {
		t.Fatal("user confirmation failed")
	}
	requestID := strings.TrimSuffix(strings.TrimPrefix(reviewPath, "/archive/requests/"), "/review")
	adminPath := "/verwaltung/einreichungen/" + requestID
	if status, _, _ := b.get(adminPath); status != 403 {
		t.Fatal("ordinary user accessed administration review")
	}
	if status, _, _ := b.get(adminPath + "/edit"); status != 403 {
		t.Fatal("ordinary uploader accessed reviewed editing")
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
	if err := db.QueryRow(`SELECT COUNT(*) FROM archive WHERE is_latest=1`).Scan(&count); err != nil || count != 1 {
		t.Fatal("confirmed PDFs were not published")
	}
	if status, body, _ := b.get("/archive/10"); status != 200 || !htmlContains(body, "1 Reko") || !htmlContains(body, "Neue Revision einreichen") {
		t.Fatal("uploaded PDFs not available in module view")
	}
	for _, decision := range []string{"reject", "approve"} {
		if status, body, _ := b.get(adminPath); status != 200 || !htmlContains(body, `name="decision" value="`+decision+`"`) {
			t.Fatal("completed review does not offer the opposite decision")
		}
		if status, _, _ := b.post(adminPath, url.Values{"decision": {decision}}); status != 403 {
			t.Fatal("decision reversal accepted without CSRF")
		}
		if status, _, _ := b.post(adminPath, url.Values{"csrf": {csrf}, "decision": {decision}}); status != 303 {
			t.Fatal("decision could not be reversed")
		}
		want := 0
		if decision == "approve" {
			want = 1
		}
		if err := db.QueryRow(`SELECT count(*) FROM archive WHERE is_latest=1`).Scan(&count); err != nil || count != want {
			t.Fatal("reversed decision did not update public archive")
		}
	}
	status, body, _ = b.get(adminPath + "/edit")
	if status != 200 {
		t.Fatal("reviewer cannot edit approved submission")
	}
	_, tokenPart, ok := strings.Cut(body, `name="edit_token" value="`)
	if !ok {
		t.Fatal("missing edit concurrency token")
	}
	token, _, _ := strings.Cut(tokenPart, `"`)
	edit := url.Values{"scope": {"1"}, "module_selection": {"1"}, "module": {"Datenbanken"}, "degree": {"Bachelor"}, "program": {"Informatik"}, "version": {"2023"}, "date-day": {"27"}, "date-month": {"3"}, "date-year": {"2026"}, "comment": {"Redaktionell korrigiert"}, "edit_token": {token}}
	if status, _, _ := b.post(adminPath+"/edit", edit); status != 403 {
		t.Fatal("reviewer edit accepted without CSRF")
	}
	edit.Set("csrf", csrf)
	if status, _, headers := b.post(adminPath+"/edit", edit); status != 303 || headers.Get("Location") != adminPath {
		t.Fatal("reviewer edit did not return to review")
	}
	if status, body, _ := b.get(adminPath); status != 200 || !strings.Contains(body, "Redaktionell korrigiert") || !strings.Contains(body, adminPath+"/edit#comment") || strings.Contains(body, "govuk-button--warning") {
		t.Fatal("reviewed corrections or neutral actions missing")
	}
	if status, _, _ := b.post(adminPath+"/edit", edit); status != 409 {
		t.Fatal("stale reviewer edit accepted")
	}
}
