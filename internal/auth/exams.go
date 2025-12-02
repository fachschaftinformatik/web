package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"errors"

	"github.com/fachschaftinformatik/web/internal/api"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const maxUploadSize = 10 << 20 

func toPtr[T any](v T) *T {
	return &v
}

type ExamAssignment struct {
	ProgramID int64  `json:"programid"`
	Version   string `json:"version"`
	ModuleID  int64  `json:"moduleid"`
}

func (s *Server) PostExams(w http.ResponseWriter, r *http.Request, params api.PostExamsParams) {
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
		s.jsonError(w, "bad_request", "Invalid assignments", http.StatusBadRequest)
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
	
	for _, assign := range assignments {
		paramsDB := database.CreateExamParams{
			ID:        uuid.NewString(),
			Userid:    user.ID,
			Programid: assign.ProgramID,
			Version:   assign.Version,
			Moduleid:  sql.NullInt64{Int64: assign.ModuleID, Valid: true},
			ExamDate:  examDate,
			Accesskey: accessKey,
			MimeType:  "application/pdf",
			Nbytes:    header.Size,
			Checksum:  checksum,
			Comment:   sql.NullString{String: comment, Valid: comment != ""},
		}

		_, err = s.DB.CreateExam(ctx, paramsDB)
		if err != nil {
			s.Log.Printf("DB CreateExam failed for assignment %+v: %v", assign, err)
		}
	}

	s.respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (s *Server) GetExams(w http.ResponseWriter, r *http.Request, params api.GetExamsParams) {
	_, _, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	dbParams := database.ListExamsParams{}
	if params.Programid != nil {
		dbParams.Programid = sql.NullInt64{Int64: int64(*params.Programid), Valid: true}
	}
	if params.Version != nil {
		dbParams.Version = sql.NullString{String: *params.Version, Valid: true}
	}
	if params.Moduleid != nil {
		dbParams.Moduleid = sql.NullInt64{Int64: int64(*params.Moduleid), Valid: true}
	}

	rows, err := s.DB.ListExams(r.Context(), dbParams)
	if err != nil {
		s.Log.Printf("Failed to list exams: %v", err)
		s.jsonError(w, "database_error", "Could not fetch exams", http.StatusInternalServerError)
		return
	}

	apiExams := make([]api.ExamListEntry, 0, len(rows))
	for _, row := range rows {
		entry := api.ExamListEntry{
			Id:           toPtr(row.ID),
			Programid:    toPtr(int(row.Programid)),
			Version:      toPtr(row.Version),
			Moduleid:     toPtr(int(row.Moduleid.Int64)),
			ModuleName:   toPtr(row.ModuleName),
			UploaderName: toPtr(row.UploaderName),
			Comment:      toPtr(row.Comment.String),
		}
		
		if t, err := time.Parse("2006-01-02", row.ExamDate); err == nil {
			entry.ExamDate = toPtr(openapi_types.Date{Time: t})
		}
		if t, err := time.Parse(time.RFC3339, row.UploadedAt); err == nil {
			entry.UploadedAt = toPtr(t)
		}

		apiExams = append(apiExams, entry)
	}

	s.respondJSON(w, http.StatusOK, apiExams)
}

func (s *Server) GetExamsFile(w http.ResponseWriter, r *http.Request, id string) {
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

func (s *Server) PutExamsId(w http.ResponseWriter, r *http.Request, id string) {
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if user.Role != "admin" && user.Role != "editor" {
		s.jsonError(w, "forbidden", "Insufficient permissions", http.StatusForbidden)
		return
	}

	var payload struct {
		Programid int    `json:"programid"`
		Version   string `json:"version"`
		Moduleid  int    `json:"moduleid"`
		Date      string `json:"date"`
		Comment   string `json:"comment"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.jsonError(w, "bad_request", "Invalid JSON body", http.StatusBadRequest)
		return
	}

	params := database.UpdateExamParams{
		ID:        id,
		Programid: int64(payload.Programid),
		Version:   payload.Version,
		Moduleid:  sql.NullInt64{Int64: int64(payload.Moduleid), Valid: true},
		ExamDate:  payload.Date,
		Comment:   sql.NullString{String: payload.Comment, Valid: payload.Comment != ""},
	}

	updated, err := s.DB.UpdateExam(r.Context(), params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.jsonError(w, "not_found", "Exam not found", http.StatusNotFound)
		} else {
			s.Log.Printf("DB UpdateExam error: %v", err)
			s.jsonError(w, "database_error", "Failed to update exam", http.StatusInternalServerError)
		}
		return
	}
	
	entry := api.ExamListEntry{
		Id:           toPtr(updated.ID),
		Programid:    toPtr(int(updated.Programid)),
		Version:      toPtr(updated.Version),
		Moduleid:     toPtr(int(updated.Moduleid.Int64)),
		Comment:      toPtr(updated.Comment.String),
	}
	
	if t, err := time.Parse("2006-01-02", updated.ExamDate); err == nil {
		entry.ExamDate = toPtr(openapi_types.Date{Time: t})
	}

	s.respondJSON(w, http.StatusOK, entry)
}

func (s *Server) DeleteExamsId(w http.ResponseWriter, r *http.Request, id string) {
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if user.Role != "admin" && user.Role != "editor" {
		s.jsonError(w, "forbidden", "Insufficient permissions", http.StatusForbidden)
		return
	}

	// Fetch to get accesskey for S3 deletion
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

	// Delete from DB
	if err := s.DB.DeleteExam(r.Context(), id); err != nil {
		s.Log.Printf("DB DeleteExam error: %v", err)
		s.jsonError(w, "database_error", "Failed to delete exam", http.StatusInternalServerError)
		return
	}

	// Delete from S3
	objectKey := fmt.Sprintf("exams/%s.pdf", exam.Accesskey)
	if err := s.Store.Delete(r.Context(), objectKey); err != nil {
		s.Log.Printf("S3 Delete error (non-fatal): %v", err)
	}

	w.WriteHeader(http.StatusNoContent)
}
