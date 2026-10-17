package web

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/content/media"
	"github.com/fachschaftinformatik/web/internal/data/blob"
)

func TestImagesNavigation(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path     string
		status   int
		contains string
	}{
		{"/events", 200, "LAN Party 2025"},
		{"/events/140125680309370881", 200, "image-140126291885031425"},
		{"/events/140125680309370881?page=2", 303, "/events/140125680309370881"},
		{"/events/140125680309370881/140126279717355521", 200, "/community/events/140125680309370881?page=1#image-140126279717355521"},
		{"/events/unknown", 404, "Seite nicht gefunden"},
		{"/events/140125680309370881/unknown", 404, "Seite nicht gefunden"},
		{"/events/140117348945231873/140126279717355521", 404, "Seite nicht gefunden"},
		{"/events/140125680309370881?page=3", 303, "/events/140125680309370881"},
		{"/events/140125680309370881?page=-1", 303, "/events/140125680309370881"},
		{"/events/140125680309370881?page=bad", 303, "/events/140125680309370881"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", "/community"+tc.path, nil))
			if response.Code != tc.status || !strings.Contains(response.Body.String(), tc.contains) {
				t.Fatalf("status %d, body %s", response.Code, response.Body.String())
			}
		})
	}
	for _, tc := range []struct{ path, absent string }{
		{"/events/140125680309370881/140126343856652289", "Vorheriges Bild"},
		{"/events/140125680309370881/140126279717355521", "Nächstes Bild"},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/community"+tc.path, nil))
		if response.Code != 200 || strings.Contains(response.Body.String(), tc.absent) {
			t.Fatalf("unexpected content for %s: %s", tc.path, response.Body.String())
		}
	}
}

func TestLegacyAlbumRedirect(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/events/140125680309370881?page=2", nil))
	if response.Code != 303 || response.Header().Get("Location") != "/community/events/140125680309370881?page=2" {
		t.Fatal("legacy album URL lost its path or page")
	}
}

type galleryBlob struct {
	blob.Provider
	stream bool
	keys   []string
}

type seekablePhoto struct{ *strings.Reader }

func (seekablePhoto) Close() error { return nil }

func (s *galleryBlob) GetObject(ctx context.Context, key string) (*blob.Object, error) {
	s.keys = append(s.keys, key)
	if key != "media/101/p/400.jpg" && key != "media/101/source" && key != "events/10/cover/p/400.jpg" {
		return nil, errors.New("missing")
	}
	var reader io.ReadCloser = seekablePhoto{strings.NewReader("image bytes")}
	if s.stream {
		reader = io.NopCloser(strings.NewReader("image bytes"))
	}
	return &blob.Object{ReadCloser: reader, Info: blob.ObjectInfo{Size: 11, ContentType: "application/octet-stream"}}, nil
}

func liveGalleryHandler(t *testing.T, storage blob.Provider) *Handler {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec(`CREATE TABLE events (id INTEGER PRIMARY KEY,title TEXT,created_at TEXT,deleted_at TEXT,cover_path TEXT);
CREATE TABLE media (id INTEGER PRIMARY KEY,event_id INTEGER,title TEXT,description TEXT,mime_type TEXT,nbytes INTEGER,created_at TEXT,deleted_at TEXT);
INSERT INTO events (id,title,created_at,deleted_at) VALUES (10,'Live Album','2026',NULL),(20,'Other Album','2025',NULL),(30,'Deleted Album','2027','deleted');
INSERT INTO media VALUES (101,10,'Live Title','Real alt','image/jpeg',11,'2026',NULL),(102,10,'Hidden','Hidden','image/jpeg',11,'2026','deleted'),(103,30,'Hidden','Hidden','image/jpeg',11,'2026',NULL),(104,10,'Unsafe','Unsafe','image/svg+xml',11,'2026',NULL);
UPDATE events SET cover_path='events/' || id || '/cover/source' WHERE id IN (10,30);`)
	if err != nil {
		t.Fatal(err)
	}
	for id := 110; id < 180; id++ {
		if _, err := db.Exec(`INSERT INTO media VALUES (?,10,?,NULL,'image/jpeg',11,'2025',NULL)`, id, "Bild "+strconv.Itoa(id)); err != nil {
			t.Fatal(err)
		}
	}
	views, err := newRenderer()
	if err != nil {
		t.Fatal(err)
	}
	return &Handler{views: views, media: media.NewService(db, storage)}
}

