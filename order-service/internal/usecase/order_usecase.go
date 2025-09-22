package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"gotrade/order-service/internal/domain"
	"gotrade/order-service/internal/repository"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderUsecase interface {
	PlaceOrder(ctx context.Context, userID int64, instrumentID int64, orderType domain.OrderType, side domain.OrderSide, price float64, quantity int) (*domain.Order, error)
	CancelOrder(ctx context.Context, userID int64, orderID int64) (*domain.Order, error)
	StartOutboxPolling(ctx context.Context)
}

type MessagePublisher interface {
	Publish(ctx context.Context, channel string, message []byte) error
}

type orderUsecase struct {
	repo      repository.OrderRepository
	publisher MessagePublisher
	dbPool    *pgxpool.Pool
}

func NewOrderUsecase(repo repository.OrderRepository, publisher MessagePublisher, dbPool *pgxpool.Pool) OrderUsecase {
	return &orderUsecase{
		repo:      repo,
		publisher: publisher,
		dbPool:    dbPool,
	}
}

func (uc *orderUsecase) PlaceOrder(ctx context.Context, userID int64, instrumentID int64, orderType domain.OrderType, side domain.OrderSide, price float64, quantity int) (*domain.Order, error) {
	order := &domain.Order{
		UserID:       userID,
		InstrumentID: instrumentID,
		Type:         orderType,
		Side:         side,
		Price:        price,
		Quantity:     quantity,
		Status:       domain.Open,
	}

	payload, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order payload: %w", err)
	}

	event := &domain.OutboxEvent{
		ID:            uuid.New(),
		AggregateID:   strconv.FormatInt(order.ID, 10), // Will be 0 initially, but that's okay
		AggregateType: "Order",
		EventType:     "OrderPlaced",
		Payload:       payload,
	}

	err = uc.repo.CreateOrderWithOutbox(ctx, order, event)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Update event with the actual order ID
	event.AggregateID = strconv.FormatInt(order.ID, 10)
	payload, _ = json.Marshal(order)
	event.Payload = payload

	// This is a bit of a hack. The event in the DB will have the correct payload but AggregateID will be 0.
	// A better approach would be to update the event in the same transaction, but that complicates the repo.
	// For now, we publish the correct event immediately. The poller will handle the DB record.
	go uc.publisher.Publish(context.Background(), "order_events", payload)

	return order, nil
}

func (uc *orderUsecase) CancelOrder(ctx context.Context, userID int64, orderID int64) (*domain.Order, error) {
	// For MVP, we just update the status. A real system would have more complex logic.
	order, err := uc.repo.UpdateOrderStatus(ctx, orderID, userID, domain.Cancelled)
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("order not found or cannot be cancelled")
	}

	payload, err := json.Marshal(order)
	if err != nil {
		log.Printf("Failed to marshal cancel order payload: %v", err)
		return order, nil // Return success even if publishing fails
	}

	go uc.publisher.Publish(context.Background(), "order_events", payload)

	return order, nil
}

func (uc *orderUsecase) StartOutboxPolling(ctx context.Context) {
	log.Println("Starting outbox poller...")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping outbox poller...")
			return
		case <-ticker.C:
			uc.pollAndPublish(ctx)
		}
	}
}

func (uc *orderUsecase) pollAndPublish(ctx context.Context) {
	tx, err := uc.dbPool.Begin(ctx)
	if err != nil {
		log.Printf("Error beginning transaction for outbox polling: %v", err)
		return
	}
	defer tx.Rollback(ctx)

	events, err := uc.repo.GetPollingOutboxEvents(ctx, 10)
	if err != nil {
		log.Printf("Error polling outbox events: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	log.Printf("Polled %d events from outbox", len(events))

	eventIDsToDelete := make([]string, 0, len(events))
	for _, event := range events {
		err := uc.publisher.Publish(ctx, "order_events", event.Payload)
		if err != nil {
			log.Printf("Failed to publish outbox event %s: %v. Will retry.", event.ID, err)
			// Don't add to delete list, so it will be retried
			continue
		}
		eventIDsToDelete = append(eventIDsToDelete, event.ID.String())
	}

	if len(eventIDsToDelete) > 0 {
		if err := uc.repo.DeleteOutboxEvents(ctx, tx, eventIDsToDelete); err != nil {
			log.Printf("Error deleting processed outbox events: %v", err)
			return // Rollback will happen
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("Error committing transaction for outbox polling: %v", err)
	}
}

