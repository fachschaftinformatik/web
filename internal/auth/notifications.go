package auth

import (
	"net/http"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/sid"
	"github.com/go-chi/chi/v5"
)

type NotificationResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	Link      string `json:"link"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"created_at"`
}

// GetAuthNotifications lists notifications for the current user
// @Summary List notifications
// @Tags Auth
// @ID getAuthNotifications
// @Success 200 {array} NotificationResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/notifications [get]
func (s *Server) GetAuthNotifications(w http.ResponseWriter, r *http.Request) {
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	rows, err := s.DB.ListNotifications(r.Context(), user.ID)
	if err != nil {
		s.Log.Printf("Failed to list notifications: %v", err)
		s.jsonError(w, "database_error", "Could not fetch notifications", http.StatusInternalServerError)
		return
	}

	resp := make([]NotificationResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, NotificationResponse{
			ID:        row.ID,
			Title:     row.Title,
			Message:   row.Message,
			Type:      row.Type,
			Link:      row.Link,
			Read:      row.Read == 1,
			CreatedAt: row.CreatedAt,
		})
	}

	s.respondJSON(w, http.StatusOK, resp)
}

// PutAuthNotificationsIdRead marks a notification as read
// @Summary Mark notification as read
// @Tags Auth
// @ID putAuthNotificationsIdRead
// @Param id path string true "Notification ID"
// @Success 200 {object} NotificationResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /auth/notifications/{id}/read [put]
func (s *Server) PutAuthNotificationsIdRead(w http.ResponseWriter, r *http.Request) {
	notificationID := chi.URLParam(r, "id")
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	row, err := s.DB.MarkNotificationAsRead(r.Context(), database.MarkNotificationAsReadParams{
		ID:     notificationID,
		Userid: user.ID,
	})
	if err != nil {
		s.Log.Printf("Failed to mark notification as read: %v", err)
		s.jsonError(w, "not_found", "Notification not found", http.StatusNotFound)
		return
	}

	s.respondJSON(w, http.StatusOK, NotificationResponse{
		ID:        row.ID,
		Title:     row.Title,
		Message:   row.Message,
		Type:      row.Type,
		Link:      row.Link,
		Read:      row.Read == 1,
		CreatedAt: row.CreatedAt,
	})
}

// PutAuthNotificationsReadAll marks all notifications as read
// @Summary Mark all notifications as read
// @Tags Auth
// @ID putAuthNotificationsReadAll
// @Success 204
// @Failure 401 {object} ErrorResponse
// @Router /auth/notifications/read-all [put]
func (s *Server) PutAuthNotificationsReadAll(w http.ResponseWriter, r *http.Request) {
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if err := s.DB.MarkAllNotificationsAsRead(r.Context(), user.ID); err != nil {
		s.Log.Printf("Failed to mark all notifications as read: %v", err)
		s.jsonError(w, "database_error", "Failed to update notifications", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) createNotification(r *http.Request, userid, title, message, nType, link string) error {
	_, err := s.DB.CreateNotification(r.Context(), database.CreateNotificationParams{
		ID:      sid.New(),
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
		s.Log.Printf("Failed to list users for broadcast: %v", err)
		return
	}

	for _, uID := range users {
		if err := s.createNotification(r, uID, title, message, nType, link); err != nil {
			s.Log.Printf("Failed to create notification for user %s: %v", uID, err)
		}
	}
}
