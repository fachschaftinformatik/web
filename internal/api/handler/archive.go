package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/id"
	"github.com/fachschaftinformatik/web/internal/storage"
	"github.com/go-chi/chi/v5"
)

// @Summary List archive entry versions
// @Tags Archive
// @ID getArchiveVersions
// @Param entryId path string true "Entry ID"
// @Success 200 {array} dto.ArchiveEntryResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /archive/{entryId}/versions [get]
func (s *Server) GetArchiveVersions(w http.ResponseWriter, r *http.Request) {
	entryIDStr := chi.URLParam(r, "entryId")
	eid, err := id.Parse(entryIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid entry ID", http.StatusBadRequest)
		return
	}

	baseArchive, err := s.DB.GetArchiveEntry(r.Context(), int64(eid))
	if err != nil {
		s.JsonError(w, "not_found", "Entry not found", http.StatusNotFound)
		return
	}

	rows, err := s.DB.ListArchiveVersions(r.Context(), baseArchive.GroupID)
	if err != nil {
		s.Log.Error("Failed to list archive versions", "err", err)
		s.JsonError(w, "database_error", "Could not fetch history", http.StatusInternalServerError)
		return
	}

	resp := make([]dto.ArchiveEntryResponse, 0, len(rows))
	for _, row := range rows {
		comment := ""
		if row.Comment != nil {
			comment = *row.Comment
		}

		mod, _ := s.DB.GetModule(r.Context(), row.ModuleID)

		resp = append(resp, dto.ArchiveEntryResponse{
			ID:           id.ID(row.ID),
			ModuleID:     id.ID(row.ModuleID),
			ModuleName:   mod.Name,
			ProgramID:    id.ID(mod.ProgramID),
			Version:      row.Version,
			ExamDate:     row.ExamDate,
			CreatedAt:    row.CreatedAt,
			UploaderName: row.UploaderName,
			Comment:      comment,
			EditVersion:  row.EditVersion,
			FileID:       id.ID(row.ID),
		})
	}

	s.RespondJSON(w, http.StatusOK, resp)
}