func TestLiveGalleryPaginationAndViewer(t *testing.T) {
	c := liveGalleryHandler(t, nil)
	router := http.NewServeMux()
	router.HandleFunc("GET /community/events", c.Images)
	router.HandleFunc("GET /community/events/{album}", c.Images)
	router.HandleFunc("GET /community/events/{album}/{image}", c.Images)
	for _, tc := range []struct {
		path string
		code int
		want string
	}{
		{"/community/events", 200, "Live Album"},
		{"/community/events/10", 200, "/media/101/400"},
		{"/community/events/10?page=3", 200, "Bild 110"},
		{"/community/events/10?page=4", 303, "/community/events/10"},
		{"/community/events/10/115", 200, "/community/events/10?page=3#image-115"},
		{"/community/events/10/101", 200, `alt="Real alt"`},
		{"/community/events/20/101", 404, "Seite nicht gefunden"},
		{"/community/events/10/102", 404, "Seite nicht gefunden"},
		{"/community/events/30", 404, "Seite nicht gefunden"},
	} {
		r := httptest.NewRecorder()
		router.ServeHTTP(r, httptest.NewRequest("GET", tc.path, nil))
		if r.Code != tc.code || !strings.Contains(r.Body.String(), tc.want) {
			t.Errorf("%s: status %d, body %s", tc.path, r.Code, r.Body.String())
		}
	}
}

func TestMediaHTTPRangesAndPathProtections(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(strconv.FormatBool(stream), func(t *testing.T) {
			storage := &galleryBlob{stream: stream}
			c := liveGalleryHandler(t, storage)
			router := http.NewServeMux()
			router.HandleFunc("GET /media/{id}/{size}", c.MediaFile)
			for _, tc := range []struct {
				method, path, byteRange string
				code                    int
				body                    string
			}{
				{"GET", "/media/101/400", "", 200, "image bytes"},
				{"GET", "/media/101/original", "bytes=2-5", 206, "age "},
				{"HEAD", "/media/101/original", "", 200, ""},
				{"GET", "/media/101/400", "bytes=100-", 416, ""},
				{"GET", "/media/102/400", "", 404, ""},
				{"GET", "/media/103/original", "", 404, ""},
				{"GET", "/media/104/original", "", 404, ""},
				{"GET", "/media/999/original", "", 404, ""},
				{"GET", "/media/avatars/source", "", 404, ""},
				{"GET", "/media/101/%2e%2e%2fsource", "", 404, ""},
				{"GET", "/media/%2e%2e%2favatars%2f1/source", "", 404, ""},
				{"GET", "/media/101/800", "", 503, ""},
			} {
				r := httptest.NewRecorder()
				req := httptest.NewRequest(tc.method, tc.path, nil)
				req.Header.Set("Range", tc.byteRange)
				router.ServeHTTP(r, req)
				if r.Code != tc.code || ((tc.code == 200 || tc.code == 206) && r.Body.String() != tc.body) {
					t.Errorf("%s %s range %s: status %d, body %q", tc.method, tc.path, tc.byteRange, r.Code, r.Body.String())
				}
				if tc.code == 200 || tc.code == 206 {
					if r.Header().Get("Content-Type") != "image/jpeg" || r.Header().Get("Cache-Control") != "public, no-cache" || r.Header().Get("Content-Length") == "" || r.Header().Get("Accept-Ranges") != "bytes" {
						t.Errorf("missing image headers: %v", r.Header())
					}
					length := "11"
					if tc.code == 206 {
						length = "4"
						if got := r.Header().Get("Content-Range"); got != "bytes 2-5/11" {
							t.Errorf("Content-Range: %s", got)
						}
					}
					if got := r.Header().Get("Content-Length"); got != length {
						t.Errorf("Content-Length: %s, want %s", got, length)
					}
				}
			}
			for _, key := range storage.keys {
				if !strings.HasPrefix(key, "media/101/") {
					t.Errorf("unauthorized storage access: %s", key)
				}
			}
		})
	}
}

func TestEventCoverHTTPRangesAndVisibility(t *testing.T) {
	for _, stream := range []bool{false, true} {
		storage := &galleryBlob{stream: stream}
		c := liveGalleryHandler(t, storage)
		router := http.NewServeMux()
		router.HandleFunc("GET /media/events/{id}/cover/{size}", c.EventCover)
		for _, tc := range []struct {
			method, path string
			code         int
		}{{"GET", "/media/events/10/cover/400", 206}, {"HEAD", "/media/events/10/cover/400", 206}, {"GET", "/media/events/20/cover/400", 404}, {"GET", "/media/events/30/cover/400", 404}, {"GET", "/media/events/999/cover/400", 404}, {"GET", "/media/events/10/cover/original", 404}} {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(tc.method, tc.path, nil)
			request.Header.Set("Range", "bytes=2-5")
			router.ServeHTTP(response, request)
			if response.Code != tc.code {
				t.Fatalf("%s %s: %d", tc.method, tc.path, response.Code)
			}
			if tc.code == 206 && (response.Header().Get("Content-Type") != "image/jpeg" || response.Header().Get("Content-Range") != "bytes 2-5/11" || (tc.method == "GET" && response.Body.String() != "age ")) {
				t.Fatalf("cover bytes or headers: %v %q", response.Header(), response.Body.String())
			}
		}
		for _, key := range storage.keys {
			if key != "events/10/cover/p/400.jpg" {
				t.Fatalf("unexpected storage access: %s", key)
			}
		}
	}
}
