package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/felixsimpemba/home-rent-api/internal/config"
	"github.com/felixsimpemba/home-rent-api/internal/models"
)

// InitDB initializes the GORM database connection and runs schema migrations
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	// DSN: username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBDatabase,
	)

	log.Printf("Connecting to MySQL via GORM on %s:%s (database: %s)...", cfg.DBHost, cfg.DBPort, cfg.DBDatabase)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("error opening gorm database: %v", err)
	}

	// Run schema auto-migrations for all models
	log.Println("Running GORM database schema auto-migrations...")
	err = db.AutoMigrate(
		// Auth
		&models.User{},
		&models.Session{},
		&models.PasswordResetToken{},
		&models.EmailVerificationToken{},

		// Properties
		&models.Category{},
		&models.Property{},
		&models.PropertyImage{},
		&models.Amenity{},
		&models.AvailabilityBlock{},
		&models.PropertyHistoryLog{},
		&models.Report{},

		// Tenant activity
		&models.Favorite{},
		&models.SavedSearch{},

		// Inquiries & Leases
		&models.Inquiry{},
		&models.InquiryReply{},
		&models.Viewing{},
		&models.Booking{},
		&models.Lease{},
		&models.TenantScreening{},

		// Payments
		&models.Transaction{},
		&models.Refund{},
		&models.SubscriptionPlan{},
		&models.Subscription{},

		// Messaging
		&models.Conversation{},
		&models.Message{},
		&models.Notification{},

		// Uploads
		&models.UploadedFile{},

		// CRM
		&models.Lead{},
		&models.LeadNote{},
		&models.Task{},
		&models.Meeting{},

		// Admin
		&models.AuditLog{},
		&models.SystemSetting{},
	)
	if err != nil {
		return nil, fmt.Errorf("error during auto-migration: %v", err)
	}

	log.Println("Database schema migration completed successfully.")

	return db, nil
}
