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
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS instruments (
        id BIGSERIAL PRIMARY KEY,
        symbol VARCHAR(50) NOT NULL UNIQUE,
        name VARCHAR(255) NOT NULL,
        type VARCHAR(20) NOT NULL,
        exchange VARCHAR(50) NOT NULL,
        tick_size NUMERIC(10, 4) NOT NULL,
        lot_size INTEGER NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    `
	_, err := pool.Exec(context.Background(), createTableSQL)
	return err
}

