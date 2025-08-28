package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration from .env
type Config struct {
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
	return Config{}
}

// GetEnv returns an environment variable or a default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
