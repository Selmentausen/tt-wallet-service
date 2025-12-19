package api

import (
	"database/sql"
	"fmt"
	"log"
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

	db, err := sql.Open("pgx", cfg.DBUrl)
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

	log.Printf("Listening on port %d", cfg.Port)
	if err := router.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatal("failed to start server:", err)
	}
}

func runMigrations(db *sql.DB) error {
	query := `
			CREATE TABLE IF NOT EXISTS wallet (
					id UUID PRIMARY KEY,
					balance DECIMAL(20, 2) NOT NULL DEFAULT 0
			);
	`
	_, err := db.Exec(query)
	return err
}
