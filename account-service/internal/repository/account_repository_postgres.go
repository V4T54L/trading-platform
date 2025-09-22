package repository

import (
	"context"
	"errors"
	"gotrade/account-service/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresAccountRepository struct {
	db *pgxpool.Pool
}

func NewPostgresAccountRepository(db *pgxpool.Pool) AccountRepository {
	return &postgresAccountRepository{db: db}
}

func (r *postgresAccountRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Account, error) {
	query := `SELECT id, user_id, balance, currency, created_at, updated_at FROM accounts WHERE user_id = $1`
	var acc domain.Account
	err := r.db.QueryRow(ctx, query, userID).Scan(&acc.ID, &acc.UserID, &acc.Balance, &acc.Currency, &acc.CreatedAt, &acc.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found is not an error here
		}
		return nil, err
	}
	return &acc, nil
}

func (r *postgresAccountRepository) Create(ctx context.Context, account *domain.Account) error {
	query := `INSERT INTO accounts (user_id, balance, currency) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query, account.UserID, account.Balance, account.Currency).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
}

func (r *postgresAccountRepository) UpdateBalanceInTx(ctx context.Context, tx pgx.Tx, accountID int64, newBalance float64) error {
	query := `UPDATE accounts SET balance = $1, updated_at = $2 WHERE id = $3`
	_, err := tx.Exec(ctx, query, newBalance, time.Now(), accountID)
	return err
}

func (r *postgresAccountRepository) CreateTransactionInTx(ctx context.Context, tx pgx.Tx, transaction *domain.Transaction) error {
	query := `INSERT INTO transactions (account_id, type, amount) VALUES ($1, $2, $3)`
	_, err := tx.Exec(ctx, query, transaction.AccountID, transaction.Type, transaction.Amount)
	return err
}

func (r *postgresAccountRepository) RunWithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

