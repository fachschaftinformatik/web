package auth

import (
	"net/http"

	"github.com/fachschaftinformatik/web/internal/api"
)

func (s *Server) GetProgramModules(w http.ResponseWriter, r *http.Request, id int) {
	modules, err := s.DB.ListModulesByProgram(r.Context(), int64(id))
	if err != nil {
		s.Log.Printf("Failed to list modules: %v", err)
		s.jsonError(w, "database_error", "Could not fetch modules", http.StatusInternalServerError)
		return
	}

	apiModules := make([]api.Module, 0, len(modules))
	for _, m := range modules {
		apiModules = append(apiModules, api.Module{
			Id:        int(m.ID),
			Programid: int(m.Programid),
			Name:      m.Name,
		})
	}

	s.respondJSON(w, http.StatusOK, apiModules)
}
