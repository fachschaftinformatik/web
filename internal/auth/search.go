package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/database"
)

type SearchResult struct {
	Type     string `json:"type"` // "exam" or "module"
	ID       string `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	URL      string `json:"url"`
}

// GetSearch searches for exams and modules
// @Summary Global search
// @Tags Search
// @ID getSearch
// @Param q query string true "Search query"
// @Success 200 {array} SearchResult
// @Router /search [get]
func (s *Server) GetSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		s.respondJSON(w, http.StatusOK, []SearchResult{})
		return
	}

	// Only for authenticated users
	user, _, authErr := s.authenticate(nil, r)
	if authErr != nil {
		fmt.Printf("[Search] Auth failed: %v\n", authErr)
		s.respondJSON(w, http.StatusOK, []SearchResult{})
		return
	}
	fmt.Printf("[Search] Query: %s, User: %s\n", query, user.ID)

	results := []SearchResult{}

	// Search Modules
	modules, err := s.DB.SearchModules(r.Context(), &query)
	if err == nil {
		fmt.Printf("[Search] Found %d modules\n", len(modules))
		for _, m := range modules {
			results = append(results, SearchResult{
				Type:     "module",
				ID:       strconv.FormatInt(m.ID, 10),
				Title:    m.Name,
				Subtitle: "Modul",
				URL:      fmt.Sprintf("/exams/%d?mod=%s", m.ID, url.QueryEscape(m.Name)),
			})
		}
	} else {
		fmt.Printf("[Search] Module search error: %v\n", err)
	}

	// Search Exams
	exams, err := s.DB.SearchExams(r.Context(), &query)
	if err == nil {
		fmt.Printf("[Search] Found %d exams\n", len(exams))
		for _, e := range exams {
			var modID int64
			if e.Moduleid != nil {
				modID = *e.Moduleid
			}
			results = append(results, SearchResult{
				Type:     "exam",
				ID:       e.ID,
				Title:    e.ModuleName,
				Subtitle: "Klausur",
				URL:      fmt.Sprintf("/exams/%d?mod=%s&examId=%s", modID, url.QueryEscape(e.ModuleName), url.QueryEscape(e.ID)),
			})
		}
	} else {
		fmt.Printf("[Search] Exam search error: %v\n", err)
	}

	// Search Users
	users, err := s.DB.SearchUsers(r.Context(), query)
	if err == nil {
		fmt.Printf("[Search] Found %d users\n", len(users))
		for _, u := range users {
			results = append(results, SearchResult{
				Type:     "user",
				ID:       u.ID,
				Title:    u.Name,
				Subtitle: "Benutzer",
				URL:      fmt.Sprintf("/user/%s", u.ID),
			})
		}
	} else {
		fmt.Printf("[Search] User search error: %v\n", err)
	}

	// Search Forum Posts
	posts, err := s.DB.SearchForumPosts(r.Context(), database.SearchForumPostsParams{
		Query:         query,
		CurrentUserID: &user.ID,
	})
	if err == nil {
		fmt.Printf("[Search] Found %d posts\n", len(posts))
		for _, p := range posts {
			results = append(results, SearchResult{
				Type:     "post",
				ID:       p.ID,
				Title:    p.Title,
				Subtitle: "Forum: " + p.AuthorName,
				URL:      fmt.Sprintf("/forum/%s", p.ID),
			})
		}
	} else {
		fmt.Printf("[Search] Post search error: %v\n", err)
	}

	s.respondJSON(w, http.StatusOK, results)
}
