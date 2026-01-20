package handler

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/id"
	"github.com/fachschaftinformatik/web/internal/images"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxMediaUploadSize = 256 << 20

// @Summary List events
// @Tags Media
// @Produce json
// @Success 200 {array} dto.EventResponse
// @Router /events [get]
func (s *Server) GetEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.DB.ListEvents(r.Context())
	if err != nil {
		s.Log.Error("Failed to list events", "err", err)
		s.JsonError(w, "database_error", "Could not fetch events", http.StatusInternalServerError)
		return
	}

	res := make([]dto.EventResponse, 0, len(events))
	for _, e := range events {
		var cover string
		if e.CoverPath != nil {
			cover = *e.CoverPath
		}
		res = append(res, dto.EventResponse{
			ID:        e.ID,
			Title:     e.Title,
			CoverPath: cover,
			CreatedAt: e.CreatedAt,
		})
	}
	s.RespondJSON(w, http.StatusOK, res)
}

// @Summary Create event
// @Tags Media
// @Accept mpfd
// @Produce json
// @Param title formData string true "Event Title"
// @Param file formData file false "Cover Image"
// @Success 201 {object} dto.EventResponse
// @Router /events [post]
func (s *Server) PostEvents(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMediaUploadSize+1024*1024)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		if errors.As(err, new(*http.MaxBytesError)) {
			s.JsonError(w, "bad_request", "File too large", http.StatusBadRequest)
		} else {
			s.JsonError(w, "bad_request", "Invalid form data", http.StatusBadRequest)
		}
		return
	}

	title := r.FormValue("title")
	req := dto.CreateEventRequest{Title: title}
	if err := s.Validate(req); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
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
			s.Log.Error("S3 Upload failed", "err", err)
			s.JsonError(w, "server_error", "Failed to store cover image", http.StatusInternalServerError)
			return
		}

		// Generate and upload preview for event cover
		if strings.HasPrefix(contentType, "image/") {
			file.Seek(0, 0)
			previewData, err := images.GeneratePreview(file, 800, 800)
			if err == nil {
				previewKey := fmt.Sprintf("previews/%s.jpg", objectKey)
				if err := s.Store.Upload(r.Context(), previewKey, bytes.NewReader(previewData), int64(len(previewData)), "image/jpeg"); err != nil {
					s.Log.Error("Failed to upload event cover preview", "key", previewKey, "err", err)
				}
			}
		}

		coverPathPtr = &objectKey
	}

	eventID := id.New()

	params := database.CreateEventParams{
		ID:        eventID,
		Title:     s.Sanitize(title),
		CoverPath: coverPathPtr,
	}

	event, err := s.DB.CreateEvent(r.Context(), params)
	if err != nil {
		s.Log.Error("Failed to create event", "err", err)
		s.JsonError(w, "database_error", "Could not create event", http.StatusInternalServerError)
		return
	}

	res := dto.EventResponse{
		ID:        event.ID,
		Title:     event.Title,
		CreatedAt: event.CreatedAt,
	}
	if event.CoverPath != nil {
		res.CoverPath = *event.CoverPath
	}

	s.RespondJSON(w, http.StatusCreated, res)
}

// @Summary List media
// @Tags Media
// @Param event_id query string true "Event ID"
// @Produce json
// @Success 200 {array} dto.MediaResponse
// @Router /media [get]
func (s *Server) GetMedia(w http.ResponseWriter, r *http.Request) {
	eventId := r.URL.Query().Get("event_id")
	if eventId == "" {
		s.JsonError(w, "bad_request", "Invalid event_id", http.StatusBadRequest)
		return
	}

	media, err := s.DB.ListMediaByEvent(r.Context(), eventId)
	if err != nil {
		s.Log.Error("Failed to list media", "err", err)
		s.JsonError(w, "database_error", "Could not fetch media", http.StatusInternalServerError)
		return
	}

	res := make([]dto.MediaResponse, 0, len(media))
	for _, m := range media {
		var desc string
		if m.Description != nil {
			desc = *m.Description
		}
		var title string
		if m.Title != nil {
			title = *m.Title
		}

		res = append(res, dto.MediaResponse{
			ID:           m.ID,
			EventID:      m.EventID,
			Title:        title,
			Description:  desc,
			UploadedAt:   m.UploadedAt,
			UploaderName: m.UploaderName,
			MimeType:     m.MimeType,
		})
	}
	s.RespondJSON(w, http.StatusOK, res)
}

