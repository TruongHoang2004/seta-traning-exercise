package database

import (
	"fmt"
	"log"
	"os"
	"seta-training-exercise-1/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	// Load file .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Lấy thông tin từ biến môi trường
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	log.Printf("Connecting to database: %s on %s:%s", dbname, host, port)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Printf("Successfully connected to database: %s", dbname)

	// Auto migrate models
	log.Println("Running auto migrations...")

	if err := DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("Auto migration for User failed:", err)
	}
	if err := DB.AutoMigrate(&models.Team{}); err != nil {
		log.Fatal("Auto migration for Team failed:", err)
	}
	if err := DB.AutoMigrate(&models.Roster{}); err != nil {
		log.Fatal("Auto migration for Roster failed:", err)
	}

	log.Println("Auto migrations completed successfully")
}
