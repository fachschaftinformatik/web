package handler

import (
	"net/http"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/go-chi/chi/v5"
)

// @Summary List user activities
// @Tags Activities
// @Param id path string true "User ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} dto.ActivityResponse
// @Router /users/{id}/activities [get]
func (s *Server) GetUsersActivities(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// Check for privacy
	targetUser, err := s.DB.GetUser(r.Context(), userID)
	if err != nil {
		s.JsonError(w, "not_found", "User not found", http.StatusNotFound)
		return
	}

	authUser, hasAuth := s.User(r)
	isOwnerOrAdmin := hasAuth && (authUser.ID == userID || authUser.Role == "admin")

	if targetUser.Private == 1 && !isOwnerOrAdmin {
		w.Header().Set("X-Total-Count", "0")
		s.RespondJSON(w, http.StatusOK, []dto.ActivityResponse{})
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
		s.Log.Error("Failed to list user activities", "err", err)
		s.JsonError(w, "database_error", "Could not fetch activities", http.StatusInternalServerError)
		return
	}

	count, err := s.DB.CountUserActivities(r.Context(), userID)
	if err != nil {
		s.Log.Error("Failed to count user activities", "err", err)
		count = 0
	}

	w.Header().Set("X-Total-Count", strconv.FormatInt(count, 10))
	resp := make([]dto.ActivityResponse, 0, len(activities))
	for _, a := range activities {
		resp = append(resp, dto.ActivityResponse{
			ID:         a.ID,
			UserID:     a.UserID,
			UserName:   a.UserName,
			Type:       a.Type,
			TargetID:   a.TargetID,
			TargetName: a.TargetName,
			CreatedAt:  a.CreatedAt,
		})
	}
	s.RespondJSON(w, http.StatusOK, resp)
}

// @Summary List all activities
// @Tags Activities
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} dto.ActivityResponse
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
		s.Log.Error("Failed to list all activities", "err", err)
		s.JsonError(w, "database_error", "Could not fetch activities", http.StatusInternalServerError)
		return
	}

	resp := make([]dto.ActivityResponse, 0, len(activities))
	for _, a := range activities {
		resp = append(resp, dto.ActivityResponse{
			ID:         a.ID,
			UserID:     a.UserID,
			UserName:   a.UserName,
			Type:       a.Type,
			TargetID:   a.TargetID,
			TargetName: a.TargetName,
			CreatedAt:  a.CreatedAt,
		})
	}
	s.RespondJSON(w, http.StatusOK, resp)
}
