package kafka

import (
	"collab-consumer/internal/cache"
	"collab-consumer/internal/event"
	"collab-consumer/internal/handler"
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

func StartTeamConsumer(ctx context.Context, brokers []string, topic, groupID string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 1e3, // 1KB
		MaxBytes: 1e6, // 1MB
	})

	defer r.Close()

	// khởi tạo handler
	teamHandler := handler.NewTeamEventHandler(
		cache.NewCache(cache.GetRedisClient()),
	)

	log.Printf("Kafka consumer started, topic=%s group=%s\n", topic, groupID)

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			log.Printf("Error fetching message: %v", err)
			continue
		}

		// parse event từ message.Value
		ev, err := event.ParseEvent(m.Value)
		if err != nil {
			log.Printf("Error parsing event: %v", err)
			continue
		}

		// gọi handler
		if err := teamHandler.HandleTeamEvent(ctx, ev); err != nil {
			log.Printf("Error handling event: %v", err)
		}

		// Commit offset sau khi xử lý thành công
		if err := r.CommitMessages(ctx, m); err != nil {
			log.Printf("Error committing message: %v", err)
		}
	}
}

func StartAssetChangesConsumer(ctx context.Context, brokers []string, topic, groupID string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 1e3, // 1KB
		MaxBytes: 1e6, // 1MB
	})

	defer r.Close()

	// khởi tạo handler
	assetHandler := handler.NewAssetChangesHandler(
		cache.NewAssetCache(cache.GetRedisClient()),
	)

	log.Printf("Kafka consumer started, topic=%s group=%s\n", topic, groupID)

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			log.Printf("Error fetching message: %v", err)
			continue
		}

		// parse event từ message.Value
		ev, err := event.ParseEvent(m.Value)
		if err != nil {
			log.Printf("Error parsing event: %v", err)
			continue
		}

		// gọi handler
		if err := assetHandler.HandleAssetEvent(ctx, ev); err != nil {
			log.Printf("Error handling event: %v", err)
		}

		// Commit offset sau khi xử lý thành công
		if err := r.CommitMessages(ctx, m); err != nil {
			log.Printf("Error committing message: %v", err)
		}
	}
}
