package kafka

import (
	"collab-service/internal/infrastructure/logger"
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	logger.Info("Initializing Kafka producer", "brokers", brokers, "topic", topic)
	return &Producer{
		writer: NewWriter(brokers, topic),
	}
}

func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	err := p.writer.WriteMessages(ctx,
		kafka.Message{
			Key:   key,
			Value: value,
		},
	)
	if err != nil {
		logger.Error("Failed to publish message to kafka", "error", err.Error())
	}
	return err
}
