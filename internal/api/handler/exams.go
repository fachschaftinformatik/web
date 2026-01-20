package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fachschaftinformatik/web/internal/id"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/go-chi/chi/v5"
)

const maxUploadSize = 256 << 20 // 256 MB

// @Summary List exam versions
// @Tags Exams
// @ID getExamVersions
// @Param groupId path string true "Exam Group ID"
// @Success 200 {array} dto.ExamResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /exams/versions/{groupId} [get]
func (s *Server) GetExamVersions(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")

	rows, err := s.DB.ListExamVersions(r.Context(), groupID)
	if err != nil {
		s.Log.Error("Failed to list exam versions", "err", err)
		s.JsonError(w, "database_error", "Could not fetch exam history", http.StatusInternalServerError)
		return
	}

	apiExams := make([]dto.ExamResponse, 0, len(rows))
	for _, row := range rows {
		var modID string
		if row.Moduleid != nil {
			modID = *row.Moduleid
		}
		var comm string
		if row.Comment != nil {
			comm = *row.Comment
		}

		apiExams = append(apiExams, dto.ExamResponse{
			ID:           row.ID,
			ProgramID:    row.Programid,
			Version:      row.Version,
			ModuleID:     modID,
			ModuleName:   row.ModuleName,
			ExamDate:     row.ExamDate,
			UploadedAt:   row.UploadedAt,
			UploaderName: row.UploaderName,
			Comment:      comm,
			EditVersion:  row.EditVersion,
			GroupID:      row.GroupID,
			IsLatest:     row.IsLatest,
		})
	}

	s.RespondJSON(w, http.StatusOK, apiExams)
}

