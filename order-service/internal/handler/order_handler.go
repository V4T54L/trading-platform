package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gotrade/order-service/internal/domain"
	"gotrade/order-service/internal/usecase"
)

type OrderHandler struct {
	orderUsecase usecase.OrderUsecase
}

func NewOrderHandler(uc usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{orderUsecase: uc}
}

type PlaceOrderRequest struct {
	InstrumentID int64             `json:"instrument_id"`
	Type         domain.OrderType  `json:"type"`
	Side         domain.OrderSide  `json:"side"`
	Price        float64           `json:"price"`
	Quantity     int               `json:"quantity"`
}

// PlaceOrder godoc
// @Summary Place a new order
// @Description Creates a new trading order for the authenticated user.
// @Tags orders
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   order body PlaceOrderRequest true "Order Placement Info"
// @Success 201 {object} domain.Order
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router / [post]
func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Quantity <= 0 {
		http.Error(w, `{"error":"Quantity must be positive"}`, http.StatusBadRequest)
		return
	}
	if req.Type == domain.Limit && req.Price <= 0 {
		http.Error(w, `{"error":"Price must be positive for LIMIT orders"}`, http.StatusBadRequest)
		return
	}

	order, err := h.orderUsecase.PlaceOrder(r.Context(), userID, req.InstrumentID, req.Type, req.Side, req.Price, req.Quantity)
	if err != nil {
		http.Error(w, `{"error":"Failed to place order"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Cancels an existing open order for the authenticated user.
// @Tags orders
// @Produce  json
// @Security BearerAuth
// @Param   orderID path int true "Order ID"
// @Success 200 {object} domain.Order
// @Failure 400 {object} map[string]string "Invalid order ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Order not found or not cancellable"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /{orderID} [delete]
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	orderIDStr := chi.URLParam(r, "orderID")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid order ID"}`, http.StatusBadRequest)
		return
	}

	order, err := h.orderUsecase.CancelOrder(r.Context(), userID, orderID)
	if err != nil {
		http.Error(w, `{"error":"Failed to cancel order"}`, http.StatusInternalServerError)
		return
	}

	if order == nil {
		http.Error(w, `{"error":"Order not found or cannot be cancelled"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

// GetOrders godoc
// @Summary Get all user orders
// @Description Retrieves a list of all open and historical orders for the authenticated user.
// @Tags orders
// @Produce  json
// @Security BearerAuth
// @Success 200 {array} domain.Order
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router / [get]
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	orders, err := h.orderUsecase.GetAllOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

