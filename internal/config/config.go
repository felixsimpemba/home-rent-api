package config

import (
	"bufio"
	"os"
	"strings"
)

// Config represents the application configuration
type Config struct {
	Port           string
	DBConnection   string
	DBHost         string
	DBPort         string
	DBDatabase     string
	DBUsername     string
	DBPassword     string
	JWTSecret      string
	MtnMomoSecret  string
	AirtelMomoKey  string
	ZamtelMomoKey  string
}

// LoadConfig reads configuration variables from environment or returns defaults
func LoadConfig() *Config {
	// Attempt to load .env file if present
	loadEnvFile(".env")

	return &Config{
		Port:           getEnv("PORT", "8080"),
		DBConnection:   getEnv("DB_CONNECTION", "mysql"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "3306"),
		DBDatabase:     getEnv("DB_DATABASE", "home-rent"),
		DBUsername:     getEnv("DB_USERNAME", "root"),
		DBPassword:     getEnv("DB_PASSWORD", ""),
		JWTSecret:      getEnv("JWT_SECRET", "super-secret-key-change-in-production"),
		MtnMomoSecret:  getEnv("MTN_MOMO_SECRET", "mock-mtn-secret"),
		AirtelMomoKey:  getEnv("AIRTEL_MOMO_KEY", "mock-airtel-key"),
		ZamtelMomoKey:  getEnv("ZAMTEL_MOMO_KEY", "mock-zamtel-key"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// loadEnvFile reads a .env file and sets the environment variables
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // Ignore if the file doesn't exist (e.g. in prod, environments are loaded natively)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Strip optional quotes
			value = strings.Trim(value, `"'`)
			
			// Set environment variable if it's not already set
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		// Standard error check after loop finishes
		_ = err
	}
}
