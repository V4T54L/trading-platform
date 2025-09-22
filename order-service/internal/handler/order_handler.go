package handler

import (
	"encoding/json"
	"errors"
	"gotrade/order-service/internal/domain"
	"gotrade/order-service/internal/usecase"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	uc usecase.OrderUsecase
}

func NewOrderHandler(uc usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

type PlaceOrderRequest struct {
	InstrumentID int64            `json:"instrument_id"`
	Type         domain.OrderType `json:"type"`
	Side         domain.OrderSide `json:"side"`
	Price        float64          `json:"price"`
	Quantity     int              `json:"quantity"`
}

func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Quantity <= 0 {
		http.Error(w, "Quantity must be positive", http.StatusBadRequest)
		return
	}
	if req.Type == domain.Limit && req.Price <= 0 {
		http.Error(w, "Price must be positive for limit orders", http.StatusBadRequest)
		return
	}

	order, err := h.uc.PlaceOrder(r.Context(), userID, req.InstrumentID, req.Type, req.Side, req.Price, req.Quantity)
	if err != nil {
		http.Error(w, "Failed to place order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orderIDStr := chi.URLParam(r, "orderID")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.uc.CancelOrder(r.Context(), userID, orderID)
	if err != nil {
		if errors.Is(err, errors.New("order not found or cannot be cancelled")) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to cancel order", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

