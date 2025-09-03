package main

import (
	"collab-consumer/internal/cache"
	"collab-consumer/internal/kafka"
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// Initialize Redis
	cache.InitRedis("localhost:6379", "", 0)
	defer cache.CloseRedis()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a WaitGroup to wait for goroutines to finish
	var wg sync.WaitGroup

	// Kafka configuration
	brokers := []string{"localhost:9092"}
	topic1 := "team.activity"
	groupID1 := "collab-consumer-group"
	topic2 := "asset.changes"
	groupID2 := "collab-consumer-group"

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
