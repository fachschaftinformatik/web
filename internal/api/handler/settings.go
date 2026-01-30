package handler

import (
	"encoding/json"
	"net/http"

	"github.com/fachschaftinformatik/web/internal/api/dto"
)

// @Summary Get office occupancy status
// @Tags Settings
// @Produce json
// @Success 200 {object} dto.OfficeStatusResponse
// @Router /office-status [get]
func (s *Server) GetOfficeStatus(w http.ResponseWriter, r *http.Request) {
	config, err := s.DB.GetConfig(r.Context())
	if err != nil {
		s.RespondJSON(w, http.StatusOK, dto.OfficeStatusResponse{Occupied: false})
		return
	}

	s.RespondJSON(w, http.StatusOK, dto.OfficeStatusResponse{Occupied: config.OfficeOccupied == 1})
}

// @Summary Update office occupancy status
// @Tags Settings
// @Accept json
// @Produce json
// @Param request body dto.UpdateOfficeStatusRequest true "Office Status"
// @Success 200 {object} dto.OfficeStatusResponse
// @Router /office-status [put]
func (s *Server) PutOfficeStatus(w http.ResponseWriter, r *http.Request) {
	var payload dto.UpdateOfficeStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	err := s.DB.UpdateOfficeOccupied(r.Context(), s.BoolToInt(payload.Occupied))
	if err != nil {
		s.JsonError(w, "database_error", "Could not update office status", http.StatusInternalServerError)
		return
	}

	s.RespondJSON(w, http.StatusOK, dto.OfficeStatusResponse{Occupied: payload.Occupied})
}
