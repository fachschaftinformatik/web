package auth

import (
	"net/http"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/go-chi/chi/v5"
)

// GetUsersActivities lists activities for a specific user
// @Summary List user activities
// @Tags Activities
// @Param id path string true "User ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} database.Activity
// @Router /users/{id}/activities [get]
func (s *Server) GetUsersActivities(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// Check for privacy
	targetUser, err := s.DB.GetUser(r.Context(), userID)
	if err != nil {
		s.jsonError(w, "not_found", "User not found", http.StatusNotFound)
		return
	}

	_, authUser, authErr := s.authenticate(w, r)
	isOwnerOrAdmin := authErr == nil && (authUser.ID == userID || authUser.Role == "admin")

	if targetUser.Private == 1 && !isOwnerOrAdmin {
		w.Header().Set("X-Total-Count", "0")
		s.respondJSON(w, http.StatusOK, []database.Activity{})
		return
	}

	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)

	activities, err := s.DB.ListUserActivities(r.Context(), database.ListUserActivitiesParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		s.Log.Printf("Failed to list user activities: %v", err)
		s.jsonError(w, "database_error", "Could not fetch activities", http.StatusInternalServerError)
		return
	}

	count, err := s.DB.CountUserActivities(r.Context(), userID)
	if err != nil {
		s.Log.Printf("Failed to count user activities: %v", err)
		// Don't fail the request, just log and continue without count
		count = 0
	}

	w.Header().Set("X-Total-Count", strconv.FormatInt(count, 10))
	s.respondJSON(w, http.StatusOK, activities)
}

// GetActivities lists all activities
// @Summary List all activities
// @Tags Activities
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} database.Activity
// @Router /activities [get]
func (s *Server) GetActivities(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)

	activities, err := s.DB.ListAllActivities(r.Context(), database.ListAllActivitiesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		s.Log.Printf("Failed to list all activities: %v", err)
		s.jsonError(w, "database_error", "Could not fetch activities", http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, http.StatusOK, activities)
}
