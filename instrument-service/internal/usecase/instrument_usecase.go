package usecase

import (
	"context"
	"gotrade/instrument-service/internal/domain"
	"gotrade/instrument-service/internal/repository"
)

type InstrumentUsecase interface {
	GetAllInstruments(ctx context.Context) ([]*domain.Instrument, error)
}

type instrumentUsecase struct {
	repo repository.InstrumentRepository
}

func NewInstrumentUsecase(repo repository.InstrumentRepository) InstrumentUsecase {
	return &instrumentUsecase{repo: repo}
}

func (uc *instrumentUsecase) GetAllInstruments(ctx context.Context) ([]*domain.Instrument, error) {
	return uc.repo.GetAll(ctx)
}

