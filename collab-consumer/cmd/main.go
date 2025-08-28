package main

import (
	"collab-consumer/internal/cache"
	"collab-consumer/internal/kafka"
	"context"
)

func main() {

	cache.InitRedis("localhost:6379", "", 0)
	defer cache.CloseRedis()

	ctx := context.Background()
	brokers := []string{"localhost:9092"}
	topic := "team.activity"
	groupID := "collab-consumer-group"

	kafka.StartConsumer(ctx, brokers, topic, groupID)
}
