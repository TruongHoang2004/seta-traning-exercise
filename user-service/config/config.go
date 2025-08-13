package config

import (
	"os"
	"user-service/pkg/logger"

	"github.com/joho/godotenv"
)

// Config holds all configuration from .env
type Config struct {
	Port                 string
	Production           bool
	DBHost               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBPort               string
	JWTAccessSecret      string
	JWTAccessExpiration  string
	JWTRefreshSecret     string
	JWTRefreshExpiration string
	LogFilePath          string
}

// LoadEnv loads environment variables from .env file
func LoadEnv() error {
	// Find .env file in the project root or current directory
	err := godotenv.Load()
	if err != nil {
		logger.Warn("Warning: .env file not found")
	}
	return nil
}

// GetConfig returns a Config struct populated with values from environment
func GetConfig() Config {
	return Config{
		Port:                 GetEnv("PORT", "8080"),
		Production:           GetEnv("GO_ENV", "development") == "production",
		DBHost:               GetEnv("DB_HOST", "localhost"),
		DBUser:               GetEnv("DB_USER", "postgres"),
		DBPassword:           GetEnv("DB_PASSWORD", ""),
		DBName:               GetEnv("DB_NAME", "seta"),
		DBPort:               GetEnv("DB_PORT", "5432"),
		JWTAccessSecret:      GetEnv("JWT_ACCESS_SECRET", ""),
		JWTAccessExpiration:  GetEnv("JWT_ACCESS_EXPIRATION", "15m"),
		JWTRefreshSecret:     GetEnv("JWT_REFRESH_SECRET", ""),
		JWTRefreshExpiration: GetEnv("JWT_REFRESH_EXPIRATION", "7d"),
		LogFilePath:          GetEnv("LOG_FILE_PATH", "./logs/user_service.log"),
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
