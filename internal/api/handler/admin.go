package handler

import (
	"encoding/json"
	"net/http"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/id"
)

// @Summary Create a program
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body dto.ProgramResponse true "Program Data"
// @Success 201 {object} dto.ProgramResponse
// @Router /admin/programs [post]
func (s *Server) PostAdminPrograms(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name     string   `json:"name" validate:"required"`
		Versions []string `json:"versions" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "bad_request", "Invalid JSON", http.StatusBadRequest)
		return
	}

	prog, err := s.DB.CreateProgram(r.Context(), database.CreateProgramParams{
		ID:   int64(id.New()),
		Name: payload.Name,
	})
	if err != nil {
		s.JsonError(w, "database_error", err.Error(), http.StatusInternalServerError)
		return
	}

	for _, v := range payload.Versions {
		_ = s.DB.CreateProgramVersion(r.Context(), database.CreateProgramVersionParams{
			ProgramID: prog.ID,
			Name:      v,
		})
	}

	s.RespondJSON(w, http.StatusCreated, dto.ProgramResponse{
		ID:       id.ID(prog.ID),
		Name:     prog.Name,
		Versions: payload.Versions,
	})
}

// @Summary Create a module
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body dto.ModuleResponse true "Module Data"
// @Success 201 {object} dto.ModuleResponse
// @Router /admin/modules [post]
func (s *Server) PostAdminModules(w http.ResponseWriter, r *http.Request) {
	var payload dto.ModuleResponse
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "bad_request", "Invalid JSON", http.StatusBadRequest)
		return
	}

	mod, err := s.DB.CreateModule(r.Context(), database.CreateModuleParams{
		ID:        int64(id.New()),
		ProgramID: int64(payload.ProgramID),
		Name:      payload.Name,
		Alias:     &payload.Alias,
	})
	if err != nil {
		s.JsonError(w, "database_error", err.Error(), http.StatusInternalServerError)
		return
	}

	s.RespondJSON(w, http.StatusCreated, dto.ModuleResponse{
		ID:        id.ID(mod.ID),
		ProgramID: id.ID(mod.ProgramID),
		Name:      mod.Name,
		Alias:     *mod.Alias,
	})
}

// @Summary Get all user roles
// @Tags Admin
// @ID getAdminRefRoles
// @Success 200 {array} string
// @Router /admin/ref/roles [get]
func (s *Server) GetAdminRefRoles(w http.ResponseWriter, r *http.Request) {
	// For now, hardcoded or fetch from DB if I add a ListRefRoles query
	// Since I squashed them into 0001, I should add queries for them if needed.
	// But let's just return the known ones for now.
	roles := []string{"admin", "editor", "user"}
	s.RespondJSON(w, http.StatusOK, roles)
}

// @Summary Get all discussion types
// @Tags Admin
// @ID getAdminRefDiscussionTypes
// @Success 200 {array} string
// @Router /admin/ref/discussion-types [get]
func (s *Server) GetAdminRefDiscussionTypes(w http.ResponseWriter, r *http.Request) {
	types := []string{"discussion", "news", "event"}
	s.RespondJSON(w, http.StatusOK, types)
}

// Add more as needed
