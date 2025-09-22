package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"gotrade/order-service/internal/domain"
	"gotrade/order-service/internal/repository"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderUsecase interface {
	PlaceOrder(ctx context.Context, userID int64, instrumentID int64, orderType domain.OrderType, side domain.OrderSide, price float64, quantity int) (*domain.Order, error)
	CancelOrder(ctx context.Context, userID int64, orderID int64) (*domain.Order, error)
	GetAllOrders(ctx context.Context, userID int64) ([]*domain.Order, error)
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
	finalPayload, _ := json.Marshal(order) // Re-marshal with updated ID

	// Best-effort immediate publish. Outbox poller will ensure delivery.
	go uc.publisher.Publish(context.Background(), "orders", finalPayload)

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

	// For cancellations, we can publish directly as it's less critical than order creation
	// and doesn't require the same level of transactional integrity for this MVP.
	go uc.publisher.Publish(context.Background(), "orders", payload)

	return order, nil
}

func (uc *orderUsecase) GetAllOrders(ctx context.Context, userID int64) ([]*domain.Order, error) {
	openOrders, err := uc.repo.GetOpenOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	historyOrders, err := uc.repo.GetOrderHistoryByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	allOrders := append(openOrders, historyOrders...)
	// Sort by creation time descending to be consistent
	sort.Slice(allOrders, func(i, j int) bool {
		return allOrders[i].CreatedAt.After(allOrders[j].CreatedAt)
	})
	return allOrders, nil
}

func (uc *orderUsecase) StartOutboxPolling(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("Starting outbox poller...")

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
		log.Printf("Error starting transaction for outbox polling: %v", err)
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

	log.Printf("Found %d events in outbox to publish", len(events))

	eventIDsToDelete := make([]string, 0, len(events))
	for _, event := range events {
		err := uc.publisher.Publish(ctx, "orders", event.Payload)
		if err != nil {
			log.Printf("Error publishing event %s: %v. Will retry.", event.ID, err)
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

