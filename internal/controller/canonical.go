package controller

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/fachschaftinformatik/web/internal/publicid"
	"github.com/fachschaftinformatik/web/internal/urls"
)

func validRouteIDs(r *http.Request) bool {
	for _, name := range []string{"id", "post", "comment", "event", "album", "image", "module", "submission"} {
		if value := r.PathValue(name); value != "" && !publicid.Valid(value) {
			return false
		}
	}
	return true
}

// parsePage rejects ambiguous and malformed pagination instead of silently clamping.
func parsePage(r *http.Request) (int, bool) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return 0, false
	}
	raw, exists := values["page"]
	if !exists {
		return 1, true
	}
	if len(raw) != 1 || raw[0] == "" {
		return 0, false
	}
	for _, digit := range raw[0] {
		if digit < '0' || digit > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(raw[0])
	return n, err == nil && n > 0
}

func (c *Handler) validateRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validRouteIDs(r) {
			c.notFound(w, r)
			return
		}
		if _, ok := parsePage(r); !ok {
			c.pageError(w, "Gib eine gültige Seitenzahl an.", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// validPageRange must run after obtaining the total, before rendering a model's
// potentially clamped result. Empty collections still have a valid first page.
func (c *Handler) validPageRange(w http.ResponseWriter, r *http.Request, total, size int) bool {
	n, ok := parsePage(r)
	if !ok {
		c.pageError(w, "Gib eine gültige Seitenzahl an.", http.StatusBadRequest)
		return false
	}
	if size <= 0 || n > max(1, (total+size-1)/size) {
		c.notFound(w, r)
		return false
	}
	return true
}

func (c *Handler) canonicalize(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			mux.ServeHTTP(w, r)
			return
		}
		candidate := r.Clone(r.Context())
		u := *r.URL
		candidate.URL = &u
		if u.Path != urls.HomePath {
			u.Path = strings.TrimRight(u.Path, "/")
		}
		u.RawPath = ""
		_, pattern := mux.Handler(candidate)
		// The fallback and asset subtree are not declared HTML/application routes.
		if pattern == "/" {
			c.notFound(w, r)
			return
		}
		if pattern == "" || strings.HasPrefix(u.Path, urls.AssetsPath) {
			mux.ServeHTTP(w, r)
			return
		}
		// ServeMux.Handler does not populate captures. Validate those in the matched
		// pattern before redirecting so legacy/invalid IDs remain a plain 404.
		parts := strings.Split(strings.TrimPrefix(pattern, "GET "), "/")
		path := strings.Split(u.Path, "/")
		for i, part := range parts {
			if strings.HasPrefix(part, "{") && part != "{$}" && part != "{size}" && i < len(path) {
				value, err := url.PathUnescape(path[i])
				if err != nil || !publicid.Valid(value) {
					c.notFound(w, r)
					return
				}
			}
		}
		n, ok := parsePage(r)
		if !ok {
			c.pageError(w, "Gib eine gültige Seitenzahl an.", http.StatusBadRequest)
			return
		}
		query := u.Query()
		if _, exists := query["page"]; exists {
			if n == 1 {
				query.Del("page")
			} else {
				query.Set("page", strconv.Itoa(n))
			}
			if query.Encode() != r.URL.Query().Encode() {
				u.RawQuery = query.Encode()
			}
		}
		if u.Path != r.URL.Path || u.RawQuery != r.URL.RawQuery {
			http.Redirect(w, r, u.RequestURI(), http.StatusPermanentRedirect)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

// canonicalURL is deliberately an allowlist of indexable public HTML routes.
// Root-relative URLs avoid deriving an origin from an untrusted Host header.
func canonicalURL(r *http.Request) string {
	pattern := strings.TrimPrefix(r.Pattern, "GET ")
	var keys []string
	switch pattern {
	case urls.HomePattern, urls.AboutPath, urls.TeamPath, urls.LegalPath,
		urls.PrivacyPath, urls.AccessibilityPath, urls.ContactPath, urls.ElectionsPath,
		urls.RegulationsPath, urls.CommunityPath, urls.EventPattern,
		urls.MediaItemPattern:
	case urls.MeetingsPath:
		keys = []string{"year"}
	case urls.PostsPath, urls.NewsPath:
		keys = []string{"page", "search"}
	case urls.PostPattern:
		keys = []string{"page"}
	case urls.EventsPath, urls.MediaPath, urls.AlbumPattern, urls.SitemapPath:
		keys = []string{"page"}
	case urls.ExamsPath:
		keys = []string{"page", "search", "program", "degree", "version"}
	case urls.ExamModulePattern:
		keys = []string{"page", "search", "program", "degree", "version"}
	default:
		return ""
	}
	q := url.Values{}
	for _, key := range keys {
		for _, value := range r.URL.Query()[key] {
			if value != "" && !(key == "page" && value == "1") {
				q.Add(key, value)
			}
		}
	}
	u := url.URL{Path: r.URL.Path, RawQuery: q.Encode()}
	return u.String()
}
