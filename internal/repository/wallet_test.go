package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWalletRepository_Deposit_SQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWalletRepository(db)
	uid := uuid.New()
	amount := 100.0

	mock.ExpectExec("INSERT INTO wallets").
		WithArgs(uid, amount).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Deposit(context.Background(), uid, amount)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWalletRepository_Withdraw_SQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWalletRepository(db)
	uid := uuid.New()
	amount := 50.0

	mock.ExpectExec("UPDATE wallets").
		WithArgs(amount, uid).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Withdraw(context.Background(), uid, amount)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
