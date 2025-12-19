package main

import (
	"database/sql"
	"log"
	"time"
	"wallet-service/internal/config"
	"wallet-service/internal/handler"
	"wallet-service/internal/repository"
	"wallet-service/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	var db *sql.DB
	for i := 0; i < 10; i++ {
		db, err = sql.Open("pgx", cfg.DBUrl)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Connected to DB!")
				break
			}
		}
		log.Printf("DB not ready, waiting... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping database:", err)
	}

	walletRepo := repository.NewWalletRepository(db)
	walletService := service.NewWalletService(walletRepo)
	walletHandler := handler.NewWalletHandler(walletService)

	err = runMigrations(db)
	if err != nil {
		log.Fatal("failed to run migrations:", err)
	}

	router := gin.Default()
	api := router.Group("/api/v1")
	{
		api.POST("/wallet", walletHandler.UpdateWallet)
		api.GET("/wallets/:WALLET_UUID", walletHandler.GetBalance)
	}

	log.Printf("Listening on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("failed to start server:", err)
	}
}

func runMigrations(db *sql.DB) error {
	query := `
			CREATE TABLE IF NOT EXISTS wallets (
					id UUID PRIMARY KEY,
					balance DECIMAL(20, 2) NOT NULL DEFAULT 0
			);
	`
	_, err := db.Exec(query)
	return err
}
