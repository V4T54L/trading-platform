package repository

import (
	"context"
	"gotrade/account-service/internal/domain"

	"github.com/jackc/pgx/v5"
)

type AccountRepository interface {
	GetByUserID(ctx context.Context, userID int64) (*domain.Account, error)
	Create(ctx context.Context, account *domain.Account) error
	UpdateBalanceInTx(ctx context.Context, tx pgx.Tx, accountID int64, newBalance float64) error
	CreateTransactionInTx(ctx context.Context, tx pgx.Tx, transaction *domain.Transaction) error
	RunWithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error
}

