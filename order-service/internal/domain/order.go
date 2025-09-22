package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderType string
type OrderSide string
type OrderStatus string

const (
	Market OrderType = "MARKET"
	Limit  OrderType = "LIMIT"
)

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

const (
	Open      OrderStatus = "OPEN"
	Completed OrderStatus = "COMPLETED"
	Cancelled OrderStatus = "CANCELLED"
	Rejected  OrderStatus = "REJECTED"
)

type Order struct {
	ID           int64       `json:"id"`
	UserID       int64       `json:"user_id"`
	InstrumentID int64       `json:"instrument_id"`
	Type         OrderType   `json:"type"`
	Side         OrderSide   `json:"side"`
	Price        float64     `json:"price"` // 0 for market orders
	Quantity     int         `json:"quantity"`
	Status       OrderStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type OutboxEvent struct {
	ID            uuid.UUID
	AggregateID   string
	AggregateType string
	EventType     string
	Payload       []byte
	CreatedAt     time.Time
}

