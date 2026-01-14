package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/sid"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxMediaUploadSize = 50 << 20

type EventResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CoverPath string    `json:"cover_path,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type MediaResponse struct {
	ID           string    `json:"id"`
	EventID      int64     `json:"event_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	UploadedAt   time.Time `json:"uploaded_at"`
	UploaderName string    `json:"uploader_name"`
	MimeType     string    `json:"mime_type"`
}

func (s *Server) GetEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.DB.ListEvents(r.Context())
	if err != nil {
		s.Log.Printf("Failed to list events: %v", err)
		s.jsonError(w, "database_error", "Could not fetch events", http.StatusInternalServerError)
		return
	}

	res := make([]EventResponse, 0, len(events))
	for _, e := range events {
		t, _ := time.Parse(time.RFC3339, e.CreatedAt)
		var cover string
		if e.CoverPath != nil {
			cover = *e.CoverPath
		}
		res = append(res, EventResponse{
			ID:        e.ID,
			Title:     e.Title,
			CoverPath: cover,
			CreatedAt: t,
		})
	}
	s.respondJSON(w, http.StatusOK, res)
}

func (s *Server) PostEvents(w http.ResponseWriter, r *http.Request) {
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if user.Role != "admin" && user.Role != "editor" {
		s.jsonError(w, "forbidden", "Insufficient permissions", http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		s.jsonError(w, "bad_request", "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		s.jsonError(w, "bad_request", "Title is required", http.StatusBadRequest)
		return
	}

	var coverPathPtr *string
	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		accessKey := uuid.NewString()
		objectKey := fmt.Sprintf("events/%s", accessKey)
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		if err := s.Store.Upload(r.Context(), objectKey, file, header.Size, contentType); err != nil {
			s.Log.Printf("S3 Upload failed: %v", err)
			s.jsonError(w, "server_error", "Failed to store cover image", http.StatusInternalServerError)
			return
		}
		coverPathPtr = &objectKey
	}

	params := database.CreateEventParams{
		Title:     title,
		CoverPath: coverPathPtr,
	}

	event, err := s.DB.CreateEvent(r.Context(), params)
	if err != nil {
		s.Log.Printf("Failed to create event: %v", err)
		s.jsonError(w, "database_error", "Could not create event", http.StatusInternalServerError)
		return
	}

	t, _ := time.Parse(time.RFC3339, event.CreatedAt)
	res := EventResponse{
		ID:        event.ID,
		Title:     event.Title,
		CreatedAt: t,
	}
	if event.CoverPath != nil {
		res.CoverPath = *event.CoverPath
	}

	s.respondJSON(w, http.StatusCreated, res)
}

func (s *Server) GetMedia(w http.ResponseWriter, r *http.Request) {
	eventIdStr := r.URL.Query().Get("event_id")
	eventId, err := strconv.ParseInt(eventIdStr, 10, 64)
	if err != nil {
		s.jsonError(w, "bad_request", "Invalid event_id", http.StatusBadRequest)
		return
	}

	media, err := s.DB.ListMediaByEvent(r.Context(), eventId)
	if err != nil {
		s.Log.Printf("Failed to list media: %v", err)
		s.jsonError(w, "database_error", "Could not fetch media", http.StatusInternalServerError)
		return
	}

	res := make([]MediaResponse, 0, len(media))
	for _, m := range media {
		t, _ := time.Parse(time.RFC3339, m.UploadedAt)
		var desc string
		if m.Description != nil {
			desc = *m.Description
		}
		var title string
		if m.Title != nil {
			title = *m.Title
		}

		res = append(res, MediaResponse{
			ID:           m.ID,
			EventID:      m.EventID,
			Title:        title,
			Description:  desc,
			UploadedAt:   t,
			UploaderName: m.UploaderName,
			MimeType:     m.MimeType,
		})
	}
	s.respondJSON(w, http.StatusOK, res)
}

func (s *Server) PostMedia(w http.ResponseWriter, r *http.Request) {
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if user.Role != "admin" && user.Role != "editor" {
		s.jsonError(w, "forbidden", "Insufficient permissions", http.StatusForbidden)
		return
	}

	if err := s.checkCSRF(r); err != nil {
		s.jsonError(w, "invalid_csrf", err.Error(), http.StatusForbidden)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxMediaUploadSize+1024*1024)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		if errors.As(err, new(*http.MaxBytesError)) {
			s.jsonError(w, "bad_request", "File too large", http.StatusBadRequest)
		} else {
			s.Log.Printf("Multipart parse error: %v", err)
			s.jsonError(w, "bad_request", "Upload failed", http.StatusBadRequest)
		}
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.jsonError(w, "bad_request", "File upload failed", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > maxMediaUploadSize {
		s.jsonError(w, "bad_request", "File too large", http.StatusBadRequest)
		return
	}

	eventIdStr := r.FormValue("event_id")
	eventId, err := strconv.ParseInt(eventIdStr, 10, 64)
	if err != nil {
		s.jsonError(w, "bad_request", "Invalid event_id", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	accessKey := uuid.NewString()
	objectKey := fmt.Sprintf("media/%s", accessKey)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	file.Seek(0, 0)
	if err := s.Store.Upload(r.Context(), objectKey, file, header.Size, contentType); err != nil {
		s.Log.Printf("S3 Upload failed: %v", err)
		s.jsonError(w, "server_error", "Failed to store file", http.StatusInternalServerError)
		return
	}

	var titlePtr, descPtr *string
	if title != "" {
		titlePtr = &title
	}
	if description != "" {
		descPtr = &description
	}

	params := database.CreateMediaParams{
		ID:          sid.New(),
		EventID:     eventId,
		Userid:      user.ID,
		Title:       titlePtr,
		Description: descPtr,
		Accesskey:   accessKey,
		MimeType:    contentType,
		Nbytes:      header.Size,
	}

	_, err = s.DB.CreateMedia(r.Context(), params)
	if err != nil {
		s.Log.Printf("DB CreateMedia failed: %v", err)
		s.jsonError(w, "database_error", "Failed to save metadata", http.StatusInternalServerError)
		return
	}
	// Log activity
	activityTargetName := "Bild in Galerie"
	if title != "" {
		activityTargetName = title
	}
	_, _ = s.DB.CreateActivity(r.Context(), database.CreateActivityParams{
		ID:         sid.New(),
		UserID:     user.ID,
		Type:       "MEDIA_UPLOADED",
		TargetID:   params.ID,
		TargetName: &activityTargetName,
	})

	s.respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (s *Server) GetMediaById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	m, err := s.DB.GetMedia(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.jsonError(w, "not_found", "Media not found", http.StatusNotFound)
		} else {
			s.Log.Printf("DB GetMedia error: %v", err)
			s.jsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	t, _ := time.Parse(time.RFC3339, m.UploadedAt)
	var desc string
	if m.Description != nil {
		desc = *m.Description
	}
	var title string
	if m.Title != nil {
		title = *m.Title
	}

	res := MediaResponse{
		ID:           m.ID,
		EventID:      m.EventID,
		Title:        title,
		Description:  desc,
		UploadedAt:   t,
		UploaderName: m.UploaderName,
		MimeType:     m.MimeType,
	}
	s.respondJSON(w, http.StatusOK, res)
}

func (s *Server) GetMediaFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	media, err := s.DB.GetMedia(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.jsonError(w, "not_found", "Media not found", http.StatusNotFound)
		} else {
			s.Log.Printf("DB GetMedia error: %v", err)
			s.jsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	objectKey := fmt.Sprintf("media/%s", media.Accesskey)
	obj, err := s.Store.GetObject(r.Context(), objectKey)
	if err != nil {
		s.Log.Printf("S3 GetObject error: %v", err)
		s.jsonError(w, "server_error", "Failed to retrieve file", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		s.jsonError(w, "not_found", "File content missing", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", media.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
	w.Header().Set("Cache-Control", "public, max-age=86400")

	if _, err := io.Copy(w, obj); err != nil {
		s.Log.Printf("Stream error: %v", err)
	}
}

func (s *Server) GetEventCover(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.jsonError(w, "bad_request", "Invalid event ID", http.StatusBadRequest)
		return
	}

	event, err := s.DB.GetEvent(r.Context(), id)
	if err != nil {
		s.jsonError(w, "not_found", "Event not found", http.StatusNotFound)
		return
	}

	if event.CoverPath == nil {
		s.jsonError(w, "not_found", "No cover image", http.StatusNotFound)
		return
	}

	obj, err := s.Store.GetObject(r.Context(), *event.CoverPath)
	if err != nil {
		s.jsonError(w, "server_error", "Failed to retrieve file", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		s.jsonError(w, "not_found", "File content missing", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", stat.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
	w.Header().Set("Cache-Control", "public, max-age=86400")

	if _, err := io.Copy(w, obj); err != nil {
		s.Log.Printf("Stream error: %v", err)
	}
}
