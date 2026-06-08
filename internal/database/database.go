package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/felixsimpemba/home-rent-api/internal/config"
)

// InitDB initializes database connection based on config
func InitDB(cfg *config.Config) (*sql.DB, error) {
	// DSN format: username:password@tcp(host:port)/database?parseTime=true
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBDatabase,
	)

	log.Printf("Connecting to %s database on %s:%s (database: %s)...", cfg.DBConnection, cfg.DBHost, cfg.DBPort, cfg.DBDatabase)

	// Note: In your real implementation, import the driver anonymously in main.go:
	// import _ "github.com/go-sql-driver/mysql"
	// Then call sql.Open("mysql", dsn)
	
	// We use standard driver open stub for initialization compilation safety
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	return db, nil
}
