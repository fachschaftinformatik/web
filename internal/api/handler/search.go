package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/database"
)

// @Summary Global search
// @Tags Search
// @ID getSearch
// @Param q query string true "Search query"
// @Success 200 {array} dto.SearchResult
// @Router /search [get]
func (s *Server) GetSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		s.RespondJSON(w, http.StatusOK, []dto.SearchResult{})
		return
	}

	user, hasAuth := s.User(r)
	if !hasAuth {
		s.RespondJSON(w, http.StatusOK, []dto.SearchResult{})
		return
	}

	results := []dto.SearchResult{}

	// Search Modules
	modules, err := s.DB.SearchModules(r.Context(), &query)
	if err == nil {
		for _, m := range modules {
			results = append(results, dto.SearchResult{
				Type:     "module",
				ID:       m.ID,
				Title:    m.Name,
				Subtitle: "Modul",
				URL:      fmt.Sprintf("/exams/%s?mod=%s", m.ID, url.QueryEscape(m.Name)),
			})
		}
	}

	// Search Exams
	exams, err := s.DB.SearchExams(r.Context(), &query)
	if err == nil {
		for _, e := range exams {
			var modID string
			if e.Moduleid != nil {
				modID = *e.Moduleid
			}
			results = append(results, dto.SearchResult{
				Type:     "exam",
				ID:       e.ID,
				Title:    e.ModuleName,
				Subtitle: "Klausur",
				URL:      fmt.Sprintf("/exams/%s?mod=%s&examId=%s", modID, url.QueryEscape(e.ModuleName), url.QueryEscape(e.ID)),
			})
		}
	}

	// Search Users
	users, err := s.DB.SearchUsers(r.Context(), query)
	if err == nil {
		for _, u := range users {
			results = append(results, dto.SearchResult{
				Type:     "user",
				ID:       u.ID,
				Title:    s.ToPublicUserResponse(u).Name,
				Subtitle: "Benutzer",
				URL:      fmt.Sprintf("/user/%s", u.ID),
			})
		}
	}

	// Search Forum Posts
	posts, err := s.DB.SearchForumPosts(r.Context(), database.SearchForumPostsParams{
		Query:         query,
		CurrentUserID: &user.ID,
	})
	if err == nil {
		for _, p := range posts {
			results = append(results, dto.SearchResult{
				Type:     "post",
				ID:       p.ID,
				Title:    p.Title,
				Subtitle: "Forum: " + p.AuthorName,
				URL:      fmt.Sprintf("/forum/%s", p.ID),
			})
		}
	}

	s.RespondJSON(w, http.StatusOK, results)
}
