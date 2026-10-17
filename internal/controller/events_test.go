package controller

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/view"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/model/events"
	"github.com/fachschaftinformatik/web/internal/model/forum"
	"github.com/fachschaftinformatik/web/internal/model/users"
	"github.com/fachschaftinformatik/web/internal/storage/disk"
)

func TestEventPublicationWorkflow(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "events.db")
	if err := database.Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	db, err := database.NewConnection(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(`INSERT INTO users(id,email,name,password,active,verified,role) VALUES(1,'editor@fsv-wh.de','Editor','unused',1,1,'editor'),(2,'member@fsv-wh.de','Member','unused',1,1,'user')`); err != nil {
		t.Fatal(err)
	}
	storage, err := disk.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	views, err := view.NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	c := &Handler{views: views, identity: users.NewService(db, "http://example.test", users.DomainPolicy{Registration: []string{"studmail.w-hs.de"}, Accounts: []string{"studmail.w-hs.de", "fsv-wh.de"}}), events: events.NewService(db, storage), forum: forum.NewService(db)}
	mux := http.NewServeMux()
	mux.HandleFunc("/community/veranstaltungen/new", c.EditEvent)
	mux.HandleFunc("/community/veranstaltungen/{event}/edit", c.EditEvent)
	mux.HandleFunc("/community/veranstaltungen/{event}/cover", c.CalendarCover)
	mux.HandleFunc("/community/veranstaltungen/{event}", c.Events)
	mux.HandleFunc("/community/veranstaltungen", c.Events)
	mux.HandleFunc("/community/neuigkeiten/new", c.EditDiscussion)
	mux.HandleFunc("/community/neuigkeiten", c.News)
	mux.HandleFunc("/discussions", c.Forum)
	handler := c.withSession(mux)
	staff, err := c.identity.CreateSession(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	member, err := c.identity.CreateSession(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	request := func(method, path string, s *session, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
		t.Helper()
		var reader io.Reader
		if body != nil {
			reader = body
		}
		r := httptest.NewRequest(method, path, reader)
		if s != nil {
			r.AddCookie(&http.Cookie{Name: sessionCookie, Value: s.ID})
		}
		if contentType != "" {
			r.Header.Set("Content-Type", contentType)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	for _, tc := range []struct {
		s    *session
		code int
	}{{nil, 303}, {member, 403}, {staff, 200}} {
		w := request("GET", view.EventsURL+"/new", tc.s, nil, "")
		if w.Code != tc.code {
			t.Fatalf("editor access: %d %s", w.Code, w.Body.String())
		}
	}
	upload := func(csrf string) (*bytes.Buffer, string) {
		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		for key, value := range map[string]string{"csrf": csrf, "title": "Sommergrillen", "body": "Gemeinsam im Innenhof grillen.", "date": "30.06.2027", "time": "17:00", "location": "Innenhof", "cover_alt": "Plakat zum Sommergrillen", "type": "discussion"} {
			if err := form.WriteField(key, value); err != nil {
				t.Fatal(err)
			}
		}
		file, err := form.CreateFormFile("cover", "poster.png")
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(file, image.NewRGBA(image.Rect(0, 0, 120, 200))); err != nil {
			t.Fatal(err)
		}
		if err := form.Close(); err != nil {
			t.Fatal(err)
		}
		return &body, form.FormDataContentType()
	}
	body, kind := upload("invalid")
	if w := request("POST", view.EventsURL+"/new", staff, body, kind); w.Code != 403 {
		t.Fatal("event accepted without CSRF")
	}
	body, kind = upload(staff.CSRF)
	w := request("POST", view.EventsURL+"/new", staff, body, kind)
	if w.Code != 303 {
		t.Fatalf("upload: %d %s", w.Code, w.Body.String())
	}
	location := w.Header().Get("Location")
	w = request("GET", location, nil, nil, "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Plakat zum Sommergrillen") || !strings.Contains(w.Body.String(), "Innenhof") || strings.Contains(w.Body.String(), "Veranstaltung bearbeiten") {
		t.Fatalf("public detail: %d %s", w.Code, w.Body.String())
	}
	if w = request("GET", location+"/cover", nil, nil, ""); w.Code != 200 || w.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatal("public cover unavailable")
	}
	if w = request("GET", location+"/edit", member, nil, ""); w.Code != 403 {
		t.Fatal("member could edit event")
	}
	if w = request("GET", view.EventsURL, nil, nil, ""); !strings.Contains(w.Body.String(), location) {
		t.Fatal("event absent from list")
	}
	if w = request("GET", "/discussions", nil, nil, ""); strings.Contains(w.Body.String(), "Sommergrillen") {
		t.Fatal("event leaked into forum")
	}
	if w = request("GET", view.NewsURL+"/new", member, nil, ""); w.Code != 403 {
		t.Fatal("member could publish news")
	}
	if w = request("GET", view.NewsURL+"/new", staff, nil, ""); w.Code != 200 || strings.Contains(w.Body.String(), `name="type"`) {
		t.Fatal("dedicated news editor unavailable")
	}
}
