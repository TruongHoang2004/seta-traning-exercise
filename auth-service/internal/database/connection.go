package database

import (
	"fmt"
	"log"
	"os"
	"user-service/config"
	"user-service/internal/models"
	"user-service/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {

	// Lấy thông tin từ biến môi trường
	host := config.GetConfig().DBHost
	user := config.GetConfig().DBUser
	password := config.GetConfig().DBPassword
	dbname := config.GetConfig().DBName
	port := config.GetConfig().DBPort

	log.Printf("Connecting to database: %s on %s:%s", dbname, host, port)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	// Open database connection
	var dbErr error
	DB, dbErr = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if dbErr != nil {
		logger.Error("Failed to connect to database", dbErr)
		os.Exit(1)
	}

	// Get the underlying SQL DB object
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Error("Failed to get SQL DB object", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Note: To close the connection, call sqlDB.Close() when shutting down the application
	// This should be done in a main shutdown function or defer statement

	log.Printf("Successfully connected to database: %s", dbname)

	// Auto migrate models
	log.Println("Running auto migrations...")

	if err := DB.AutoMigrate(&models.User{}); err != nil {
		logger.Error("Auto migration for User failed", err)
	}

	logger.Info("Auto migrations completed successfully")
}

func Close() {
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Error("Failed to get SQL DB object for closing", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		logger.Error("Failed to close database connection", err)
	} else {
		logger.Info("Database connection closed successfully")
	}
}
