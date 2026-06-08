package main

import (
	"log"
	"net/http"

	"github.com/felixsimpemba/home-rent-api/internal/config"
	"github.com/felixsimpemba/home-rent-api/internal/database"
	"github.com/felixsimpemba/home-rent-api/internal/router"
)

func main() {
	// 1. Load configurations
	cfg := config.LoadConfig()

	// 2. Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 3. Initialize route configurations
	handler := router.RegisterRoutes(cfg)

	// 4. Launch HTTP Server
	log.Printf("Server is starting on port %s...", cfg.Port)
	log.Printf("Base API endpoints at: http://localhost:%s/api/v1", cfg.Port)
	log.Printf("Press Ctrl+C to terminate.")

	err = http.ListenAndServe(":"+cfg.Port, handler)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
