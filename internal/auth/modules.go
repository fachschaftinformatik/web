package auth

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// @Description A study module
type ModuleResponse struct {
	ID        int    `json:"id"`
	ProgramID int    `json:"programid"`
	Name      string `json:"name"`
}

// GetProgramModules lists modules
// @Summary List modules for a program
// @Tags Programs
// @ID getProgramModules
// @Param id path int true "Program ID"
// @Success 200 {array} ModuleResponse
// @Failure 500 {object} ErrorResponse
// @Router /programs/{id}/modules [get]
func (s *Server) GetProgramModules(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	modules, err := s.DB.ListModulesByProgram(r.Context(), int64(id))
	if err != nil {
		s.Log.Printf("Failed to list modules: %v", err)
		s.jsonError(w, "database_error", "Could not fetch modules", http.StatusInternalServerError)
		return
	}

	apiModules := make([]ModuleResponse, 0, len(modules))
	for _, m := range modules {
		apiModules = append(apiModules, ModuleResponse{
			ID:        int(m.ID),
			ProgramID: int(m.Programid),
			Name:      m.Name,
		})
	}

	s.respondJSON(w, http.StatusOK, apiModules)
}
