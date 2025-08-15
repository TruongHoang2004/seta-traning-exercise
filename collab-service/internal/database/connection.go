package database

import (
	"fmt"
	//  use zap log for better performance
	"collab-service/config"
	"collab-service/internal/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
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
		log.Fatal("Failed to connect to database:", dbErr)
	}

	// Get the underlying SQL DB object
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get SQL DB object:", err)
	}

	// Ping the database to ensure the connection is valid
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Database connection verified with ping")

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Note: To close the connection, call sqlDB.Close() when shutting down the application
	// This should be done in a main shutdown function or defer statement

	log.Printf("Successfully connected to database: %s", dbname)

	// Auto migrate models
	log.Println("Running auto migrations...")

	if err := DB.AutoMigrate(&models.Team{}); err != nil {
		log.Fatal("Auto migration for Team failed:", err)
	}
	if err := DB.AutoMigrate(&models.Roster{}); err != nil {
		log.Fatal("Auto migration for Roster failed:", err)
	}
	if err := DB.AutoMigrate(&models.Folder{}); err != nil {
		log.Fatal("Auto migration for Folder failed:", err)
	}
	if err := DB.AutoMigrate(&models.Note{}); err != nil {
		log.Fatal("Auto migration for Note failed:", err)
	}
	if err := DB.AutoMigrate(&models.NoteShare{}); err != nil {
		log.Fatal("Auto migration for NoteShare failed:", err)
	}
	if err := DB.AutoMigrate(&models.FolderShare{}); err != nil {
		log.Fatal("Auto migration for FolderShare failed:", err)
	}

	log.Println("Auto migrations completed successfully")
}

func Close() {
	if DB == nil {
		log.Println("Database connection is already closed or not initialized")
		return
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Println("Failed to get SQL DB object:", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Println("Failed to close database connection:", err)
	} else {
		log.Println("Database connection closed successfully")
	}
}