// @Summary Get exam details
// @Tags Exams
// @ID getExamsId
// @Param id path string true "Exam ID"
// @Success 200 {object} dto.ExamResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /exams/{id} [get]
func (s *Server) GetExamsId(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	row, err := s.DB.GetExamDetails(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.JsonError(w, "not_found", "Exam not found", http.StatusNotFound)
		} else {
			s.Log.Error("DB GetExamDetails error", "err", err)
			s.JsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	var modID string
	if row.Moduleid != nil {
		modID = *row.Moduleid
	}
	var comm string
	if row.Comment != nil {
		comm = *row.Comment
	}

	resp := dto.ExamResponse{
		ID:           row.ID,
		ProgramID:    row.Programid,
		Version:      row.Version,
		ModuleID:     modID,
		ModuleName:   row.ModuleName,
		ExamDate:     row.ExamDate,
		UploadedAt:   row.UploadedAt,
		UploaderName: row.UploaderName,
		Comment:      comm,
		EditVersion:  row.EditVersion,
		GroupID:      row.GroupID,
		IsLatest:     row.IsLatest,
	}

	s.RespondJSON(w, http.StatusOK, resp)
}

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
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /exams [post]
func (s *Server) PostExams(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+1024*1024)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		if errors.As(err, new(*http.MaxBytesError)) {
			s.JsonError(w, "bad_request", "File too large", http.StatusBadRequest)
		} else {
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

	if header.Size > maxUploadSize {
		s.JsonError(w, "bad_request", "File too large", http.StatusBadRequest)
		return
	}
	mimeType := header.Header.Get("Content-Type")
	if mimeType != "application/pdf" {
		s.JsonError(w, "bad_request", "Only PDF files are allowed", http.StatusBadRequest)
		return
	}

	examDate := r.FormValue("date")
	comment := r.FormValue("comment")
	assignmentsJSON := r.FormValue("assignments")

	if _, err := time.Parse("2006-01-02", examDate); err != nil {
		s.JsonError(w, "invalid_input", "Invalid exam date format (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	if assignmentsJSON == "" {
		s.JsonError(w, "bad_request", "Missing metadata", http.StatusBadRequest)
		return
	}

	var assignments []dto.ExamAssignment
	if err := json.Unmarshal([]byte(assignmentsJSON), &assignments); err != nil || len(assignments) == 0 {
		s.JsonError(w, "bad_request", "Invalid assignments JSON", http.StatusBadRequest)
		return
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		s.JsonError(w, "server_error", "Failed to process file", http.StatusInternalServerError)
		return
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))

	ctx := r.Context()
	var commentPtr *string
	if comment != "" {
		c := s.Sanitize(comment)
		commentPtr = &c
	}

	for _, assign := range assignments {
		if err := s.Validate(assign); err != nil {
			s.JsonError(w, "invalid_input", fmt.Sprintf("Invalid assignment: %v", err), http.StatusBadRequest)
			return
		}
		modID := assign.ModuleID
		newExamID := id.New()

		objectKey := fmt.Sprintf("exams/%s/1.pdf", newExamID)

		file.Seek(0, 0)
		if err := s.Store.Upload(r.Context(), objectKey, file, header.Size, "application/pdf"); err != nil {
			s.Log.Error("S3 Upload failed", "objectKey", objectKey, "err", err)
			continue
		}

		paramsDB := database.CreateExamParams{
			ID:          newExamID,
			Userid:      user.ID,
			Programid:   assign.ProgramID,
			Version:     assign.Version,
			Moduleid:    &modID,
			ExamDate:    examDate,
			Accesskey:   id.New(),
			MimeType:    "application/pdf",
			Nbytes:      header.Size,
			Checksum:    checksum,
			Comment:     commentPtr,
			GroupID:     newExamID,
			EditVersion: 1,
			IsLatest:    1,
		}

		_, err = s.DB.CreateExam(ctx, paramsDB)
		if err != nil {
			s.Log.Error("DB CreateExam failed", "assignment", assign, "err", err)
		} else {
			moduleName := "einem Modul"
			mod, err := s.DB.GetModule(ctx, modID)
			if err == nil {
				moduleName = mod.Name
			}

			msg := fmt.Sprintf("%s (%s)", moduleName, examDate)
			link := fmt.Sprintf("/exams/%s?mod=%s&examId=%s", modID, url.QueryEscape(moduleName), url.QueryEscape(newExamID))
			s.broadcastNotification(r, "Neue Klausur", msg, "exam", link)

			var targetName *string
			if moduleName != "" {
				targetName = &moduleName
			}
			_, _ = s.DB.CreateActivity(r.Context(), database.CreateActivityParams{
				ID:         id.New(),
				UserID:     user.ID,
				Type:       "EXAM_UPLOADED",
				TargetID:   newExamID,
				TargetName: targetName,
			})
		}
	}

	s.RespondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

// @Summary List exams
// @Tags Exams
// @ID getExams
// @Param programid query string false "Program ID"
// @Param version query string false "Version"
// @Param moduleid query string false "Module ID"
// @Success 200 {array} dto.ExamResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /exams [get]
func (s *Server) GetExams(w http.ResponseWriter, r *http.Request) {
	dbParams := database.ListExamsParams{}

	if pid := r.URL.Query().Get("programid"); pid != "" {
		dbParams.Programid = &pid
	}
	if ver := r.URL.Query().Get("version"); ver != "" {
		dbParams.Version = &ver
	}
	if mid := r.URL.Query().Get("moduleid"); mid != "" {
		dbParams.Moduleid = &mid
	}

	rows, err := s.DB.ListExams(r.Context(), dbParams)
	if err != nil {
		s.Log.Error("Failed to list exams", "err", err)
		s.JsonError(w, "database_error", "Could not fetch exams", http.StatusInternalServerError)
		return
	}

	apiExams := make([]dto.ExamResponse, 0, len(rows))
	for _, row := range rows {
		var modID string
		if row.Moduleid != nil {
			modID = *row.Moduleid
		}
		var comm string
		if row.Comment != nil {
			comm = *row.Comment
		}

		apiExams = append(apiExams, dto.ExamResponse{
			ID:           row.ID,
			ProgramID:    row.Programid,
			Version:      row.Version,
			ModuleID:     modID,
			ModuleName:   row.ModuleName,
			ExamDate:     row.ExamDate,
			UploadedAt:   row.UploadedAt,
			UploaderName: row.UploaderName,
			Comment:      comm,
			EditVersion:  row.EditVersion,
			GroupID:      row.GroupID,
			IsLatest:     row.IsLatest,
		})
	}

	s.RespondJSON(w, http.StatusOK, apiExams)
}

// @Summary Update exam
// @Tags Exams
// @ID putExamsId
// @Param id path string true "Exam ID"
// @Param body body dto.UpdateExamRequest true "Update Data"
// @Success 200 {object} dto.ExamResponse
// @Router /exams/{id} [put]
func (s *Server) PutExamsId(w http.ResponseWriter, r *http.Request) {
	examID := chi.URLParam(r, "id")
	user, _ := s.User(r)

	var payload dto.UpdateExamRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "bad_request", "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	oldExam, err := s.DB.GetExam(ctx, examID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.JsonError(w, "not_found", "Exam not found", http.StatusNotFound)
		} else {
			s.Log.Error("DB GetExam error", "err", err)
			s.JsonError(w, "database_error", "Failed to fetch exam", http.StatusInternalServerError)
		}
		return
	}

	newAccessKey := id.New()
	oldPath := fmt.Sprintf("exams/%s/%d.pdf", oldExam.GroupID, oldExam.EditVersion)
	newPath := fmt.Sprintf("exams/%s/%d.pdf", oldExam.GroupID, oldExam.EditVersion+1)

	if err := s.Store.CopyObject(ctx, oldPath, newPath); err != nil {
		s.Log.Error("Storage CopyObject error", "err", err)
		s.JsonError(w, "server_error", "Failed to duplicate attachment", http.StatusInternalServerError)
		return
	}

	if err := s.DB.ClearLatestFlag(ctx, oldExam.GroupID); err != nil {
		s.Log.Error("DB ClearLatestFlag error", "err", err)
		s.JsonError(w, "database_error", "Failed to update version status", http.StatusInternalServerError)
		return
	}

	modID := payload.ModuleID
	var commentPtr *string
	if payload.Comment != "" {
		c := s.Sanitize(payload.Comment)
		commentPtr = &c
	}

	newID := id.New()
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
		s.Log.Error("DB CreateExam (version) error", "err", err)
		s.JsonError(w, "database_error", "Failed to create new version", http.StatusInternalServerError)
		return
	}

	var uModID string
	if updated.Moduleid != nil {
		uModID = *updated.Moduleid
	}
	var uComm string
	if updated.Comment != nil {
		uComm = *updated.Comment
	}

	entry := dto.ExamResponse{
		ID:           updated.ID,
		ProgramID:    updated.Programid,
		Version:      updated.Version,
		ModuleID:     uModID,
		ExamDate:     updated.ExamDate,
		UploadedAt:   updated.UploadedAt,
		UploaderName: user.Name,
		Comment:      uComm,
		EditVersion:  updated.EditVersion,
		GroupID:      updated.GroupID,
		IsLatest:     updated.IsLatest,
	}

	s.RespondJSON(w, http.StatusOK, entry)
}