// @Summary Get archive entry details
// @Tags Archive
// @ID getArchiveId
// @Param entryId path string true "Entry ID"
// @Success 200 {object} dto.ArchiveEntryResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /archive/{entryId} [get]
func (s *Server) GetArchiveId(w http.ResponseWriter, r *http.Request) {
	entryIDStr := chi.URLParam(r, "entryId")
	eid, err := id.Parse(entryIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid entry ID", http.StatusBadRequest)
		return
	}

	row, err := s.DB.GetArchiveEntryDetails(r.Context(), int64(eid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.JsonError(w, "not_found", "Entry not found", http.StatusNotFound)
		} else {
			s.Log.Error("DB GetArchiveEntryDetails error", "err", err)
			s.JsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	comm := ""
	if row.Comment != nil {
		comm = *row.Comment
	}

	resp := dto.ArchiveEntryResponse{
		ID:           id.ID(row.ID),
		ModuleID:     id.ID(row.ModuleID),
		ModuleName:   row.ModuleName,
		ProgramID:    id.ID(row.ProgramID),
		Version:      row.Version,
		ExamDate:     row.ExamDate,
		CreatedAt:    row.CreatedAt,
		UploaderName: row.UploaderName,
		Comment:      comm,
		EditVersion:  row.EditVersion,
		FileID:       id.ID(row.ID),
	}

	s.RespondJSON(w, http.StatusOK, resp)
}

// @Summary Upload archive entry
// @Tags Archive
// @ID postArchive
// @Accept mpfd
// @Produce json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Param file formData file true "Exam PDF"
// @Param date formData string false "Exam Date (YYYY-MM-DD)"
// @Param module_id formData string false "Module ID"
// @Param version formData string false "PO Version"
// @Param group_id formData string false "Optional specific group ID to add revision to"
// @Param comment formData string false "Optional comment"
// @Success 201 {object} dto.ArchiveEntryResponse
// @Router /archive [post]
func (s *Server) PostArchive(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		s.JsonError(w, "bad_request", "Upload failed", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.JsonError(w, "bad_request", "File upload failed", http.StatusBadRequest)
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if mimeType != "application/pdf" {
		s.JsonError(w, "bad_request", "Only PDF allowed", http.StatusBadRequest)
		return
	}

	examDate := strings.TrimSpace(r.FormValue("date"))
	moduleIDStr := strings.TrimSpace(r.FormValue("module_id"))
	version := strings.TrimSpace(r.FormValue("version"))
	comment := strings.TrimSpace(r.FormValue("comment"))
	groupIDStr := strings.TrimSpace(r.FormValue("group_id"))

	hasher := sha256.New()
	_, _ = io.Copy(hasher, file)
	checksum := hex.EncodeToString(hasher.Sum(nil))

	var groupID int64
	var mid int64
	var editVersion int64 = 1

	if groupIDStr != "" {
		gVal, err := id.Parse(groupIDStr)
		if err != nil {
			s.JsonError(w, "invalid_id", "Invalid group ID", http.StatusBadRequest)
			return
		}
		groupID = int64(gVal)

		latest, err := s.DB.GetLatestByGroupId(r.Context(), groupID)
		if err != nil {
			s.JsonError(w, "not_found", "Group not found", http.StatusNotFound)
			return
		}
		editVersion = latest.EditVersion + 1
		mid = latest.ModuleID
		if examDate == "" {
			examDate = latest.ExamDate
		}
		if version == "" {
			version = latest.Version
		}

		_ = s.DB.ClearLatestFlagByGroupId(r.Context(), groupID)
	} else {
		if moduleIDStr == "" || version == "" || examDate == "" {
			s.JsonError(w, "bad_request", "Missing metadata for new entry", http.StatusBadRequest)
			return
		}
		mVal, err := id.Parse(moduleIDStr)
		if err != nil {
			s.JsonError(w, "invalid_id", "Invalid module ID", http.StatusBadRequest)
			return
		}
		mid = int64(mVal)

		existing, err := s.DB.GetArchiveEntryByMetadata(r.Context(), database.GetArchiveEntryByMetadataParams{
			ModuleID: mid,
			Version:  version,
			ExamDate: examDate,
		})
		if err == nil {
			groupID = existing.GroupID
			latest, _ := s.DB.GetLatestByGroupId(r.Context(), groupID)
			editVersion = latest.EditVersion + 1
			_ = s.DB.ClearLatestFlagByGroupId(r.Context(), groupID)
		} else {
			newID := int64(id.New())
			groupID = newID
		}
	}

	newID := int64(id.New())
	if groupID == 0 {
		groupID = newID
	}

	objectKey := storage.ArchiveSourceKey(id.ID(newID).String())
	_, _ = file.Seek(0, 0)
	if err := s.Store.Upload(r.Context(), objectKey, file, header.Size, "application/pdf"); err != nil {
		s.Log.Error("S3 Upload failed", "err", err)
		s.JsonError(w, "storage_error", "Upload failed", http.StatusInternalServerError)
		return
	}

	var commentPtr *string
	if comment != "" {
		c := s.Sanitize(comment)
		commentPtr = &c
	}

	entry, err := s.DB.CreateArchiveEntry(r.Context(), database.CreateArchiveEntryParams{
		ID:          newID,
		GroupID:     groupID,
		UserID:      user.ID,
		ModuleID:    mid,
		Version:     version,
		ExamDate:    examDate,
		Comment:     commentPtr,
		EditVersion: editVersion,
		IsLatest:    1,
		AccessKey:   id.New().String(),
		MimeType:    "application/pdf",
		Nbytes:      header.Size,
		Checksum:    checksum,
	})
	if err != nil {
		s.Log.Error("Failed to create archive entry", "err", err)
		s.JsonError(w, "database_error", "Failed to save entry", http.StatusInternalServerError)
		return
	}

	moduleName := "einem Modul"
	mod, _ := s.DB.GetModule(r.Context(), entry.ModuleID)
	if mod.Name != "" {
		moduleName = mod.Name
	}

	msg := fmt.Sprintf("%s (%s)", moduleName, entry.ExamDate)
	link := fmt.Sprintf("/archive/%d/%d", entry.ModuleID, entry.ID)
	s.broadcastNotification(r, "Neue Klausur", msg, "archive", link)

	_, _ = s.DB.CreateActivity(r.Context(), database.CreateActivityParams{
		ID:         int64(id.New()),
		UserID:     user.ID,
		Type:       "EXAM_UPLOADED",
		TargetID:   fmt.Sprintf("%d/%d", entry.ModuleID, entry.ID),
		TargetName: &moduleName,
	})

	s.RespondJSON(w, http.StatusCreated, dto.ArchiveEntryResponse{
		ID:           id.ID(entry.ID),
		ModuleID:     id.ID(entry.ModuleID),
		ModuleName:   moduleName,
		ProgramID:    id.ID(mod.ProgramID),
		Version:      entry.Version,
		ExamDate:     entry.ExamDate,
		CreatedAt:    entry.CreatedAt,
		UploaderName: user.Name,
		Comment:      comment,
		EditVersion:  editVersion,
		FileID:       id.ID(entry.ID),
	})
}

// @Summary List archive entries
// @Tags Archive
// @ID getArchive
// @Param program_id query string false "Program ID"
// @Param version query string false "Version"
// @Param module_id query string false "Module ID"
// @Success 200 {array} dto.ArchiveEntryResponse
// @Router /archive [get]
func (s *Server) GetArchive(w http.ResponseWriter, r *http.Request) {
	dbParams := database.ListArchiveEntriesParams{}

	if pidStr := r.URL.Query().Get("program_id"); pidStr != "" {
		pid, _ := id.Parse(pidStr)
		v := int64(pid)
		dbParams.ProgramID = &v
	}
	if ver := r.URL.Query().Get("version"); ver != "" {
		dbParams.Version = &ver
	}
	if midStr := r.URL.Query().Get("module_id"); midStr != "" {
		mid, _ := id.Parse(midStr)
		v := int64(mid)
		dbParams.ModuleID = &v
	}

	rows, err := s.DB.ListArchiveEntries(r.Context(), dbParams)
	if err != nil {
		s.Log.Error("Failed to list archive entries", "err", err)
		s.JsonError(w, "database_error", err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]dto.ArchiveEntryResponse, 0, len(rows))
	for _, row := range rows {
		comm := ""
		if row.Comment != nil {
			comm = *row.Comment
		}
		resp = append(resp, dto.ArchiveEntryResponse{
			ID:           id.ID(row.ID),
			ModuleID:     id.ID(row.ModuleID),
			ModuleName:   row.ModuleName,
			ProgramID:    id.ID(row.ProgramID),
			Version:      row.Version,
			ExamDate:     row.ExamDate,
			CreatedAt:    row.CreatedAt,
			UploaderName: row.UploaderName,
			Comment:      comm,
			EditVersion:  row.EditVersion,
			FileID:       id.ID(row.ID),
		})
	}

	s.RespondJSON(w, http.StatusOK, resp)
}

// @Summary Update archive entry (new version)
// @Tags Archive
// @ID putArchiveId
// @Param entryId path string true "Entry ID"
// @Param body body dto.UpdateArchiveEntryRequest true "Update Data"
// @Success 200 {object} dto.ArchiveEntryResponse
// @Router /archive/{entryId} [put]
func (s *Server) PutArchiveId(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)
	entryIDStr := chi.URLParam(r, "entryId")
	eid, err := id.Parse(entryIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid entry ID", http.StatusBadRequest)
		return
	}

	var payload dto.UpdateArchiveEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "bad_request", "Invalid JSON body", http.StatusBadRequest)
		return
	}

	baseArchRow, err := s.DB.GetArchiveEntry(r.Context(), int64(eid))
	if err != nil {
		s.JsonError(w, "not_found", "Entry not found", http.StatusNotFound)
		return
	}

	_ = s.DB.ClearLatestFlagByGroupId(r.Context(), baseArchRow.GroupID)

	revUID := int64(id.New())

	oldRevFileKey := storage.ArchiveSourceKey(id.ID(baseArchRow.ID).String())
	newRevFileKey := storage.ArchiveSourceKey(id.ID(revUID).String())
	_ = s.Store.CopyObject(r.Context(), oldRevFileKey, newRevFileKey)

	finRevComment := strings.TrimSpace(payload.Comment)
	var finRevCommentPtr *string
	if finRevComment != "" {
		finRevCommentPtr = &finRevComment
	}

	finalArchiveRevRow, err := s.DB.CreateArchiveEntry(r.Context(), database.CreateArchiveEntryParams{
		ID:          revUID,
		GroupID:     baseArchRow.GroupID,
		UserID:      user.ID,
		ModuleID:    int64(payload.ModuleID),
		Version:     payload.Version,
		ExamDate:    payload.Date,
		Comment:     finRevCommentPtr,
		EditVersion: baseArchRow.EditVersion + 1,
		IsLatest:    1,
		AccessKey:   id.New().String(),
		MimeType:    baseArchRow.MimeType,
		Nbytes:      baseArchRow.Nbytes,
		Checksum:    baseArchRow.Checksum,
	})

	if err != nil {
		s.Log.Error("Update failed", "err", err)
		s.JsonError(w, "database_error", "Update failed", http.StatusInternalServerError)
		return
	}

	link := fmt.Sprintf("/archive/%d/%d", int64(payload.ModuleID), finalArchiveRevRow.ID)
	s.broadcastNotification(r, "Neue Revision", baseArchRow.ExamDate, "archive", link)

	// Return the new ID so frontend can redirect/refresh
	row, _ := s.DB.GetArchiveEntryDetails(r.Context(), finalArchiveRevRow.ID)

	comm := ""
	if row.Comment != nil {
		comm = *row.Comment
	}

	s.RespondJSON(w, http.StatusOK, dto.ArchiveEntryResponse{
		ID:           id.ID(row.ID),
		ModuleID:     id.ID(row.ModuleID),
		ModuleName:   row.ModuleName,
		ProgramID:    id.ID(row.ProgramID),
		Version:      row.Version,
		ExamDate:     row.ExamDate,
		CreatedAt:    row.CreatedAt,
		UploaderName: row.UploaderName,
		Comment:      comm,
		EditVersion:  row.EditVersion,
		FileID:       id.ID(row.ID),
	})
}

