package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(connStr string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 10
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 1 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pool, nil
}

func InitDatabase(pool *pgxpool.Pool) error {
	createOrdersTableSQL := `
    CREATE TABLE IF NOT EXISTS orders (
        id BIGSERIAL PRIMARY KEY,
        user_id BIGINT NOT NULL,
        instrument_id BIGINT NOT NULL,
        type VARCHAR(20) NOT NULL,
        side VARCHAR(10) NOT NULL,
        price NUMERIC(12, 4) NOT NULL,
        quantity INTEGER NOT NULL,
        status VARCHAR(20) NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    `
	if _, err := pool.Exec(context.Background(), createOrdersTableSQL); err != nil {
		return err
	}

	createOutboxTableSQL := `
    CREATE TABLE IF NOT EXISTS outbox_events (
        id UUID PRIMARY KEY,
        aggregate_id VARCHAR(255) NOT NULL,
        aggregate_type VARCHAR(255) NOT NULL,
        event_type VARCHAR(255) NOT NULL,
        payload JSONB NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    `
	_, err := pool.Exec(context.Background(), createOutboxTableSQL)
	return err
}

