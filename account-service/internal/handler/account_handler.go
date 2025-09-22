package handler

import (
	"encoding/json"
	"errors"
	"gotrade/account-service/internal/usecase"
	"net/http"
)

type AccountHandler struct {
	accountUsecase usecase.AccountUsecase
}

func NewAccountHandler(uc usecase.AccountUsecase) *AccountHandler {
	return &AccountHandler{accountUsecase: uc}
}

type FundTransferRequest struct {
	Type   string  `json:"type"` // "DEPOSIT" or "WITHDRAWAL"
	Amount float64 `json:"amount"`
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	account, err := h.accountUsecase.GetAccountByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get account", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func (h *AccountHandler) FundTransfer(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req FundTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	var account *usecase.AccountInfo
	var err error

	switch req.Type {
	case "DEPOSIT":
		account, err = h.accountUsecase.Deposit(r.Context(), userID, req.Amount)
	case "WITHDRAWAL":
		account, err = h.accountUsecase.Withdraw(r.Context(), userID, req.Amount)
	default:
		http.Error(w, "Invalid transfer type", http.StatusBadRequest)
		return
	}

	if err != nil {
		if errors.Is(err, usecase.ErrInsufficientFunds) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Fund transfer failed", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