// @Summary Delete archive entry
// @Tags Archive
// @ID deleteArchiveId
// @Param entryId path string true "Entry ID"
// @Success 204
// @Router /archive/{entryId} [delete]
func (s *Server) DeleteArchiveId(w http.ResponseWriter, r *http.Request) {
	entryIDStr := chi.URLParam(r, "entryId")
	eid, err := id.Parse(entryIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid entry ID", http.StatusBadRequest)
		return
	}

	entry, err := s.DB.GetArchiveEntry(r.Context(), int64(eid))
	if err != nil {
		s.JsonError(w, "not_found", "Entry not found", http.StatusNotFound)
		return
	}

	if err := s.DB.DeleteArchiveEntry(r.Context(), entry.ID); err != nil {
		s.JsonError(w, "database_error", "Delete failed", http.StatusInternalServerError)
		return
	}

	// Promote next latest in group
	_ = s.DB.PromoteLatestInGroup(r.Context(), entry.GroupID)

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Delete a specific archive file version
// @Tags Archive
// @ID deleteArchiveFile
// @Param fileId path string true "File ID"
// @Success 204
// @Router /archive/files/{fileId} [delete]
func (s *Server) DeleteArchiveFile(w http.ResponseWriter, r *http.Request) {
	// In consolidated mode, ArchiveFile IS ArchiveEntry.
	s.DeleteArchiveId(w, r)
}

// @Summary Download archive file
// @Tags Archive
// @ID getArchiveFile
// @Param entryId path string true "Entry ID"
// @Param file_id query string false "Specific File ID (ignored in simple mode)"
// @Success 200 {file} file
// @Router /archive/{entryId}/file [get]
func (s *Server) GetArchiveFile(w http.ResponseWriter, r *http.Request) {
	entryIDStr := chi.URLParam(r, "entryId")
	eid, err := id.Parse(entryIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid entry ID", http.StatusBadRequest)
		return
	}

	entry, err := s.DB.GetArchiveEntry(r.Context(), int64(eid))
	if err != nil {
		s.JsonError(w, "not_found", "File not found", http.StatusNotFound)
		return
	}

	objectKey := storage.ArchiveSourceKey(id.ID(entry.ID).String())
	obj, err := s.Store.GetObject(r.Context(), objectKey)
	if err != nil {
		s.JsonError(w, "server_error", "Failed to retrieve file", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Type", "application/pdf")
	_, _ = io.Copy(w, obj)
}
