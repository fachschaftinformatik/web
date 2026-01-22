package handler

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/id"
	"github.com/fachschaftinformatik/web/internal/images"
	"github.com/fachschaftinformatik/web/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxMediaUploadSize = 256 << 20

// @Summary List events
// @Tags Events
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
		res = append(res, dto.EventResponse{
			ID:        id.ID(e.ID),
			Title:     e.Title,
			CoverPath: e.CoverPath,
			CreatedAt: e.CreatedAt,
		})
	}
	s.RespondJSON(w, http.StatusOK, res)
}

// @Summary Get event details
// @Tags Events
// @Param eventId path string true "Event ID"
// @Produce json
// @Success 200 {object} dto.EventResponse
// @Router /events/{eventId} [get]
func (s *Server) GetEventsId(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "eventId")
	eid, err := id.Parse(idStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid event ID", http.StatusBadRequest)
		return
	}

	event, err := s.DB.GetEvent(r.Context(), int64(eid))
	if err != nil {
		s.JsonError(w, "not_found", "Event not found", http.StatusNotFound)
		return
	}

	s.RespondJSON(w, http.StatusOK, dto.EventResponse{
		ID:        id.ID(event.ID),
		Title:     event.Title,
		CoverPath: event.CoverPath,
		CreatedAt: event.CreatedAt,
	})
}

// @Summary Create event
// @Tags Events
// @Accept mpfd
// @Produce json
// @Param title formData string true "Event Title"
// @Param file formData file false "Cover Image"
// @Success 201 {object} dto.EventResponse
// @Router /events [post]
func (s *Server) PostEvents(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)
	r.Body = http.MaxBytesReader(w, r.Body, maxMediaUploadSize+1024*1024)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		s.JsonError(w, "bad_request", "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	req := dto.CreateEventRequest{Title: title}
	if err := s.Validate(req); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	eventID := id.New()
	var coverPathPtr *string

	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		objectKey := storage.EventCoverSourceKey(eventID.String())
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		if err := s.Store.Upload(r.Context(), objectKey, file, header.Size, contentType); err != nil {
			s.Log.Error("S3 Upload failed", "err", err)
			s.JsonError(w, "server_error", "Failed to store cover image", http.StatusInternalServerError)
			return
		}

		// Generate and upload previews
		if strings.HasPrefix(contentType, "image/") {
			file.Seek(0, 0)
			previews, err := images.GeneratePreviews(file)
			if err == nil {
				for size, data := range previews {
					previewKey := storage.EventCoverPreviewKey(eventID.String(), size)
					_ = s.Store.Upload(r.Context(), previewKey, bytes.NewReader(data), int64(len(data)), "image/jpeg")
				}
			}
		}

		coverPathPtr = &objectKey
	}

	params := database.CreateEventParams{
		ID:        int64(eventID),
		Title:     s.Sanitize(title),
		CoverPath: coverPathPtr,
	}

	event, err := s.DB.CreateEvent(r.Context(), params)
	if err != nil {
		s.Log.Error("Failed to create event", "err", err)
		s.JsonError(w, "database_error", "Could not create event", http.StatusInternalServerError)
		return
	}

	link := "/events/" + id.ID(event.ID).String()
	s.broadcastNotification(r, "Neues Event", event.Title, "news", link)

	activityTargetName := event.Title
	_, _ = s.DB.CreateActivity(r.Context(), database.CreateActivityParams{
		ID:         int64(id.New()),
		UserID:     user.ID,
		Type:       "EVENT_CREATED",
		TargetID:   id.ID(event.ID).String(),
		TargetName: &activityTargetName,
	})

	s.RespondJSON(w, http.StatusCreated, dto.EventResponse{
		ID:        id.ID(event.ID),
		Title:     event.Title,
		CoverPath: event.CoverPath,
		CreatedAt: event.CreatedAt,
	})
}

