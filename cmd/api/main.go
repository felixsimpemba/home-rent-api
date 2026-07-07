package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/felixsimpemba/home-rent-api/internal/config"
	"github.com/felixsimpemba/home-rent-api/internal/database"
	"github.com/felixsimpemba/home-rent-api/internal/router"
	"github.com/felixsimpemba/home-rent-api/internal/ws"
)

func main() {
	// 1. Load configuration from .env
	cfg := config.LoadConfig()

	// 2. Initialize database connection and run migrations
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Ensure DB connection is closed on exit
	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	// 3. Create the WebSocket hub (singleton for the lifetime of the server)
	hub := ws.NewHub()

	// 4. Register all routes with DB and hub injected into handlers
	handler := router.RegisterRoutes(cfg, db, hub)

	// 5. Start HTTP + WebSocket server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	// 6. Graceful shutdown on SIGINT / SIGTERM
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutdown signal received. Stopping server gracefully...")
		if err := srv.Close(); err != nil {
			log.Printf("Server close error: %v", err)
		}
	}()

	log.Printf("✓ Server running on http://localhost:%s", cfg.Port)
	log.Printf("✓ REST API:   http://localhost:%s/api/v1", cfg.Port)
	log.Printf("✓ WebSocket:  ws://localhost:%s/api/v1/ws/connect  (requires Authorization header)", cfg.Port)
	log.Printf("✓ Press Ctrl+C to stop.")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("Server stopped.")
}
