//go:build integration

package repository

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/assert"
)

// Подключение к Docker-базе (внешний порт 5436)
const TestDBUrl = "postgres://postgres:secret@localhost:5436/wallet_db?sslmode=disable"

func setupRealDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", TestDBUrl)
	if err != nil {
		t.Skip("Skipping integration test: connection failed")
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		t.Skip("Skipping integration test: db not ready")
	}
	return db
}

func TestIntegration_Concurrency_1000RPS(t *testing.T) {
	db := setupRealDB(t)
	defer db.Close()
	repo := NewWalletRepository(db)
	ctx := context.Background()

	id := uuid.New()

	// Очистка тестовых данных
	defer db.Exec("DELETE FROM wallets WHERE id = $1", id)

	// Создаем кошелек с 0
	err := repo.Deposit(ctx, id, 0)
	assert.NoError(t, err)

	workers := 100
	requestsPerWorker := 10

	var wg sync.WaitGroup
	wg.Add(workers)

	errChan := make(chan error, workers*requestsPerWorker)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				// ВАЖНО: Не игнорируем ошибку!
				if err := repo.Deposit(ctx, id, 1); err != nil {
					errChan <- err // Отправляем ошибку в канал
				}
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Читаем ошибки
	errorCount := 0
	for err := range errChan {
		errorCount++
		// Выводим первую ошибку, чтобы понять суть
		if errorCount == 1 {
			t.Logf("First error encountered: %v", err)
		}
	}

	if errorCount > 0 {
		t.Logf("Total failed requests: %d", errorCount)
	}

	bal, err := repo.GetBalance(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, 1000.0, bal, "Lost update detected!")
}
