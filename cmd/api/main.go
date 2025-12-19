package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")

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
