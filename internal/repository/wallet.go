package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	ErrWalletNotFound = errors.New("wallet not found")
	ErrInsuffFunds    = errors.New("insufficient funds")
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Deposit(ctx context.Context, walletID uuid.UUID, amount float64) error {
	query := `
		INSERT INTO wallets (id, balance) 
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE 
		SET balance = wallets.balance + $2
	`
	_, err := r.db.ExecContext(ctx, query, walletID, amount)
	if err != nil {
		return fmt.Errorf("failed to deposit: %w", err)
	}
	return nil
}

func (r *WalletRepository) Withdraw(ctx context.Context, walletID uuid.UUID, amount float64) error {
	query := `
			UPDATE wallets
			SET balance = balance - $1
			WHERE id = $2 AND balance >= $1
`
	res, err := r.db.ExecContext(ctx, query, walletID, amount)
	if err != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	// Проверяем сколько строк изменилось, если 0 строк - значит операция не прошла (либо нет денег, либо нет кошелька)
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		var exists bool
		err = r.db.QueryRowContext(ctx, query, walletID, amount).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			return ErrWalletNotFound
		}
		return ErrInsuffFunds
	}
	return nil
}

func (r *WalletRepository) GetBalance(ctx context.Context, walletID uuid.UUID) (float64, error) {
	var balance float64
	query := "SELECT balance FROM wallets WHERE id = $1;"
	err := r.db.QueryRowContext(ctx, query, walletID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrWalletNotFound
		}
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}
