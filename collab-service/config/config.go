package config

import (
	"collab-service/internal/infrastructure/logger"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration from .env
type Config struct {
	Port                string
	Production          bool
	DBHost              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBPort              string
	JWTAccessSecret     string
	JWTRefreshSecret    string
	UserServiceEndpoint string
	LogFilePath         string
	RedisAddress        string
	RedisPassword       string
	KafkaAddresses      []string
	TeamActivityTopic   string
	AssetChangeTopic    string
}

// LoadEnv loads environment variables from .env file
func LoadEnv() error {
	// Find .env file in the project root or current directory
	err := godotenv.Load()
	if err != nil {
		logger.Warn("Warning: .env file not found")
		log.Printf("Warning: .env file not found: %v", err)
		// Continue execution even if .env is not found
		// Variables might be set in the environment directly
	}
	return nil
}

// GetConfig returns a Config struct populated with values from environment
func GetConfig() Config {
	return Config{
		Port:                GetEnv("PORT", "8080"),
		Production:          GetEnv("GO_ENV", "development") == "production",
		DBHost:              GetEnv("DB_HOST", "localhost"),
		DBUser:              GetEnv("DB_USER", "postgres"),
		DBPassword:          GetEnv("DB_PASSWORD", ""),
		DBName:              GetEnv("DB_NAME", "seta"),
		DBPort:              GetEnv("DB_PORT", "5432"),
		JWTAccessSecret:     GetEnv("JWT_ACCESS_SECRET", ""),
		JWTRefreshSecret:    GetEnv("JWT_REFRESH_SECRET", ""),
		UserServiceEndpoint: GetEnv("USER_SERVICE_ENDPOINT", "http://localhost:8081/graphql"),
		LogFilePath:         GetEnv("LOG_FILE_PATH", "./logs/collab_service.log"),
		RedisAddress:        GetEnv("REDIS_ADDRESS", "localhost:6379"),
		RedisPassword:       GetEnv("REDIS_PASSWORD", ""),
		KafkaAddresses:      []string{GetEnv("KAFKA_ADDRESS", "localhost:9092")},
		TeamActivityTopic:   GetEnv("TEAM_ACTIVITY_TOPIC", "team-activity"),
		AssetChangeTopic:    GetEnv("ASSET_CHANGE_TOPIC", "asset-change"),
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
