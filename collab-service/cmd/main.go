package main

import (
	"collab-service/config"
	"collab-service/internal/database"
	"collab-service/internal/routes"
	"collab-service/pkg/cache"
	"collab-service/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	// Load env
	config.LoadEnv()

	// Init logger
	logger.Init(config.GetConfig().Production, config.GetConfig().LogFilePath, zerolog.DebugLevel)
	defer logger.Close()

	// Connect DB + Redis
	database.Connect()
	defer database.Close()
	cache.InitRedis(config.GetConfig().RedisAddress, config.GetConfig().RedisPassword, 0)

	// Setup routes
	router := routes.SetupRoutes()

	// Create HTTP server manually (so we can shut it down)
	srv := &http.Server{
		Addr:    ":" + config.GetConfig().Port,
		Handler: router,
	}

	// Run server in goroutine
	go func() {
		logger.Info(fmt.Sprintf("Server starting on port %s", config.GetConfig().Port))
		logger.Info(fmt.Sprintf("Swagger at http://localhost:%s/swagger/index.html", config.GetConfig().Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server: " + err.Error())
		}
	}()

	// Wait for interrupt signal (CTRL+C, SIGTERM, SIGINT)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown HTTP server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: " + err.Error())
	}

	// Extra cleanup if needed
	logger.Info("Server exited cleanly")
}
