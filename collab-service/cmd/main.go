package main

import (
	"collab-service/internal/database"
	"collab-service/internal/routes"
	"collab-service/pkg/config"
	"collab-service/pkg/logger"

	"github.com/rs/zerolog"
	// Chỉnh đúng path nếu cần
)

// @title           Collab Service API
// @version         1.0
// @description     API for managing teams, folders, notes, and sharing.
// @host            localhost:8080
// @BasePath        /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter JWT token like: Bearer <your-token>

func main() {
	// Load environment variables
	config.LoadEnv()

	// Init logger
	logger.Init(config.GetConfig().Production, config.GetConfig().LogFilePath, zerolog.DebugLevel)
	defer logger.Close()

	// Connect to database
	database.Connect()
	defer database.Close()

	// Setup API routes
	router := routes.SetupRoutes()

	// Start server
	port := config.GetConfig().Port

	logger.Info("Server starting on port " + port)
	logger.Info("Swagger at http://localhost:" + port + "/swagger/index.html")

	if err := router.Run(":" + port); err != nil {
		logger.Error("Failed to start server: " + err.Error())
	}
}
