package repository

import (
	"context"
	"fmt"
	"gotrade/order-service/internal/domain"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresOrderRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOrderRepository(db *pgxpool.Pool) OrderRepository {
	return &postgresOrderRepository{db: db}
}

func (r *postgresOrderRepository) CreateOrderWithOutbox(ctx context.Context, order *domain.Order, event *domain.OutboxEvent) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	orderQuery := `
        INSERT INTO orders (user_id, instrument_id, type, side, price, quantity, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at
    `
	err = tx.QueryRow(ctx, orderQuery,
		order.UserID,
		order.InstrumentID,
		order.Type,
		order.Side,
		order.Price,
		order.Quantity,
		order.Status,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return err
	}

	outboxQuery := `
        INSERT INTO outbox_events (id, aggregate_id, aggregate_type, event_type, payload)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err = tx.Exec(ctx, outboxQuery,
		event.ID,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.Payload,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *postgresOrderRepository) UpdateOrderStatus(ctx context.Context, orderID int64, userID int64, status domain.OrderStatus) (*domain.Order, error) {
	query := `
        UPDATE orders
        SET status = $1, updated_at = NOW()
        WHERE id = $2 AND user_id = $3 AND status = 'OPEN'
        RETURNING id, user_id, instrument_id, type, side, price, quantity, status, created_at, updated_at
    `
	var order domain.Order
	err := r.db.QueryRow(ctx, query, status, orderID, userID).Scan(
		&order.ID, &order.UserID, &order.InstrumentID, &order.Type, &order.Side,
		&order.Price, &order.Quantity, &order.Status, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Or a specific "not found or not updatable" error
		}
		return nil, err
	}
	return &order, nil
}

func (r *postgresOrderRepository) GetOpenOrdersByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	query := `
        SELECT id, user_id, instrument_id, type, side, price, quantity, status, created_at, updated_at
        FROM orders
        WHERE user_id = $1 AND status = 'OPEN'
        ORDER BY created_at DESC
    `
	return r.queryOrders(ctx, query, userID)
}

func (r *postgresOrderRepository) GetOrderHistoryByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	query := `
        SELECT id, user_id, instrument_id, type, side, price, quantity, status, created_at, updated_at
        FROM orders
        WHERE user_id = $1 AND status != 'OPEN'
        ORDER BY updated_at DESC
    `
	return r.queryOrders(ctx, query, userID)
}

func (r *postgresOrderRepository) queryOrders(ctx context.Context, query string, args ...interface{}) ([]*domain.Order, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*domain.Order, 0)
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.InstrumentID, &o.Type, &o.Side, &o.Price, &o.Quantity, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, nil
}

func (r *postgresOrderRepository) GetPollingOutboxEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	query := `
        SELECT id, aggregate_id, aggregate_type, event_type, payload, created_at
        FROM outbox_events
        ORDER BY created_at
        LIMIT $1
        FOR UPDATE SKIP LOCKED
    `
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.OutboxEvent
	for rows.Next() {
		var e domain.OutboxEvent
		if err := rows.Scan(&e.ID, &e.AggregateID, &e.AggregateType, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *postgresOrderRepository) DeleteOutboxEvents(ctx context.Context, tx pgx.Tx, eventIDs []string) error {
	if len(eventIDs) == 0 {
		return nil
	}

	// Convert []string to []uuid.UUID for the driver
	uuids := make([]uuid.UUID, len(eventIDs))
	for i, id := range eventIDs {
		parsedUUID, err := uuid.Parse(id)
		if err != nil {
			return fmt.Errorf("invalid UUID string: %s", id)
		}
		uuids[i] = parsedUUID
	}

	// Create placeholders like $1, $2, $3
	placeholders := make([]string, len(uuids))
	for i := range uuids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf("DELETE FROM outbox_events WHERE id IN (%s)", strings.Join(placeholders, ","))

	// Convert uuids to []interface{} for Exec
	args := make([]interface{}, len(uuids))
	for i, u := range uuids {
		args[i] = u
	}

	_, err := tx.Exec(ctx, query, args...)
	return err
}

