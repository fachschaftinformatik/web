package controller

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRegistrationPreview(t *testing.T) {
	handler, err := testHandler(t)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, email, password, message string
		status                         int
	}{
		{"", "", "", "Bitte gib einen Namen an.", 422},
		{"A", "name@studmail.w-hs.de", "private-test-password", "Der Name muss zwischen 2 und 64 Zeichen lang sein.", 422},
		{"Name", "name@studmail.w-hs.de.evil.example", "private-test-password", "Verwende eine E-Mail-Adresse mit", 422},
		{"Name", "name@sub.fsv-wh.de", "private-test-password", "Verwende eine E-Mail-Adresse mit", 422},
		{"Name", "Name <name@fsv-wh.de>", "private-test-password", "Gib eine vollständige E-Mail-Adresse ein", 422},
		{"Name", "name@studmail.w-hs.de", "short", "Das Passwort muss mindestens 8 Zeichen lang sein.", 422},
		{"Name", "name@studmail.w-hs.de", strings.Repeat("x", 73), "Prüfe dein E-Mail-Postfach", 200},
		{"Name", "name@studmail.w-hs.de", "äöüßéèñ", "Das Passwort muss mindestens 8 Zeichen lang sein.", 422},
		{"Name", "name@studmail.w-hs.de", "äöüßéèñø", "Prüfe dein E-Mail-Postfach", 200},
		{"Name", "name@studmail.w-hs.de", " a b c d", "Prüfe dein E-Mail-Postfach", 200},
		{"  Name  ", "  Name@STUDMAIL.W-HS.DE  ", "private-test-password", "Prüfe dein E-Mail-Postfach", 200},
		{"Name", "name@fsv-wh.de", "private-test-password", "Verwende eine E-Mail-Adresse mit @studmail.w-hs.de.", 422},
	} {
		handler, _ := testHandler(t)
		form := url.Values{"name": {tc.name}, "email": {tc.email}, "password": {tc.password}, "program_id": {"1"}}
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.Register(response, req)
		body := response.Body.String()
		if response.Code != tc.status || !strings.Contains(body, tc.message) {
			t.Fatalf("unexpected registration response: %d %s", response.Code, body)
		}
		if response.Header().Get("Cache-Control") != "no-store" || len(response.Result().Cookies()) > 0 || strings.Contains(body, `value="private-test-password"`) {
			t.Fatal("registration cached credentials, reflected a password or created a session")
		}
		if tc.status == 422 && (!strings.Contains(body, "Fehler: Konto erstellen") || !strings.Contains(body, "govuk-error-summary")) {
			t.Fatal("missing accessible validation errors")
		}
	}
	for _, program := range []string{"", "unknown-program"} {
		handler, _ := testHandler(t)
		form := url.Values{"name": {"Testperson"}, "email": {"test@studmail.w-hs.de"}, "password": {"private-test-password"}, "program_id": {program}}
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.Register(response, req)
		if program == "" {
			if response.Code != http.StatusOK {
				t.Fatal("omitting the optional programme must pass validation")
			}
		} else if response.Code != 422 || !strings.Contains(response.Body.String(), `href="#program_id"`) || !strings.Contains(response.Body.String(), "Bitte wähle deinen Studiengang aus.") {
			t.Fatal("unknown program must produce a linked field error")
		}
	}
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(method, "/register", nil))
		if response.Code != 200 || response.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("registration page unavailable or cacheable")
		}
	}
	response := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("name="+strings.Repeat("x", 65<<10)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler.Register(response, req)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("oversized registration status = %d", response.Code)
	}
}