// @Summary Upload media
// @Tags Media
// @Accept mpfd
// @Produce json
// @Param event_id formData string true "Event ID"
// @Param title formData string false "Title"
// @Param description formData string false "Description"
// @Param file formData file true "Media File"
// @Success 201 {object} map[string]string
// @Router /media [post]
func (s *Server) PostMedia(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)

	r.Body = http.MaxBytesReader(w, r.Body, maxMediaUploadSize+1024*1024)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		if errors.As(err, new(*http.MaxBytesError)) {
			s.JsonError(w, "bad_request", "File too large", http.StatusBadRequest)
		} else {
			s.Log.Error("Multipart parse error", "err", err)
			s.JsonError(w, "bad_request", "Upload failed", http.StatusBadRequest)
		}
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.JsonError(w, "bad_request", "File upload failed", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > maxMediaUploadSize {
		s.JsonError(w, "bad_request", "File too large", http.StatusBadRequest)
		return
	}

	eventId := r.FormValue("event_id")
	if eventId == "" {
		s.JsonError(w, "bad_request", "Invalid event_id", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	req := dto.CreateMediaRequest{
		EventID:     eventId,
		Title:       title,
		Description: description,
	}
	if err := s.Validate(req); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	accessKey := uuid.NewString()
	objectKey := fmt.Sprintf("media/%s", accessKey)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	file.Seek(0, 0)
	if err := s.Store.Upload(r.Context(), objectKey, file, header.Size, contentType); err != nil {
		s.Log.Error("S3 Upload failed", "err", err)
		s.JsonError(w, "server_error", "Failed to store file", http.StatusInternalServerError)
		return
	}

	// Generate and upload preview
	if strings.HasPrefix(contentType, "image/") {
		file.Seek(0, 0)
		previewData, err := images.GeneratePreview(file, 400, 400)
		if err == nil {
			previewKey := fmt.Sprintf("previews/%s.jpg", objectKey)
			if err := s.Store.Upload(r.Context(), previewKey, bytes.NewReader(previewData), int64(len(previewData)), "image/jpeg"); err != nil {
				s.Log.Error("Failed to upload preview", "key", previewKey, "err", err)
			}
		} else {
			s.Log.Error("Failed to generate preview", "key", objectKey, "err", err)
		}
	}

	var titlePtr, descPtr *string
	if title != "" {
		t := s.Sanitize(title)
		titlePtr = &t
	}
	if description != "" {
		d := s.Sanitize(description)
		descPtr = &d
	}

	params := database.CreateMediaParams{
		ID:          id.New(),
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
		s.Log.Error("DB CreateMedia failed", "err", err)
		s.JsonError(w, "database_error", "Failed to save metadata", http.StatusInternalServerError)
		return
	}

	activityTargetName := "Bild in Galerie"
	if title != "" {
		activityTargetName = *titlePtr
	}
	_, _ = s.DB.CreateActivity(r.Context(), database.CreateActivityParams{
		ID:         id.New(),
		UserID:     user.ID,
		Type:       "MEDIA_UPLOADED",
		TargetID:   params.ID,
		TargetName: &activityTargetName,
	})

	s.RespondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

// @Summary Get media by ID
// @Tags Media
// @Param id path string true "Media ID"
// @Produce json
// @Success 200 {object} dto.MediaResponse
// @Router /media/{id} [get]
func (s *Server) GetMediaById(w http.ResponseWriter, r *http.Request) {
	mediaID := chi.URLParam(r, "id")

	m, err := s.DB.GetMedia(r.Context(), mediaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.JsonError(w, "not_found", "Media not found", http.StatusNotFound)
		} else {
			s.Log.Error("DB GetMedia error", "err", err)
			s.JsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	var desc string
	if m.Description != nil {
		desc = *m.Description
	}
	var title string
	if m.Title != nil {
		title = *m.Title
	}

	res := dto.MediaResponse{
		ID:           m.ID,
		EventID:      m.EventID,
		Title:        title,
		Description:  desc,
		UploadedAt:   m.UploadedAt,
		UploaderName: m.UploaderName,
		MimeType:     m.MimeType,
	}
	s.RespondJSON(w, http.StatusOK, res)
}

// @Summary Get media file
// @Tags Media
// @Param id path string true "Media ID"
// @Produce application/octet-stream
// @Success 200 {file} binary
// @Router /media/{id}/file [get]
func (s *Server) GetMediaFile(w http.ResponseWriter, r *http.Request) {
	mediaID := chi.URLParam(r, "id")

	media, err := s.DB.GetMedia(r.Context(), mediaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.JsonError(w, "not_found", "Media not found", http.StatusNotFound)
		} else {
			s.Log.Error("DB GetMedia error", "err", err)
			s.JsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	objectKey := fmt.Sprintf("media/%s", media.Accesskey)
	obj, err := s.Store.GetObject(r.Context(), objectKey)
	if err != nil {
		s.Log.Error("S3 GetObject error", "err", err)
		s.JsonError(w, "server_error", "Failed to retrieve file", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		s.JsonError(w, "not_found", "File content missing", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", media.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
	w.Header().Set("Cache-Control", "public, max-age=86400")

	if _, err := io.Copy(w, obj); err != nil {
		s.Log.Error("Stream error", "err", err)
	}
}

// @Summary Get media preview
// @Tags Media
// @Param id path string true "Media ID"
// @Produce image/jpeg
// @Success 200 {file} binary
// @Router /media/{id}/preview [get]
func (s *Server) GetMediaPreview(w http.ResponseWriter, r *http.Request) {
	mediaID := chi.URLParam(r, "id")

	media, err := s.DB.GetMedia(r.Context(), mediaID)
	if err != nil {
		s.JsonError(w, "not_found", "Media not found", http.StatusNotFound)
		return
	}

	originalKey := fmt.Sprintf("media/%s", media.Accesskey)
	previewKey := fmt.Sprintf("previews/%s.jpg", originalKey)

	obj, err := s.Store.GetObject(r.Context(), previewKey)
	if err == nil {
		defer obj.Close()
		stat, _ := obj.Stat()
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		io.Copy(w, obj)
		return
	}

	// Lazy generate
	origObj, err := s.Store.GetObject(r.Context(), originalKey)
	if err != nil {
		s.JsonError(w, "not_found", "Original file missing", http.StatusNotFound)
		return
	}
	defer origObj.Close()

	previewData, err := images.GeneratePreview(origObj, 400, 400)
	if err != nil {
		s.Log.Error("Lazy preview generation failed", "id", mediaID, "err", err)
		s.GetMediaFile(w, r)
		return
	}

	go func() {
		if err := s.Store.Upload(context.Background(), previewKey, bytes.NewReader(previewData), int64(len(previewData)), "image/jpeg"); err != nil {
			s.Log.Error("Failed to cache lazy preview", "key", previewKey, "err", err)
		}
	}()

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(previewData)))
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(previewData)
}

// @Summary Get event cover
// @Tags Media
// @Param id path string true "Event ID"
// @Produce image/jpeg
// @Success 200 {file} binary
// @Router /events/{id}/cover [get]
func (s *Server) GetEventCover(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	event, err := s.DB.GetEvent(r.Context(), idStr)
	if err != nil {
		s.JsonError(w, "not_found", "Event not found", http.StatusNotFound)
		return
	}

	if event.CoverPath == nil {
		s.JsonError(w, "not_found", "No cover image", http.StatusNotFound)
		return
	}

	// Try to serve optimized version if it's an internal path
	coverPath := *event.CoverPath
	if strings.HasPrefix(coverPath, "events/") {
		previewKey := fmt.Sprintf("previews/%s.jpg", coverPath)

		obj, err := s.Store.GetObject(r.Context(), previewKey)
		if err == nil {
			defer obj.Close()
			stat, _ := obj.Stat()
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			io.Copy(w, obj)
			return
		}

		// Lazy generate 800px cover
		origObj, err := s.Store.GetObject(r.Context(), coverPath)
		if err == nil {
			defer origObj.Close()
			previewData, err := images.GeneratePreview(origObj, 800, 800)
			if err == nil {
				go func() {
					if err := s.Store.Upload(context.Background(), previewKey, bytes.NewReader(previewData), int64(len(previewData)), "image/jpeg"); err != nil {
						s.Log.Error("Failed to cache event cover preview", "key", previewKey, "err", err)
					}
				}()
				w.Header().Set("Content-Type", "image/jpeg")
				w.Header().Set("Content-Length", fmt.Sprintf("%d", len(previewData)))
				w.Header().Set("Cache-Control", "public, max-age=31536000")
				w.Write(previewData)
				return
			}
		}
	}

	// Fallback to original
	obj, err := s.Store.GetObject(r.Context(), coverPath)
	if err != nil {
		s.JsonError(w, "server_error", "Failed to retrieve file", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		s.JsonError(w, "not_found", "File content missing", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", stat.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
	w.Header().Set("Cache-Control", "public, max-age=86400")

	if _, err := io.Copy(w, obj); err != nil {
		s.Log.Error("Stream error", "err", err)
	}
}
