package auth

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// @Description A study program including valid PO versions
type ProgramResponse struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

// GetPrograms lists all programs
// @Summary List all programs
// @Description Returns a list of study programs and their valid PO versions
// @Tags Programs
// @ID getPrograms
// @Accept json
// @Produce json
// @Success 200 {array} ProgramResponse
// @Failure 500 {object} ErrorResponse
// @Router /programs [get]
func (s *Server) GetPrograms(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.ListProgramsWithVersions(r.Context())
	if err != nil {
		s.Log.Printf("Failed to list programs: %v", err)
		s.jsonError(w, "database_error", "Could not fetch programs", http.StatusInternalServerError)
		return
	}

	programMap := make(map[int64]*ProgramResponse)
	var orderedPrograms []*ProgramResponse

	for _, row := range rows {
		prog, exists := programMap[row.ID]
		if !exists {
			prog = &ProgramResponse{
				ID:       int(row.ID),
				Name:     row.Name,
				Versions: []string{},
			}
			programMap[row.ID] = prog
			orderedPrograms = append(orderedPrograms, prog)
		}
		prog.Versions = append(prog.Versions, row.Version)
	}

	response := make([]ProgramResponse, 0, len(orderedPrograms))
	for _, p := range orderedPrograms {
		response = append(response, *p)
	}

	s.respondJSON(w, http.StatusOK, response)
}

// GetProgramsId gets a single program
// @Summary Get program by ID
// @Tags Programs
// @ID getProgramsId
// @Param id path int true "Program ID"
// @Success 200 {object} ProgramResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /programs/{id} [get]
func (s *Server) GetProgramsId(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	rows, err := s.DB.GetProgramWithVersions(r.Context(), int64(id))
	if err != nil {
		s.Log.Printf("Failed to get program: %v", err)
		s.jsonError(w, "database_error", "Could not fetch program", http.StatusInternalServerError)
		return
	}

	if len(rows) == 0 {
		s.jsonError(w, "not_found", "Program not found", http.StatusNotFound)
		return
	}

	prog := ProgramResponse{
		ID:       int(rows[0].ID),
		Name:     rows[0].Name,
		Versions: make([]string, 0, len(rows)),
	}

	for _, row := range rows {
		prog.Versions = append(prog.Versions, row.Version)
	}

	s.respondJSON(w, http.StatusOK, prog)
}
