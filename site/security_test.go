package site

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/urls"
)

func TestSecurityHeadersCoverRouterResponses(t *testing.T) {
	handler, _ := testApplication(t)
	for _, path := range []string{"/healthz", "/login", "/login/", "/admin", "/missing", "/assets/site.css", "/assets/missing.js", urls.ResetPasswordPath + "?token=invalid"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
			for name, want := range map[string]string{
				"X-Content-Type-Options": "nosniff",
				"X-Frame-Options":        "SAMEORIGIN",
			} {
				if got := response.Header().Get(name); got != want {
					t.Errorf("%s = %q, want %q", name, got, want)
				}
			}
			if policy := response.Header().Get("Content-Security-Policy"); !strings.Contains(policy, "frame-ancestors 'self'") || !strings.Contains(policy, "form-action 'self'") {
				t.Errorf("missing framing or form protection: %q", policy)
			}
			want := "strict-origin-when-cross-origin"
			if strings.HasPrefix(path, urls.ResetPasswordPath+"?") {
				want = "no-referrer"
			}
			if got := response.Header().Get("Referrer-Policy"); got != want {
				t.Errorf("Referrer-Policy = %q, want %q", got, want)
			}
		})
	}
}
