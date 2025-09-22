package handler

import (
	"encoding/json"
	"gotrade/instrument-service/internal/usecase"
	"net/http"
)

type InstrumentHandler struct {
	uc usecase.InstrumentUsecase
}

func NewInstrumentHandler(uc usecase.InstrumentUsecase) *InstrumentHandler {
	return &InstrumentHandler{uc: uc}
}

func (h *InstrumentHandler) GetAllInstruments(w http.ResponseWriter, r *http.Request) {
	instruments, err := h.uc.GetAllInstruments(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve instruments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instruments)
}

