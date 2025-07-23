package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration from .env
type Config struct {
	Port             string
	DBHost           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBPort           string
	JWTAccessSecret  string
	JWTRefreshSecret string
}

// LoadEnv loads environment variables from .env file
func LoadEnv() error {
	// Find .env file in the project root or current directory
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found: %v", err)
		// Continue execution even if .env is not found
		// Variables might be set in the environment directly
	}
	return nil
}

// GetConfig returns a Config struct populated with values from environment
func GetConfig() Config {
	return Config{
		Port:             GetEnv("PORT", "8080"),
		DBHost:           GetEnv("DB_HOST", "localhost"),
		DBUser:           GetEnv("DB_USER", "postgres"),
		DBPassword:       GetEnv("DB_PASSWORD", ""),
		DBName:           GetEnv("DB_NAME", "seta"),
		DBPort:           GetEnv("DB_PORT", "5432"),
		JWTAccessSecret:  GetEnv("JWT_ACCESS_SECRET", ""),
		JWTRefreshSecret: GetEnv("JWT_REFRESH_SECRET", ""),
	}
}

// GetEnv returns an environment variable or a default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
