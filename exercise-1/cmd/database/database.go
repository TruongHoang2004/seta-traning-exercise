package main

import (
	"log"
	"os"
	"seta-training-exercise-1/database"
	"seta-training-exercise-1/models"
)

func reset() {
	log.Println("Resetting database...")

	// Kết nối DB
	database.Connect()
	db := database.DB

	// ⚠️ Drop toàn bộ bảng
	log.Println("Dropping tables...")
	err := db.Migrator().DropTable(&models.User{}, &models.Team{})
	if err != nil {
		log.Fatalf("Failed to drop tables: %v", err)
	}

	// Tạo lại bảng theo models
	log.Println("Migrating tables...")
	err = db.AutoMigrate(&models.User{}, &models.Team{})
	if err != nil {
		log.Fatalf("Auto migration failed: %v", err)
	}

	log.Println("✅ Database reset thành công.")
}

func main() {
	if len(os.Args) < 2 {
		log.Println("Usage: go run reset.go [reset|migrate|drop]")
		return
	}

	command := os.Args[1]

	switch command {
	case "reset":
		reset()
	case "migrate":
		migrate()
	case "drop":
		drop()
	default:
		log.Printf("Unknown command: %s\n", command)
		log.Println("Available commands: reset, migrate, drop")
	}
}

func migrate() {
	log.Println("Migrating database...")
	database.Connect()
	db := database.DB

	log.Println("Migrating tables...")
	err := db.AutoMigrate(&models.User{}, &models.Team{})
	if err != nil {
		log.Fatalf("Auto migration failed: %v", err)
	}

	log.Println("✅ Database migration completed successfully.")
}

func drop() {
	log.Println("Dropping database tables...")
	database.Connect()
	db := database.DB

	log.Println("Dropping tables...")
	err := db.Migrator().DropTable(&models.User{}, &models.Team{})
	if err != nil {
		log.Fatalf("Failed to drop tables: %v", err)
	}

	log.Println("✅ Tables dropped successfully.")
}
