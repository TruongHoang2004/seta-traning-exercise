package database

import (
	"fmt"
	//  use zap log for better performance
	"collab-service/config"
	"collab-service/internal/infrastructure/persistence/model"

	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func Connect(connectString string) {

	var dsn string
	if connectString == "" {
		host := config.GetConfig().DBHost
		user := config.GetConfig().DBUser
		password := config.GetConfig().DBPassword
		dbname := config.GetConfig().DBName
		port := config.GetConfig().DBPort

		log.Printf("Connecting to database: %s on %s:%s", dbname, host, port)

		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			host, user, password, dbname, port)
	}

	// Open database connection
	var dbErr error
	db, dbErr = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if dbErr != nil {
		log.Fatal("Failed to connect to database:", dbErr)
	}

	// Get the underlying SQL DB object
	sqlDB, err := db.DB()
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

	log.Printf("Successfully connected to database: %s", dsn)

	// Auto migrate models
	log.Println("Running auto migrations...")

	if !config.GetConfig().Production {
		db = db.Debug()
	}

	db.AutoMigrate(&model.TeamModel{}, &model.RosterModel{}, &model.FolderModel{}, &model.NoteModel{}, &model.NoteShareModel{}, &model.FolderShareModel{})

	log.Println("Auto migrations completed successfully")
}

func Close() {
	if db == nil {
		log.Println("Database connection is already closed or not initialized")
		return
	}

	sqlDB, err := db.DB()
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

func GetDB() *gorm.DB {
	return db
}
