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

func TestArchiveAssignmentChoicesAndPreferenceDefaults(t *testing.T) {
	storage, err := disk.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	b, db := identityBrowserWithStorage(t, storage)
	if _, err := db.Exec(`INSERT INTO programs(id,name) VALUES(1,'Informatik (B. Sc.)'),(2,'Wirtschaftsinformatik (B. Sc.)'),(3,'Informatik (M. Sc.)');
    INSERT INTO program_versions(program_id,name) VALUES(1,'PO2016'),(1,'PO2023'),(2,'PO2023'),(3,'PO2016');
    INSERT INTO modules(id,program_id,name) VALUES(10,1,'Mathematik'),(20,2,'Mathematik'),(30,3,'Mathematik');
    INSERT INTO users(id,email,name,password,active,verified,program_id) VALUES(1,'preferences@fsv-wh.de','Person','unused',1,1,1);
    INSERT INTO archive(id,group_id,user_id,module_id,version,exam_date,access_key,mime_type,nbytes,checksum) VALUES
    (100,100,1,10,'PO2016','2026-01-01','a','application/pdf',5,'one'),(200,200,1,20,'PO2023','2026-02-01','b','application/pdf',5,'two');`); err != nil {
		t.Fatal(err)
	}
	if status, _, _ := b.get("/archive/preferences"); status != 303 {
		t.Fatal("anonymous account preferences exposed")
	}
	session, err := users.NewService(db, "https://example.org", users.DomainPolicy{Registration: []string{"studmail.w-hs.de"}, Accounts: []string{"studmail.w-hs.de", "fsv-wh.de"}}).CreateSession(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	b.client.Jar.SetCookies(mustURL(t, b.server.URL), []*http.Cookie{{Name: sessionCookie, Value: session.ID, Path: "/"}})
	if status, body, _ := b.get("/archive"); status != 200 || !strings.Contains(body, `href="/archive/files/100"`) || strings.Contains(body, `href="/archive/files/200"`) {
		t.Fatal("account qualification is not applied by default")
	}
	if status, body, _ := b.get("/archive?applied=1"); status != 200 || !strings.Contains(body, "Alle 2 Rekos") {
		t.Fatal("explicitly cleared filters still apply preferences")
	}
	if status, _, _ := b.post("/archive/preferences", url.Values{}); status != 403 {
		t.Fatal("preferences accept missing CSRF")
	}
	if status, _, _ := b.post("/archive/preferences", url.Values{"csrf": {session.CSRF}, "program": {"Wirtschaftsinformatik"}, "degree": {"Bachelor"}, "version": {"PO2023"}}); status != 303 {
		t.Fatal("preferences could not be saved")
	}
	if status, body, _ := b.get("/archive"); status != 200 || !strings.Contains(body, `href="/archive/files/200"`) || strings.Contains(body, `href="/archive/files/100"`) {
		t.Fatal("stored preferences not applied")
	}
	if status, body, _ := b.get("/archive?program=Informatik&degree=Bachelor&version=PO2016"); status != 200 || !strings.Contains(body, `href="/archive/files/100"`) {
		t.Fatal("explicit filter does not override preferences")
	}
	scope := url.Values{"scope": {"1"}, "program": {"Informatik", "Wirtschaftsinformatik"}, "degree": {"Bachelor", "Master"}, "version": {"PO2016", "PO2023"}}
	if status, body, _ := b.get("/archive/new?" + scope.Encode()); status != 200 || !strings.Contains(body, `name="module" type="checkbox" value="Mathematik"`) || strings.Contains(body, `name="assignment"`) {
		t.Fatal("upload does not offer a single module choice")
	}
	scope.Set("module_selection", "1")
	scope.Set("csrf", session.CSRF)
	if status, body, _ := b.post("/archive/new", scope); status != 422 || !strings.Contains(body, "Wähle mindestens ein passendes Modul aus.") {
		t.Fatal("empty module selection accepted")
	}
	scope.Set("module", "Unknown module")
	if status, body, _ := b.post("/archive/new", scope); status != 422 || !strings.Contains(body, "Wähle gültige Module aus.") {
		t.Fatal("unknown module accepted")
	}
	scope.Set("module", "Mathematik")
	var buffer bytes.Buffer
	form := multipart.NewWriter(&buffer)
	scope.Set("csrf", session.CSRF)
	scope.Set("date-day", "27")
	scope.Set("date-month", "3")
	scope.Set("date-year", "2026")
	for key, values := range scope {
		for _, value := range values {
			if err := form.WriteField(key, value); err != nil {
				t.Fatal(err)
			}
		}
	}
	file, err := form.CreateFormFile("files", "one.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(file, "%PDF-1.7\none shared PDF\n%%EOF"); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	status, _, headers := b.request("POST", "/archive/new", form.FormDataContentType(), &buffer)
	if status != 303 {
		t.Fatal("multiple assignments could not be submitted")
	}
	review := headers.Get("Location")
	requestID := strings.TrimSuffix(strings.TrimPrefix(review, "/archive/requests/"), "/review")
	rows, err := db.Query(`SELECT p.module_id,p.version FROM archive a JOIN archive_placements p ON p.archive_id=a.id WHERE a.request_id=? ORDER BY p.module_id,p.version`, requestID)
	if err != nil {
		t.Fatal(err)
	}
	var placements []string
	for rows.Next() {
		var module, version string
		if err := rows.Scan(&module, &version); err != nil {
			t.Fatal(err)
		}
		placements = append(placements, module+"|"+version)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	if strings.Join(placements, ",") != "10|PO2016,10|PO2023,20|PO2023,30|PO2016" {
		t.Fatalf("module scope produced invalid assignments: %v", placements)
	}
	if status, body, _ := b.get(review); status != 200 || !strings.Contains(body, "Wirtschaftsinformatik") || !strings.Contains(body, "Master") || !strings.Contains(body, "Prüfungsordnungen") || strings.Contains(body, "Reko 2") {
		t.Fatal("check answers does not show the actual placements for one PDF")
	}
	edit := strings.TrimSuffix(review, "/review") + "/edit"
	if status, body, _ := b.get(edit); status != 200 || strings.Count(body, `name="module" type="checkbox"`) == 0 || strings.Count(body, `checked`) != 1 || strings.Contains(body, `name="assignment"`) {
		t.Fatal("editing lost module selection or reintroduced tuple choices")
	}
}
