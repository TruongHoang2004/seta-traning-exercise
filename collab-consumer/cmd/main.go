package main

import (
	"collab-consumer/internal/cache"
	"collab-consumer/internal/kafka"
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

// getEnv retrieves the environment variable with the given key
// If the variable is not set, it returns the fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	// Initialize Redis
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := 0 // Default DB is 0

	redisAddr := redisHost + ":" + redisPort
	log.Println("Connecting to Redis at", redisAddr)
	cache.InitRedis(redisAddr, redisPassword, redisDB)
	defer cache.CloseRedis()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a WaitGroup to wait for goroutines to finish
	var wg sync.WaitGroup

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Kafka configuration from environment variables
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	topic1 := getEnv("KAFKA_TOPIC_TEAM", "team.activity")
	groupID1 := getEnv("KAFKA_GROUP_ID", "collab-consumer-group")
	topic2 := getEnv("KAFKA_TOPIC_ASSET", "asset.changes")
	groupID2 := getEnv("KAFKA_GROUP_ID", "collab-consumer-group")

	log.Println("Starting consumers...")

	// Start consumers for different topics
	wg.Add(2)
	go func() {
		defer wg.Done()
		kafka.StartTeamConsumer(ctx, brokers, topic1, groupID1)
	}()

	go func() {
		defer wg.Done()
		kafka.StartAssetChangesConsumer(ctx, brokers, topic2, groupID2)
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-sigChan
	log.Printf("Received signal: %v, initiating graceful shutdown", sig)

	// Cancel the context to signal all consumers to stop
	cancel()

	// Wait for consumers to finish with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All consumers have shut down successfully")
	case <-time.After(10 * time.Second):
		log.Println("Shutdown timed out, some consumers may not have shut down properly")
	}

	log.Println("Application shutdown complete")
}
