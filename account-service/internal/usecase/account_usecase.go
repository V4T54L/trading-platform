package usecase

import (
	"context"
	"errors"
	"gotrade/account-service/internal/domain"
	"gotrade/account-service/internal/repository"

	"github.com/jackc/pgx/v5"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAccountNotFound   = errors.New("account not found")
)

type AccountInfo struct {
	UserID   int64   `json:"user_id"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

type AccountUsecase interface {
	GetAccountByUserID(ctx context.Context, userID int64) (*AccountInfo, error)
	Deposit(ctx context.Context, userID int64, amount float64) (*AccountInfo, error)
	Withdraw(ctx context.Context, userID int64, amount float64) (*AccountInfo, error)
}

type accountUsecase struct {
	repo repository.AccountRepository
}

func NewAccountUsecase(repo repository.AccountRepository) AccountUsecase {
	return &accountUsecase{repo: repo}
}

func (uc *accountUsecase) getOrCreateAccount(ctx context.Context, userID int64) (*domain.Account, error) {
	account, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		// Create a new account if it doesn't exist
		newAccount := &domain.Account{
			UserID:   userID,
			Balance:  100000.00, // Start with mock funds
			Currency: "USD",
		}
		if err := uc.repo.Create(ctx, newAccount); err != nil {
			return nil, err
		}
		return newAccount, nil
	}
	return account, nil
}

func (uc *accountUsecase) GetAccountByUserID(ctx context.Context, userID int64) (*AccountInfo, error) {
	account, err := uc.getOrCreateAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &AccountInfo{
		UserID:   account.UserID,
		Balance:  account.Balance,
		Currency: account.Currency,
	}, nil
}

func (uc *accountUsecase) Deposit(ctx context.Context, userID int64, amount float64) (*AccountInfo, error) {
	var finalAccount *domain.Account
	err := uc.repo.RunWithTransaction(ctx, func(tx pgx.Tx) error {
		account, err := uc.repo.GetByUserID(ctx, userID)
		if err != nil {
			return err
		}
		if account == nil {
			return ErrAccountNotFound
		}

		newBalance := account.Balance + amount
		if err := uc.repo.UpdateBalanceInTx(ctx, tx, account.ID, newBalance); err != nil {
			return err
		}

		transaction := &domain.Transaction{
			AccountID: account.ID,
			Type:      domain.Deposit,
			Amount:    amount,
		}
		if err := uc.repo.CreateTransactionInTx(ctx, tx, transaction); err != nil {
			return err
		}
		account.Balance = newBalance
		finalAccount = account
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &AccountInfo{
		UserID:   finalAccount.UserID,
		Balance:  finalAccount.Balance,
		Currency: finalAccount.Currency,
	}, nil
}

func (uc *accountUsecase) Withdraw(ctx context.Context, userID int64, amount float64) (*AccountInfo, error) {
	var finalAccount *domain.Account
	err := uc.repo.RunWithTransaction(ctx, func(tx pgx.Tx) error {
		account, err := uc.repo.GetByUserID(ctx, userID)
		if err != nil {
			return err
		}
		if account == nil {
			return ErrAccountNotFound
		}

		if account.Balance < amount {
			return ErrInsufficientFunds
		}

		newBalance := account.Balance - amount
		if err := uc.repo.UpdateBalanceInTx(ctx, tx, account.ID, newBalance); err != nil {
			return err
		}

		transaction := &domain.Transaction{
			AccountID: account.ID,
			Type:      domain.Withdrawal,
			Amount:    amount,
		}
		if err := uc.repo.CreateTransactionInTx(ctx, tx, transaction); err != nil {
			return err
		}
		account.Balance = newBalance
		finalAccount = account
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &AccountInfo{
		UserID:   finalAccount.UserID,
		Balance:  finalAccount.Balance,
		Currency: finalAccount.Currency,
	}, nil
}

