package api

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/api/handler"
	"github.com/fachschaftinformatik/web/internal/config"
)

func TestRequestLoggingRecoversPanicsAndOmitsCredentials(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	// A missing database deliberately makes the programs handler panic.
	router := NewRouter(&handler.Server{Config: &config.Config{}, Log: logger}, logger)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/programs", nil)
	req.Header.Set("Authorization", "Bearer private-test-token")
	req.Header.Set("Cookie", "session=private-test-cookie")
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("panic response status = %d, want 500", response.Code)
	}
	output := logs.String()
	for _, secret := range []string{"private-test-token", "private-test-cookie"} {
		if strings.Contains(output, secret) {
			t.Fatal("request log leaked credentials")
		}
	}
	for _, expected := range []string{"application/json", "/api/v1/programs", "500"} {
		if !strings.Contains(output, expected) {
			t.Errorf("request log missing %q: %s", expected, output)
		}
	}
	// Recovery must leave the router usable for subsequent requests.
	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("subsequent response status = %d, want 404", response.Code)
	}
}
