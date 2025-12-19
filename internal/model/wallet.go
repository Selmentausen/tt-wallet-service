package model

import "github.com/google/uuid"

type OperationType string

const (
	Deposit  OperationType = "DEPOSIT"
	Withdraw OperationType = "WITHDRAW"
)

type WalletRequest struct {
	WalletID      uuid.UUID     `json:"valletId" binding:"required"`
	OperationType OperationType `json:"operationType" binding:"required,oneof=DEPOSIT WITHDRAW"`
	Amount        float64       `json:"amount" binding:"required,gt=0"`
}

type WalletResponse struct {
	WalletID uuid.UUID `json:"Id"`
	Balance  float64   `json:"balance"`
}
