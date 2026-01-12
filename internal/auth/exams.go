package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// TODO: Move to env
const maxUploadSize = 10 << 20 // 10 MB

// @Description Exam details
type ExamResponse struct {
	ID           string    `json:"id"`
	ProgramID    int64     `json:"programid"`
	Version      string    `json:"version"`
	ModuleID     int64     `json:"moduleid"`
	ModuleName   string    `json:"module_name"`
	ExamDate     string    `json:"exam_date" format:"date" example:"2023-01-15"`
	UploadedAt   time.Time `json:"uploaded_at"`
	UploaderName string    `json:"uploader_name"`
	Comment      string    `json:"comment,omitempty"`
	EditVersion  int64     `json:"edit_version"`
	GroupID      string    `json:"group_id"`
	IsLatest     int64     `json:"is_latest"`
}

// GetExamVersions list exam versions
// @Summary List exam versions
// @Tags Exams
// @ID getExamVersions
// @Param groupId path string true "Exam Group ID"
// @Success 200 {array} ExamResponse
// @Failure 401 {object} ErrorResponse
// @Router /exams/versions/{groupId} [get]
func (s *Server) GetExamVersions(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")
	_, _, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	rows, err := s.DB.ListExamVersions(r.Context(), groupID)
	if err != nil {
		s.Log.Printf("Failed to list exam versions: %v", err)
		s.jsonError(w, "database_error", "Could not fetch exam history", http.StatusInternalServerError)
		return
	}

	apiExams := make([]ExamResponse, 0, len(rows))
	for _, row := range rows {
		uploaded, _ := time.Parse(time.RFC3339, row.UploadedAt)

		var modID int64
		if row.Moduleid != nil {
			modID = *row.Moduleid
		}
		var comm string
		if row.Comment != nil {
			comm = *row.Comment
		}

		entry := ExamResponse{
			ID:           row.ID,
			ProgramID:    row.Programid,
			Version:      row.Version,
			ModuleID:     modID,
			ModuleName:   row.ModuleName,
			ExamDate:     row.ExamDate,
			UploadedAt:   uploaded,
			UploaderName: row.UploaderName,
			Comment:      comm,
			EditVersion:  row.EditVersion,
			GroupID:      row.GroupID,
			IsLatest:     row.IsLatest,
		}
		apiExams = append(apiExams, entry)
	}

	s.respondJSON(w, http.StatusOK, apiExams)
}

type ExamAssignment struct {
	ProgramID int64  `json:"programid"`
	Version   string `json:"version"`
	ModuleID  int64  `json:"moduleid"`
}

type UpdateExamRequest struct {
	ProgramID int64  `json:"programid"`
	Version   string `json:"version"`
	ModuleID  int64  `json:"moduleid"`
	Date      string `json:"date"`
	Comment   string `json:"comment"`
}

