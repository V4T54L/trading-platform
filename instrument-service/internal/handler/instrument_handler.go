package handler

import (
	"encoding/json"
	"net/http"

	"gotrade/instrument-service/internal/usecase"
)

type InstrumentHandler struct {
	instrumentUsecase usecase.InstrumentUsecase
}

func NewInstrumentHandler(uc usecase.InstrumentUsecase) *InstrumentHandler {
	return &InstrumentHandler{instrumentUsecase: uc}
}

// GetAllInstruments godoc
// @Summary Get all instruments
// @Description Retrieves a list of all tradable instruments.
// @Tags instruments
// @Produce  json
// @Success 200 {array} domain.Instrument
// @Failure 500 {object} map[string]string "Internal server error"
// @Router / [get]
func (h *InstrumentHandler) GetAllInstruments(w http.ResponseWriter, r *http.Request) {
	instruments, err := h.instrumentUsecase.GetAllInstruments(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(instruments)
}

