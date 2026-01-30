package handler

import (
	"encoding/json"
	"net/http"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/database"
)

// @Summary Get office occupancy status
// @Tags Settings
// @Produce json
// @Success 200 {object} dto.OfficeStatusResponse
// @Router /office-status [get]
func (s *Server) GetOfficeStatus(w http.ResponseWriter, r *http.Request) {
	val, err := s.DB.GetSetting(r.Context(), "office_occupied")
	if err != nil {
		s.RespondJSON(w, http.StatusOK, dto.OfficeStatusResponse{Occupied: false})
		return
	}

	s.RespondJSON(w, http.StatusOK, dto.OfficeStatusResponse{Occupied: val == "true"})
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

	val := "false"
	if payload.Occupied {
		val = "true"
	}

	err := s.DB.UpdateSetting(r.Context(), database.UpdateSettingParams{
		Key:   "office_occupied",
		Value: val,
	})
	if err != nil {
		s.JsonError(w, "database_error", "Could not update office status", http.StatusInternalServerError)
		return
	}

	s.RespondJSON(w, http.StatusOK, dto.OfficeStatusResponse{Occupied: payload.Occupied})
}
