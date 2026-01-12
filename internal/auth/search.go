package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	_, _, authErr := s.authenticate(nil, r)
	if authErr != nil {
		s.respondJSON(w, http.StatusOK, []SearchResult{})
		return
	}

	results := []SearchResult{}

	// Search Modules
	modules, err := s.DB.SearchModules(r.Context(), &query)
	if err == nil {
		for _, m := range modules {
			results = append(results, SearchResult{
				Type:     "module",
				ID:       strconv.FormatInt(m.ID, 10),
				Title:    m.Name,
				Subtitle: "Modul",
				URL:      fmt.Sprintf("/rekos/klausuren/modul?modulId=%d&mod=%s", m.ID, url.QueryEscape(m.Name)),
			})
		}
	}

	// Search Exams
	exams, err := s.DB.SearchExams(r.Context(), &query)
	if err == nil {
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
				URL:      fmt.Sprintf("/rekos/klausuren/modul?modulId=%d&mod=%s&examId=%s", modID, url.QueryEscape(e.ModuleName), url.QueryEscape(e.ID)),
			})
		}
	}

	s.respondJSON(w, http.StatusOK, results)
}
