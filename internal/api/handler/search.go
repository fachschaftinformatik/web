package handler

import (
	"fmt"
	"net/http"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/id"
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

	modules, err := s.DB.SearchModules(r.Context(), &query)
	if err == nil {
		for _, m := range modules {
			results = append(results, dto.SearchResult{
				Type:     "module",
				ID:       id.ID(m.ID).String(),
				Title:    m.Name,
				Subtitle: "Modul",
				URL:      fmt.Sprintf("/archive?module_id=%s", id.ID(m.ID).String()),
			})
		}
	}

	archive, err := s.DB.SearchArchive(r.Context(), &query)
	if err == nil {
		for _, e := range archive {
			results = append(results, dto.SearchResult{
				Type:     "archive",
				ID:       id.ID(e.ID).String(),
				Title:    e.ModuleName,
				Subtitle: "Archiv",
				URL:      fmt.Sprintf("/archive/%s", id.ID(e.ID).String()),
			})
		}
	}

	users, err := s.DB.SearchUsers(r.Context(), query)
	if err == nil {
		for _, u := range users {
			results = append(results, dto.SearchResult{
				Type:     "user",
				ID:       id.ID(u.ID).String(),
				Title:    s.ToPublicUserResponse(u).Name,
				Subtitle: "Benutzer",
				URL:      fmt.Sprintf("/u/%s", id.ID(u.ID).String()),
			})
		}
	}

	posts, err := s.DB.SearchDiscussionPosts(r.Context(), database.SearchDiscussionPostsParams{
		Query:         query,
		CurrentUserID: &user.ID,
	})
	if err == nil {
		for _, p := range posts {
			results = append(results, dto.SearchResult{
				Type:     "discussion",
				ID:       id.ID(p.ID).String(),
				Title:    p.Title,
				Subtitle: "Diskussion: " + p.UserName,
				URL:      fmt.Sprintf("/d/%s", id.ID(p.ID).String()),
			})
		}
	}

	s.RespondJSON(w, http.StatusOK, results)
}
