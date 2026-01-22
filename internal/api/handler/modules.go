package handler

import (
	"net/http"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/id"
	"github.com/go-chi/chi/v5"
)

// @Summary List modules for a program
// @Tags Programs
// @ID getProgramModules
// @Param programId path string true "Program ID"
// @Success 200 {array} dto.ModuleResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /programs/{programId}/modules [get]
func (s *Server) GetProgramModules(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "programId")
	pid, err := id.Parse(idStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid program ID", http.StatusBadRequest)
		return
	}

	modules, err := s.DB.ListModulesByProgram(r.Context(), int64(pid))
	if err != nil {
		s.Log.Error("Failed to list modules", "err", err)
		s.JsonError(w, "database_error", "Could not fetch modules", http.StatusInternalServerError)
		return
	}

	apiModules := make([]dto.ModuleResponse, 0, len(modules))
	for _, m := range modules {
		alias := ""
		if m.Alias != nil {
			alias = *m.Alias
		}
		apiModules = append(apiModules, dto.ModuleResponse{
			ID:        id.ID(m.ID),
			ProgramID: id.ID(m.ProgramID),
			Name:      m.Name,
			Alias:     alias,
		})
	}

	s.RespondJSON(w, http.StatusOK, apiModules)
}
