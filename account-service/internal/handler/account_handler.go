package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"gotrade/account-service/internal/domain"
	"gotrade/account-service/internal/usecase"
)

type AccountHandler struct {
	accountUsecase usecase.AccountUsecase
}

func NewAccountHandler(uc usecase.AccountUsecase) *AccountHandler {
	return &AccountHandler{accountUsecase: uc}
}

type FundTransferRequest struct {
	Type   string  `json:"type"` // DEPOSIT or WITHDRAWAL
	Amount float64 `json:"amount"`
}

// GetAccount godoc
// @Summary Get user account details
// @Description Retrieves the financial account details for the authenticated user.
// @Tags account
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} usecase.AccountInfo
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Account not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router / [get]
func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	account, err := h.accountUsecase.GetAccountByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, usecase.ErrAccountNotFound) {
			http.Error(w, `{"error":"Account not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

// FundTransfer godoc
// @Summary Deposit or withdraw funds
// @Description Handles depositing or withdrawing funds from the user's account.
// @Tags account
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   transfer body FundTransferRequest true "Fund Transfer Request"
// @Success 200 {object} usecase.AccountInfo
// @Failure 400 {object} map[string]string "Invalid request body or insufficient funds"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /transfer [post]
func (h *AccountHandler) FundTransfer(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req FundTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, `{"error":"Amount must be positive"}`, http.StatusBadRequest)
		return
	}

	var updatedAccount *usecase.AccountInfo
	var err error

	switch domain.TransactionType(req.Type) {
	case domain.Deposit:
		updatedAccount, err = h.accountUsecase.Deposit(r.Context(), userID, req.Amount)
	case domain.Withdrawal:
		updatedAccount, err = h.accountUsecase.Withdraw(r.Context(), userID, req.Amount)
	default:
		http.Error(w, `{"error":"Invalid transaction type"}`, http.StatusBadRequest)
		return
	}

	if err != nil {
		if errors.Is(err, usecase.ErrInsufficientFunds) {
			http.Error(w, `{"error":"Insufficient funds"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedAccount)
}

