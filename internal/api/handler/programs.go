package handler

import (
	"net/http"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/id"
	"github.com/go-chi/chi/v5"
)

// @Summary List all programs
// @Description Returns a list of study programs and their valid PO versions
// @Tags Programs
// @ID getPrograms
// @Accept json
// @Produce json
// @Success 200 {array} dto.ProgramResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /programs [get]
func (s *Server) GetPrograms(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.ListProgramsWithVersions(r.Context())
	if err != nil {
		s.Log.Error("Failed to list programs", "err", err)
		s.JsonError(w, "database_error", "Could not fetch programs", http.StatusInternalServerError)
		return
	}

	programMap := make(map[int64]*dto.ProgramResponse)
	var orderedIDs []int64

	for _, row := range rows {
		prog, exists := programMap[row.ID]
		if !exists {
			prog = &dto.ProgramResponse{
				ID:       id.ID(row.ID),
				Name:     row.Name,
				Versions: []string{},
			}
			programMap[row.ID] = prog
			orderedIDs = append(orderedIDs, row.ID)
		}
		prog.Versions = append(prog.Versions, row.Version)
	}

	response := make([]dto.ProgramResponse, 0, len(orderedIDs))
	for _, pid := range orderedIDs {
		response = append(response, *programMap[pid])
	}

	s.RespondJSON(w, http.StatusOK, response)
}

// @Summary Get program by ID
// @Tags Programs
// @ID getProgramsId
// @Param programId path string true "Program ID"
// @Success 200 {object} dto.ProgramResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /programs/{programId} [get]
func (s *Server) GetProgramsId(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "programId")
	pid, err := id.Parse(idStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid program ID", http.StatusBadRequest)
		return
	}

	rows, err := s.DB.GetProgramWithVersions(r.Context(), int64(pid))
	if err != nil {
		s.Log.Error("Failed to get program", "err", err)
		s.JsonError(w, "database_error", "Could not fetch program", http.StatusInternalServerError)
		return
	}

	if len(rows) == 0 {
		s.JsonError(w, "not_found", "Program not found", http.StatusNotFound)
		return
	}

	prog := dto.ProgramResponse{
		ID:       id.ID(rows[0].ID),
		Name:     rows[0].Name,
		Versions: make([]string, 0, len(rows)),
	}

	for _, row := range rows {
		prog.Versions = append(prog.Versions, row.Version)
	}

	s.RespondJSON(w, http.StatusOK, prog)
}