// @Summary Delete exam
// @Tags Exams
// @ID deleteExamsId
// @Param id path string true "Exam ID"
// @Success 204
// @Router /exams/{id} [delete]
func (s *Server) DeleteExamsId(w http.ResponseWriter, r *http.Request) {
	examID := chi.URLParam(r, "id")

	exam, err := s.DB.GetExam(r.Context(), examID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.JsonError(w, "not_found", "Exam not found", http.StatusNotFound)
		} else {
			s.Log.Error("DB GetExam error", "err", err)
			s.JsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	if err := s.DB.DeleteExam(r.Context(), examID); err != nil {
		s.Log.Error("DB DeleteExam error", "err", err)
		s.JsonError(w, "database_error", "Failed to delete exam", http.StatusInternalServerError)
		return
	}

	objectKey := fmt.Sprintf("exams/%s/%d.pdf", exam.GroupID, exam.EditVersion)
	if err := s.Store.Delete(r.Context(), objectKey); err != nil {
		s.Log.Error("S3 Delete error (non-fatal)", "err", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Download file
// @Tags Exams
// @ID getExamsFile
// @Param id path string true "Exam ID"
// @Success 200 {file} file
// @Router /exams/{id}/file [get]
func (s *Server) GetExamsFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	exam, err := s.DB.GetExam(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.JsonError(w, "not_found", "Exam not found", http.StatusNotFound)
		} else {
			s.Log.Error("DB GetExam error", "err", err)
			s.JsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	objectKey := fmt.Sprintf("exams/%s/%d.pdf", exam.GroupID, exam.EditVersion)
	obj, err := s.Store.GetObject(r.Context(), objectKey)
	if err != nil {
		s.Log.Error("S3 GetObject error", "err", err)
		s.JsonError(w, "server_error", "Failed to retrieve file", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		s.Log.Error("S3 Stat error", "err", err)
		s.JsonError(w, "not_found", "File content missing", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"exam_%s.pdf\"", exam.ExamDate))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))

	if _, err := io.Copy(w, obj); err != nil {
		s.Log.Error("Stream error", "err", err)
	}
}
