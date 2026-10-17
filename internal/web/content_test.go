package web

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/data"
	"github.com/fachschaftinformatik/web/internal/data/blob"
	"github.com/fachschaftinformatik/web/internal/users"
)

type meetingRetryBlob struct {
	blob.Provider
	fail atomic.Bool
}

func (b *meetingRetryBlob) Upload(_ context.Context, _ string, r io.Reader, _ int64, _ string) error {
	if b.fail.Load() {
		return errors.New("Datei konnte nicht gespeichert werden.")
	}
	_, err := io.Copy(io.Discard, r)
	return err
}
func (b *meetingRetryBlob) Delete(context.Context, string) error { return nil }

func TestMeetingInitialSaveFailureCanRetry(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "meetings.sqlite")
	if err := data.Migrate(dsn, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	storage := &meetingRetryBlob{}
	storage.fail.Store(true)
	store, err := content.NewPersistentStore(db, storage)
	if err != nil {
		t.Fatal(err)
	}
	views, err := newRenderer()
	if err != nil {
		t.Fatal(err)
	}
	c := &Handler{content: store, views: views, identity: &users.Service{}}
	router := http.NewServeMux()
	router.HandleFunc("POST /verwaltung/sitzungen/{id}", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(uploadBodyLimit); err != nil {
			t.Error(err)
			return
		}
		defer r.MultipartForm.RemoveAll()
		c.EditPublication(&sessionResponse{ResponseWriter: w, session: &session{ID: "cookie", UserID: 42, Authenticated: true, Role: "admin"}}, r, &AdminPage{Section: "sitzungen", SectionTitle: "Sitzungen"})
	})
	server := httptest.NewServer(router)
	defer server.Close()
	b := newAdminBrowser(t, server)
	fields := map[string]string{"day": "28", "month": "10", "year": "2028", "number": "45", "kind": "Einladung"}
	pdf := []byte("%PDF-1.4\nmeeting\n%%EOF")
	status, body, _ := b.upload("/verwaltung/sitzungen/neu", fields, pdf)
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("failed save status %d", status)
	}
	if strings.Contains(body, "wird diese Datei beibehalten") || !strings.Contains(body, "Die Datei wurde nicht gespeichert.") {
		t.Fatal("failed initial save claimed to retain the upload")
	}
	action := regexp.MustCompile(`id="admin-form" action="([^"]+)"`).FindStringSubmatch(body)
	if action == nil || action[1] != "/verwaltung/sitzungen/neu" {
		t.Fatalf("failed save left nonretryable action: %v", action)
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM meeting_publications").Scan(&count); err != nil || count != 0 {
		t.Fatal("failed save committed a draft", count, err)
	}
	storage.fail.Store(false)
	status, _, headers := b.upload(action[1], fields, pdf)
	if status != http.StatusSeeOther || !strings.HasSuffix(headers.Get("Location"), "/pruefen") {
		t.Fatalf("retry failed: %d %s", status, headers.Get("Location"))
	}
	if err := db.QueryRow("SELECT count(*) FROM meeting_publications").Scan(&count); err != nil || count != 1 {
		t.Fatal("retry did not save exactly one draft", count, err)
	}
}

func TestMeetingDocuments(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	files, sources := map[string]bool{}, map[string]bool{}
	for _, meeting := range content.All() {
		for _, document := range meeting.Documents {
			if files[document.File] {
				t.Fatalf("duplicate document URL: %s", document.File)
			}
			files[document.File] = true
			for _, source := range document.Sources {
				if sources[source] {
					t.Fatalf("duplicate source mapping: %s", source)
				}
				sources[source] = true
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", document.URL(), nil))
			content := response.Body.Bytes()
			if response.Code != 200 || !strings.HasPrefix(response.Header().Get("Content-Type"), "application/pdf") {
				t.Fatalf("PDF not served: %s (%d)", document.File, response.Code)
			}
			if len(content) != document.Bytes || fmt.Sprintf("%x", sha256.Sum256(content)) != document.SHA256 {
				t.Fatalf("source integrity mismatch: %s", document.File)
			}
		}
	}
	if len(files) != 48 || len(sources) != 49 {
		t.Fatalf("import coverage: %d files, %d sources", len(files), len(sources))
	}
}

func TestMeetingDraftUsesStableAccountOwner(t *testing.T) {
	c := &Handler{content: content.NewStore(), identity: &users.Service{}}
	p := content.Publication{ID: "draft", Owner: "42", Date: "2028-10-28", Number: 45, Kind: "Einladung", Filename: "einladung.pdf", Data: []byte("%PDF-1.4\nmeeting\n%%EOF")}
	if err := c.content.Save(p); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		session *session
		status  int
	}{
		{"rotated login", &session{ID: "new-cookie-token", UserID: 42, Authenticated: true}, 200},
		{"other account", &session{ID: "another-cookie-token", UserID: 43, Authenticated: true}, 404},
		{"guest", nil, 404},
		{"missing account", &session{ID: "42", Authenticated: true}, 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/documents/draft", nil)
			r.SetPathValue("id", p.ID)
			recorder := httptest.NewRecorder()
			w := &sessionResponse{ResponseWriter: recorder, session: tc.session}
			c.Document(w, r)
			if recorder.Code != tc.status {
				t.Fatalf("download status %d, want %d", recorder.Code, tc.status)
			}
			if tc.status == 200 && recorder.Body.String() != string(p.Data) {
				t.Fatal("draft bytes changed")
			}
		})
	}
	r := httptest.NewRequest(http.MethodPost, "/verwaltung/sitzungen/draft/pruefen", nil)
	r.SetPathValue("id", p.ID)
	recorder := httptest.NewRecorder()
	w := &sessionResponse{ResponseWriter: recorder, session: &session{ID: "rotated-token", UserID: 42, Authenticated: true}}
	c.ReviewPublication(w, r)
	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("stable owner could not publish: %d", recorder.Code)
	}
	r = httptest.NewRequest(http.MethodGet, "/documents/draft", nil)
	r.SetPathValue("id", p.ID)
	recorder = httptest.NewRecorder()
	c.Document(&sessionResponse{ResponseWriter: recorder}, r)
	if recorder.Code != http.StatusOK {
		t.Fatal("published document is not public")
	}
}

func TestMeetingYears(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		query, want string
		status      int
	}{
		{"", "24. September 2026", 200},
		{"?jahr=2024", "10. Dezember 2024", 200},
		{"?jahr=2025", "Datum uneindeutig", 200},
		{"?jahr=2022", "2022", 200},
		{"?jahr=2020", "6. März 2020", 200},
		{"?jahr=2021", "22. Januar 2021", 200},
		{"?jahr=2023", "18. Januar 2023", 200},
		{"?jahr=1999", "Seite nicht gefunden", 404},
		{"?jahr=invalid", "Seite nicht gefunden", 404},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/fachschaft/sitzungen"+tc.query, nil))
		if response.Code != tc.status || !strings.Contains(response.Body.String(), tc.want) {
			t.Fatalf("%s: %d %s", tc.query, response.Code, response.Body.String())
		}
	}
}
