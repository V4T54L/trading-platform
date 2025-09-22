package domain

import "time"

type TransactionType string

const (
	Deposit    TransactionType = "DEPOSIT"
	Withdrawal TransactionType = "WITHDRAWAL"
)

type Account struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID        int64           `json:"id"`
	AccountID int64           `json:"account_id"`
	Type      TransactionType `json:"type"`
	Amount    float64         `json:"amount"`
	Timestamp time.Time       `json:"timestamp"`
}