// PostExams uploads a new exam with multiple assignments
// @Summary Upload exam
// @Description Upload a PDF and assign it to one or more modules
// @Tags Exams
// @ID postExams
// @Accept mpfd
// @Produce json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Param file formData file true "Exam PDF"
// @Param date formData string true "Exam Date (YYYY-MM-DD)"
// @Param assignments formData string true "JSON array of assignments: [{'programid':1, 'version':'PO2016', 'moduleid':10}]"
// @Param comment formData string false "Optional comment"
// @Success 201 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /exams [post]
func (s *Server) PostExams(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		s.jsonError(w, "bad_request", "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.jsonError(w, "bad_request", "File upload failed", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > maxUploadSize {
		s.jsonError(w, "bad_request", "File too large", http.StatusBadRequest)
		return
	}
	mimeType := header.Header.Get("Content-Type")
	if mimeType != "application/pdf" {
		s.jsonError(w, "bad_request", "Only PDF files are allowed", http.StatusBadRequest)
		return
	}

	examDate := r.FormValue("date")
	comment := r.FormValue("comment")
	assignmentsJSON := r.FormValue("assignments")

	if examDate == "" || assignmentsJSON == "" {
		s.jsonError(w, "bad_request", "Missing metadata", http.StatusBadRequest)
		return
	}

	var assignments []ExamAssignment
	if err := json.Unmarshal([]byte(assignmentsJSON), &assignments); err != nil || len(assignments) == 0 {
		s.jsonError(w, "bad_request", "Invalid assignments JSON", http.StatusBadRequest)
		return
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		s.jsonError(w, "server_error", "Failed to process file", http.StatusInternalServerError)
		return
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))

	file.Seek(0, 0)
	accessKey := uuid.NewString()
	objectKey := fmt.Sprintf("exams/%s.pdf", accessKey)

	if err := s.Store.Upload(r.Context(), objectKey, file, header.Size, "application/pdf"); err != nil {
		s.Log.Printf("S3 Upload failed: %v", err)
		s.jsonError(w, "server_error", "Failed to store file", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	var commentPtr *string
	if comment != "" {
		commentPtr = &comment
	}

	for _, assign := range assignments {
		modID := assign.ModuleID
		id := uuid.NewString()
		paramsDB := database.CreateExamParams{
			ID:          id,
			Userid:      user.ID,
			Programid:   assign.ProgramID,
			Version:     assign.Version,
			Moduleid:    &modID,
			ExamDate:    examDate,
			Accesskey:   accessKey,
			MimeType:    "application/pdf",
			Nbytes:      header.Size,
			Checksum:    checksum,
			Comment:     commentPtr,
			GroupID:     id, // Use ID as group_id for new exams
			EditVersion: 1,
			IsLatest:    1,
		}

		_, err = s.DB.CreateExam(ctx, paramsDB)
		if err != nil {
			s.Log.Printf("DB CreateExam failed for assignment %+v: %v", assign, err)
		} else {
			// Create a notification for each module assignment
			// We try to find the module name for a better message
			moduleName := "einem Modul"
			mod, err := s.DB.GetModule(ctx, modID)
			if err == nil {
				moduleName = mod.Name
			}

			msg := fmt.Sprintf("Eine neue Klausur für %s (%s) wurde hochgeladen.", moduleName, examDate)
			link := fmt.Sprintf("/rekos/klausuren/modul?modulId=%d&mod=%s&examId=%s", modID, moduleName, id)
			s.broadcastNotification(r, "Neue Klausur verfügbar", msg, "exam", link)
		}
	}

	s.respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

// GetExams list exams
// @Summary List exams
// @Tags Exams
// @ID getExams
// @Param programid query int false "Program ID"
// @Param version query string false "Version"
// @Param moduleid query int false "Module ID"
// @Success 200 {array} ExamResponse
// @Failure 401 {object} ErrorResponse
// @Router /exams [get]
func (s *Server) GetExams(w http.ResponseWriter, r *http.Request) {
	_, _, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	dbParams := database.ListExamsParams{}

	if pid := r.URL.Query().Get("programid"); pid != "" {
		if val, err := strconv.ParseInt(pid, 10, 64); err == nil {
			dbParams.Programid = &val
		}
	}
	if ver := r.URL.Query().Get("version"); ver != "" {
		dbParams.Version = &ver
	}
	if mid := r.URL.Query().Get("moduleid"); mid != "" {
		if val, err := strconv.ParseInt(mid, 10, 64); err == nil {
			dbParams.Moduleid = &val
		}
	}

	rows, err := s.DB.ListExams(r.Context(), dbParams)
	if err != nil {
		s.Log.Printf("Failed to list exams: %v", err)
		s.jsonError(w, "database_error", "Could not fetch exams", http.StatusInternalServerError)
		return
	}

	apiExams := make([]ExamResponse, 0, len(rows))
	for _, row := range rows {
		uploaded, _ := time.Parse(time.RFC3339, row.UploadedAt)

		var modID int64
		if row.Moduleid != nil {
			modID = *row.Moduleid
		}
		var comm string
		if row.Comment != nil {
			comm = *row.Comment
		}

		entry := ExamResponse{
			ID:           row.ID,
			ProgramID:    row.Programid,
			Version:      row.Version,
			ModuleID:     modID,
			ModuleName:   row.ModuleName,
			ExamDate:     row.ExamDate,
			UploadedAt:   uploaded,
			UploaderName: row.UploaderName,
			Comment:      comm,
			EditVersion:  row.EditVersion,
			GroupID:      row.GroupID,
			IsLatest:     row.IsLatest,
		}
		apiExams = append(apiExams, entry)
	}

	s.respondJSON(w, http.StatusOK, apiExams)
}

// PutExamsId updates an exam
// @Summary Update exam
// @Tags Exams
// @ID putExamsId
// @Param id path string true "Exam ID"
// @Param body body UpdateExamRequest true "Update Data"
// @Success 200 {object} ExamResponse
// @Router /exams/{id} [put]
func (s *Server) PutExamsId(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if user.Role != "admin" && user.Role != "editor" {
		s.jsonError(w, "forbidden", "Insufficient permissions", http.StatusForbidden)
		return
	}

	var payload UpdateExamRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.jsonError(w, "bad_request", "Invalid JSON body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	// 1. Get current version
	oldExam, err := s.DB.GetExam(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.jsonError(w, "not_found", "Exam not found", http.StatusNotFound)
		} else {
			s.Log.Printf("DB GetExam error: %v", err)
			s.jsonError(w, "database_error", "Failed to fetch exam", http.StatusInternalServerError)
		}
		return
	}

	// 2. Prepare new access key and duplicate file
	newAccessKey := uuid.NewString()
	oldPath := fmt.Sprintf("exams/%s.pdf", oldExam.Accesskey)
	newPath := fmt.Sprintf("exams/%s.pdf", newAccessKey)

	if err := s.Store.CopyObject(ctx, oldPath, newPath); err != nil {
		s.Log.Printf("Storage CopyObject error: %v", err)
		s.jsonError(w, "server_error", "Failed to duplicate attachment", http.StatusInternalServerError)
		return
	}

	// 3. Clear latest flag for this group and create new version
	if err := s.DB.ClearLatestFlag(ctx, oldExam.GroupID); err != nil {
		s.Log.Printf("DB ClearLatestFlag error: %v", err)
		s.jsonError(w, "database_error", "Failed to update version status", http.StatusInternalServerError)
		return
	}

	modID := payload.ModuleID
	var commentPtr *string
	if payload.Comment != "" {
		commentPtr = &payload.Comment
	}

	newID := uuid.NewString()
	params := database.CreateExamParams{
		ID:          newID,
		Userid:      user.ID,
		Programid:   payload.ProgramID,
		Version:     payload.Version,
		Moduleid:    &modID,
		ExamDate:    payload.Date,
		Checksum:    oldExam.Checksum,
		Nbytes:      oldExam.Nbytes,
		MimeType:    oldExam.MimeType,
		Accesskey:   newAccessKey,
		Comment:     commentPtr,
		GroupID:     oldExam.GroupID,
		EditVersion: oldExam.EditVersion + 1,
		IsLatest:    1,
	}

	updated, err := s.DB.CreateExam(ctx, params)
	if err != nil {
		s.Log.Printf("DB CreateExam (version) error: %v", err)
		s.jsonError(w, "database_error", "Failed to create new version", http.StatusInternalServerError)
		return
	}

	var uModID int64
	if updated.Moduleid != nil {
		uModID = *updated.Moduleid
	}
	var uComm string
	if updated.Comment != nil {
		uComm = *updated.Comment
	}

	uploaded, _ := time.Parse(time.RFC3339, updated.UploadedAt)

	entry := ExamResponse{
		ID:           updated.ID,
		ProgramID:    updated.Programid,
		Version:      updated.Version,
		ModuleID:     uModID,
		ExamDate:     updated.ExamDate,
		UploadedAt:   uploaded,
		UploaderName: user.Name, // New record is by current user
		Comment:      uComm,
		EditVersion:  updated.EditVersion,
		GroupID:      updated.GroupID,
		IsLatest:     updated.IsLatest,
	}

	s.respondJSON(w, http.StatusOK, entry)
}

// DeleteExamsId deletes an exam
// @Summary Delete exam
// @Tags Exams
// @ID deleteExamsId
// @Param id path string true "Exam ID"
// @Success 204
// @Router /exams/{id} [delete]
func (s *Server) DeleteExamsId(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if user.Role != "admin" && user.Role != "editor" {
		s.jsonError(w, "forbidden", "Insufficient permissions", http.StatusForbidden)
		return
	}

	exam, err := s.DB.GetExam(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.jsonError(w, "not_found", "Exam not found", http.StatusNotFound)
		} else {
			s.Log.Printf("DB GetExam error: %v", err)
			s.jsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	if err := s.DB.DeleteExam(r.Context(), id); err != nil {
		s.Log.Printf("DB DeleteExam error: %v", err)
		s.jsonError(w, "database_error", "Failed to delete exam", http.StatusInternalServerError)
		return
	}

	objectKey := fmt.Sprintf("exams/%s.pdf", exam.Accesskey)
	if err := s.Store.Delete(r.Context(), objectKey); err != nil {
		s.Log.Printf("S3 Delete error (non-fatal): %v", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetExamsFile downloads file
// @Summary Download file
// @Tags Exams
// @ID getExamsFile
// @Param id path string true "Exam ID"
// @Success 200 {file} file
// @Router /exams/{id}/file [get]
func (s *Server) GetExamsFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, _, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	exam, err := s.DB.GetExam(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.jsonError(w, "not_found", "Exam not found", http.StatusNotFound)
		} else {
			s.Log.Printf("DB GetExam error: %v", err)
			s.jsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	objectKey := fmt.Sprintf("exams/%s.pdf", exam.Accesskey)
	obj, err := s.Store.GetObject(r.Context(), objectKey)
	if err != nil {
		s.Log.Printf("S3 GetObject error: %v", err)
		s.jsonError(w, "server_error", "Failed to retrieve file", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		s.Log.Printf("S3 Stat error: %v", err)
		s.jsonError(w, "not_found", "File content missing", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"exam_%s.pdf\"", exam.ExamDate))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))

	if _, err := io.Copy(w, obj); err != nil {
		s.Log.Printf("Stream error: %v", err)
	}
}
