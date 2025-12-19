package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"wallet-service/internal/model"
	"wallet-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Service
type MockService struct {
	mock.Mock
}

func (m *MockService) UpdateWallet(ctx context.Context, req model.WalletRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockService) GetBalance(ctx context.Context, walletID uuid.UUID) (float64, error) {
	args := m.Called(ctx, walletID)
	return args.Get(0).(float64), args.Error(1)
}

// Tests

func TestUpdateWallet_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockService)
	h := NewWalletHandler(mockSvc)
	uid := uuid.New()
	reqModel := model.WalletRequest{
		WalletID:      uid,
		OperationType: model.Deposit,
		Amount:        1000,
	}

	// Ожидаем вызов сервиса без ошибки
	mockSvc.On("UpdateWallet", mock.Anything, reqModel).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	reqBody, _ := json.Marshal(map[string]interface{}{
		"valletId":      uid,
		"operationType": "DEPOSIT",
		"amount":        1000,
	})
	c.Request, _ = http.NewRequest("POST", "/api/v1/wallet", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateWallet(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateWallet_BadRequest_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockService)
	h := NewWalletHandler(mockSvc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Отрицательная сумма (не пройдет валидацию gt=0)
	reqBody := []byte(`{
		"valletId": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		"operationType": "DEPOSIT",
		"amount": -100
	}`)
	c.Request, _ = http.NewRequest("POST", "/api/v1/wallet", bytes.NewBuffer(reqBody))

	h.UpdateWallet(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetBalance_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockService)
	h := NewWalletHandler(mockSvc)
	uid := uuid.New()

	mockSvc.On("GetBalance", mock.Anything, uid).Return(0.0, repository.ErrWalletNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "WALLET_UUID", Value: uid.String()}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/wallets/"+uid.String(), nil)

	h.GetBalance(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