// @Summary List media for an event
// @Tags Events
// @Param eventId path string true "Event ID"
// @Produce json
// @Success 200 {array} dto.MediaResponse
// @Router /events/{eventId}/media [get]
func (s *Server) GetEventMedia(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "eventId")
	eid, err := id.Parse(idStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid event ID", http.StatusBadRequest)
		return
	}

	media, err := s.DB.ListMediaByEvent(r.Context(), int64(eid))
	if err != nil {
		s.Log.Error("Failed to list media", "err", err)
		s.JsonError(w, "database_error", "Could not fetch media", http.StatusInternalServerError)
		return
	}

	res := make([]dto.MediaResponse, 0, len(media))
	for _, m := range media {
		title := ""
		if m.Title != nil {
			title = *m.Title
		}
		desc := ""
		if m.Description != nil {
			desc = *m.Description
		}
		res = append(res, dto.MediaResponse{
			ID:           id.ID(m.ID),
			EventID:      id.ID(m.EventID),
			Title:        title,
			Description:  desc,
			CreatedAt:    m.CreatedAt,
			UploaderName: m.UploaderName,
			MimeType:     m.MimeType,
		})
	}
	s.RespondJSON(w, http.StatusOK, res)
}

// @Summary Upload media to event
// @Tags Events
// @Accept mpfd
// @Produce json
// @Param eventId path string true "Event ID"
// @Param title formData string false "Title"
// @Param description formData string false "Description"
// @Param file formData file true "Media File"
// @Success 201 {object} dto.MediaResponse
// @Router /events/{eventId}/media [post]
func (s *Server) PostEventMedia(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)
	idStr := chi.URLParam(r, "eventId")
	eid, err := id.Parse(idStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid event ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(maxMediaUploadSize); err != nil {
		s.JsonError(w, "bad_request", "Upload failed", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.JsonError(w, "bad_request", "File upload failed", http.StatusBadRequest)
		return
	}
	defer file.Close()

	mediaID := id.New()
	objectKey := storage.MediaSourceKey(mediaID.String())
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if err := s.Store.Upload(r.Context(), objectKey, file, header.Size, contentType); err != nil {
		s.Log.Error("S3 Upload failed", "err", err)
		s.JsonError(w, "server_error", "Failed to store file", http.StatusInternalServerError)
		return
	}

	// Generate previews
	if strings.HasPrefix(contentType, "image/") {
		file.Seek(0, 0)
		previews, err := images.GeneratePreviews(file)
		if err == nil {
			for size, data := range previews {
				previewKey := storage.MediaPreviewKey(mediaID.String(), size)
				_ = s.Store.Upload(r.Context(), previewKey, bytes.NewReader(data), int64(len(data)), "image/jpeg")
			}
		}
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	var titlePtr, descPtr *string
	if title != "" {
		t := s.Sanitize(title)
		titlePtr = &t
	}
	if description != "" {
		d := s.Sanitize(description)
		descPtr = &d
	}

	m, err := s.DB.CreateMedia(r.Context(), database.CreateMediaParams{
		ID:          int64(mediaID),
		EventID:     int64(eid),
		UserID:      user.ID,
		Title:       titlePtr,
		Description: descPtr,
		AccessKey:   uuid.NewString(),
		MimeType:    contentType,
		Nbytes:      header.Size,
	})
	if err != nil {
		s.Log.Error("DB CreateMedia failed", "err", err)
		s.JsonError(w, "database_error", "Failed to save metadata", http.StatusInternalServerError)
		return
	}

	activityTargetName := "Bild in Galerie"
	if title != "" {
		activityTargetName = title
	}
	_, _ = s.DB.CreateActivity(r.Context(), database.CreateActivityParams{
		ID:         int64(id.New()),
		UserID:     user.ID,
		Type:       "MEDIA_UPLOADED",
		TargetID:   id.ID(eid).String(),
		TargetName: &activityTargetName,
	})

	link := "/events/" + id.ID(eid).String()
	s.broadcastNotification(r, "Neue Medien", activityTargetName, "news", link)

	s.RespondJSON(w, http.StatusCreated, dto.MediaResponse{
		ID:           id.ID(m.ID),
		EventID:      id.ID(m.EventID),
		Title:        title,
		Description:  description,
		CreatedAt:    m.CreatedAt,
		UploaderName: user.Name,
		MimeType:     m.MimeType,
	})
}

// @Summary Get media by ID
// @Tags Events
// @Param mediaId path string true "Media ID"
// @Produce json
// @Success 200 {object} dto.MediaResponse
// @Router /media/{mediaId} [get]
func (s *Server) GetMediaById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "mediaId")
	mid, err := id.Parse(idStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid media ID", http.StatusBadRequest)
		return
	}

	m, err := s.DB.GetMedia(r.Context(), int64(mid))
	if err != nil {
		s.JsonError(w, "not_found", "Media not found", http.StatusNotFound)
		return
	}

	title := ""
	if m.Title != nil {
		title = *m.Title
	}
	desc := ""
	if m.Description != nil {
		desc = *m.Description
	}

	s.RespondJSON(w, http.StatusOK, dto.MediaResponse{
		ID:           id.ID(m.ID),
		EventID:      id.ID(m.EventID),
		Title:        title,
		Description:  desc,
		CreatedAt:    m.CreatedAt,
		UploaderName: m.UploaderName,
		MimeType:     m.MimeType,
	})
}

