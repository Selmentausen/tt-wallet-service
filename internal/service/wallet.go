package service

import (
	"context"
	"fmt"
	"wallet-service/internal/model"

	"github.com/google/uuid"
)

type Service interface {
	UpdateWallet(ctx context.Context, req model.WalletRequest) error
	GetBalance(ctx context.Context, walletID uuid.UUID) (float64, error)
}

type WalletRepository interface {
	Deposit(ctx context.Context, walletID uuid.UUID, amount float64) error
	Withdraw(ctx context.Context, walletID uuid.UUID, amount float64) error
	GetBalance(ctx context.Context, walletID uuid.UUID) (float64, error)
}

type WalletService struct {
	repo WalletRepository
}

func NewWalletService(repo WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) UpdateWallet(ctx context.Context, req model.WalletRequest) error {
	switch req.OperationType {
	case model.Deposit:
		return s.repo.Deposit(ctx, req.WalletID, req.Amount)
	case model.Withdraw:
		return s.repo.Withdraw(ctx, req.WalletID, req.Amount)
	default:
		return fmt.Errorf("unknown operation type: %s", req.OperationType)
	}
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (float64, error) {
	return s.repo.GetBalance(ctx, walletID)
}
