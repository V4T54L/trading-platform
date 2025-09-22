package repository

import (
	"context"
	"gotrade/instrument-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresInstrumentRepository struct {
	db *pgxpool.Pool
}

func NewPostgresInstrumentRepository(db *pgxpool.Pool) InstrumentRepository {
	return &postgresInstrumentRepository{db: db}
}

func (r *postgresInstrumentRepository) Create(ctx context.Context, instrument *domain.Instrument) error {
	query := `INSERT INTO instruments (symbol, name, type, exchange, tick_size, lot_size)
			  VALUES ($1, $2, $3, $4, $5, $6)
			  RETURNING id, created_at`
	err := r.db.QueryRow(ctx, query,
		instrument.Symbol,
		instrument.Name,
		instrument.Type,
		instrument.Exchange,
		instrument.TickSize,
		instrument.LotSize,
	).Scan(&instrument.ID, &instrument.CreatedAt)
	return err
}

func (r *postgresInstrumentRepository) GetAll(ctx context.Context) ([]*domain.Instrument, error) {
	query := `SELECT id, symbol, name, type, exchange, tick_size, lot_size, created_at FROM instruments ORDER BY symbol`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	instruments := make([]*domain.Instrument, 0)
	for rows.Next() {
		var i domain.Instrument
		if err := rows.Scan(&i.ID, &i.Symbol, &i.Name, &i.Type, &i.Exchange, &i.TickSize, &i.LotSize, &i.CreatedAt); err != nil {
			return nil, err
		}
		instruments = append(instruments, &i)
	}

	return instruments, nil
}

