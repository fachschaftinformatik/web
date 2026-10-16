package web

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/images"
)

func TestGalleryJourneys(t *testing.T) {
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
		{"/events/140125680309370881", 200, "Nächste Seite"},
		{"/events/140125680309370881?page=2", 200, "Gemeinsam spielen"},
		{"/events/140125680309370881/140126279717355521", 200, "/community/events/140125680309370881?page=2#image-140126279717355521"},
		{"/events/unknown", 404, "404"},
		{"/events/140125680309370881/unknown", 404, "404"},
		{"/events/140117348945231873/140126279717355521", 404, "404"},
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
		{"/events/140125680309370881?page=2", "Zusammen auf der LAN Party"},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/community"+tc.path, nil))
		if response.Code != 200 || strings.Contains(response.Body.String(), tc.absent) {
			t.Fatalf("unexpected content for %s: %s", tc.path, response.Body.String())
		}
	}
}

func TestGalleryEmptyAndSingleImageRendering(t *testing.T) {
	renderer, err := newRenderer()
	if err != nil {
		t.Fatal(err)
	}
	for _, count := range []int{0, 1} {
		var photos []images.Image
		if count == 1 {
			photos = images.Albums()[0].Images[:1]
		}
		var body bytes.Buffer
		err := renderer.Render(&body, "gallery", false, Page{Title: "Testalbum", Active: "events", Gallery: GalleryPage{Detail: true, Album: images.Album{ID: "test", Title: "Testalbum"}, Images: photos, Total: count, Page: 1, Pages: 1}})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(body.String(), "Albumseiten") {
			t.Fatal("pagination shown for a single page")
		}
		if count == 0 && !strings.Contains(body.String(), "noch keine Bilder") {
			t.Fatal("missing empty state")
		}
	}
}

func TestPublicNavigation(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ path, text string }{
		{"/", "Deine Fachschaft Informatik"},
		{"/community", "Lerne andere Studierende kennen"},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", tc.path, nil))
		if response.Code != 200 || !strings.Contains(response.Body.String(), tc.text) {
			t.Fatalf("unexpected page %s: %d", tc.path, response.Code)
		}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/events/140125680309370881?page=2", nil))
	if response.Code != 303 || response.Header().Get("Location") != "/community/events/140125680309370881?page=2" {
		t.Fatal("legacy album URL lost its path or page")
	}
}
