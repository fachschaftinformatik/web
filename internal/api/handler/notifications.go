package handler

import (
	"github.com/fachschaftinformatik/web/internal/id"
	"net/http"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/go-chi/chi/v5"
)

// @Summary List notifications
// @Tags Auth
// @ID getAuthNotifications
// @Success 200 {array} dto.NotificationResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/notifications [get]
func (s *Server) GetAuthNotifications(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)

	rows, err := s.DB.ListNotifications(r.Context(), user.ID)
	if err != nil {
		s.Log.Error("Failed to list notifications", "err", err)
		s.JsonError(w, "database_error", "Could not fetch notifications", http.StatusInternalServerError)
		return
	}

	resp := make([]dto.NotificationResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, dto.NotificationResponse{
			ID:        row.ID,
			Title:     row.Title,
			Message:   row.Message,
			Type:      row.Type,
			Link:      row.Link,
			Read:      row.Read == 1,
			CreatedAt: row.CreatedAt,
		})
	}

	s.RespondJSON(w, http.StatusOK, resp)
}

// @Summary Mark notification as read
// @Tags Auth
// @ID putAuthNotificationsIdRead
// @Param id path string true "Notification ID"
// @Success 200 {object} dto.NotificationResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /auth/notifications/{id}/read [put]
func (s *Server) PutAuthNotificationsIdRead(w http.ResponseWriter, r *http.Request) {
	notificationID := chi.URLParam(r, "id")
	user, _ := s.User(r)

	row, err := s.DB.MarkNotificationAsRead(r.Context(), database.MarkNotificationAsReadParams{
		ID:     notificationID,
		Userid: user.ID,
	})
	if err != nil {
		s.Log.Error("Failed to mark notification as read", "err", err)
		s.JsonError(w, "not_found", "Notification not found", http.StatusNotFound)
		return
	}

	s.RespondJSON(w, http.StatusOK, dto.NotificationResponse{
		ID:        row.ID,
		Title:     row.Title,
		Message:   row.Message,
		Type:      row.Type,
		Link:      row.Link,
		Read:      row.Read == 1,
		CreatedAt: row.CreatedAt,
	})
}

// @Summary Mark all notifications as read
// @Tags Auth
// @ID putAuthNotificationsReadAll
// @Success 204
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/notifications/read-all [put]
func (s *Server) PutAuthNotificationsReadAll(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)

	if err := s.DB.MarkAllNotificationsAsRead(r.Context(), user.ID); err != nil {
		s.Log.Error("Failed to mark all notifications as read", "err", err)
		s.JsonError(w, "database_error", "Failed to update notifications", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) createNotification(r *http.Request, userid, title, message, nType, link string) error {
	_, err := s.DB.CreateNotification(r.Context(), database.CreateNotificationParams{
		ID:      id.New(),
		Userid:  userid,
		Title:   title,
		Message: message,
		Type:    nType,
		Link:    link,
	})
	return err
}

func (s *Server) broadcastNotification(r *http.Request, title, message, nType, link string) {
	users, err := s.DB.ListUsersForNotification(r.Context())
	if err != nil {
		s.Log.Error("Failed to list users for broadcast", "err", err)
		return
	}

	for _, uID := range users {
		if err := s.createNotification(r, uID, title, message, nType, link); err != nil {
			s.Log.Error("Failed to create notification", "userID", uID, "err", err)
		}
	}
}
