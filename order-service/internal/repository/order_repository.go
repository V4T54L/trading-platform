package repository

import (
	"context"
	"gotrade/order-service/internal/domain"

	"github.com/jackc/pgx/v5"
)

type OrderRepository interface {
	CreateOrderWithOutbox(ctx context.Context, order *domain.Order, event *domain.OutboxEvent) error
	UpdateOrderStatus(ctx context.Context, orderID int64, userID int64, status domain.OrderStatus) (*domain.Order, error)
	GetOpenOrdersByUserID(ctx context.Context, userID int64) ([]*domain.Order, error)
	GetOrderHistoryByUserID(ctx context.Context, userID int64) ([]*domain.Order, error)
	GetPollingOutboxEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	DeleteOutboxEvents(ctx context.Context, tx pgx.Tx, eventIDs []string) error
}