// @Summary Get media file
// @Tags Events
// @Param mediaId path string true "Media ID"
// @Produce application/octet-stream
// @Success 200 {file} binary
// @Router /media/{mediaId}/file [get]
func (s *Server) GetMediaFile(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "mediaId")
	mid, err := id.Parse(idStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid media ID", http.StatusBadRequest)
		return
	}

	m, err := s.DB.GetMedia(r.Context(), int64(mid))
	if err != nil {
		s.JsonError(w, "not_found", "Media not found", http.StatusNotFound)
		return
	}

	objectKey := storage.MediaSourceKey(id.ID(m.ID).String())
	obj, err := s.Store.GetObject(r.Context(), objectKey)
	if err != nil {
		s.Log.Error("S3 GetObject error", "err", err)
		s.JsonError(w, "server_error", "Failed to retrieve file", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Type", obj.Info.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	io.Copy(w, obj)
}

// @Summary Get media preview
// @Tags Events
// @Param mediaId path string true "Media ID"
// @Param size query int false "Size (200, 400, 600, 800, 1200, 1600)"
// @Produce image/jpeg
// @Success 200 {file} binary
// @Router /media/{mediaId}/preview [get]
func (s *Server) GetMediaPreview(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "mediaId")
	mid, err := id.Parse(idStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid media ID", http.StatusBadRequest)
		return
	}

	m, err := s.DB.GetMedia(r.Context(), int64(mid))
	if err != nil {
		s.JsonError(w, "not_found", "Media not found", http.StatusNotFound)
		return
	}

	sizeStr := r.URL.Query().Get("size")
	size, _ := strconv.Atoi(sizeStr)
	if size == 0 {
		size = 400
	}

	previewKey := storage.MediaPreviewKey(id.ID(m.ID).String(), size)
	obj, err := s.Store.GetObject(r.Context(), previewKey)
	if err == nil {
		defer obj.Close()
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		_, _ = io.Copy(w, obj)
		return
	}

	// Try source as fallback
	originalKey := storage.MediaSourceKey(id.ID(m.ID).String())
	origObj, err := s.Store.GetObject(r.Context(), originalKey)
	if err == nil {
		defer origObj.Close()
		w.Header().Set("Content-Type", origObj.Info.ContentType)
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		_, _ = io.Copy(w, origObj)
		return
	}

	http.Error(w, "Media not found", http.StatusNotFound)
}

// @Summary Get event cover
// @Tags Events
// @Param eventId path string true "Event ID"
// @Param size query int false "Size (200, 400, 600, 800, 1200, 1600)"
// @Produce image/jpeg
// @Success 200 {file} binary
// @Router /events/{eventId}/cover [get]
func (s *Server) GetEventCover(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "eventId")
	eid, err := id.Parse(idStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid event ID", http.StatusBadRequest)
		return
	}

	event, err := s.DB.GetEvent(r.Context(), int64(eid))
	if err != nil {
		s.JsonError(w, "not_found", "Event not found", http.StatusNotFound)
		return
	}

	if event.CoverPath == nil {
		s.JsonError(w, "not_found", "No cover image", http.StatusNotFound)
		return
	}

	sizeStr := r.URL.Query().Get("size")
	size, _ := strconv.Atoi(sizeStr)
	if size == 0 {
		size = 800 // Default for cover
	}

	previewKey := storage.EventCoverPreviewKey(id.ID(event.ID).String(), size)
	obj, err := s.Store.GetObject(r.Context(), previewKey)
	if err == nil {
		defer obj.Close()
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		_, _ = io.Copy(w, obj)
		return
	}

	// Try source cover as fallback
	if event.CoverPath != nil {
		obj, err = s.Store.GetObject(r.Context(), *event.CoverPath)
		if err == nil {
			defer obj.Close()
			w.Header().Set("Content-Type", obj.Info.ContentType)
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			_, _ = io.Copy(w, obj)
			return
		}
	}

	http.Error(w, "Cover not found", http.StatusNotFound)
}
