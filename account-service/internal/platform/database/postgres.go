package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool creates a new PostgreSQL connection pool.
func NewPostgresPool(connStr string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	// Setting default pool configurations
	config.MaxConns = 10
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 1 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	// Ping the database to verify the connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

// InitDatabase initializes the database schema.
func InitDatabase(pool *pgxpool.Pool) error {
	schema := `
    CREATE TABLE IF NOT EXISTS accounts (
        id BIGSERIAL PRIMARY KEY,
        user_id BIGINT NOT NULL UNIQUE,
        balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
        currency VARCHAR(10) NOT NULL DEFAULT 'USD',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

    CREATE TABLE IF NOT EXISTS transactions (
        id BIGSERIAL PRIMARY KEY,
        account_id BIGINT NOT NULL REFERENCES accounts(id),
        type VARCHAR(20) NOT NULL,
        amount NUMERIC(15, 2) NOT NULL,
        timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    `
	_, err := pool.Exec(context.Background(), schema)
	return err
}

