package repository

import (
	"context"
	"gotrade/instrument-service/internal/domain"
)

type InstrumentRepository interface {
	GetAll(ctx context.Context) ([]*domain.Instrument, error)
	Create(ctx context.Context, instrument *domain.Instrument) error
}

