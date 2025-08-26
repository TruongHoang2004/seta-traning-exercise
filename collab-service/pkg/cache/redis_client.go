package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedis(addr, password string, db int) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,     // "localhost:6379"
		Password: password, // "" nếu không set pass
		DB:       db,       // default 0
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	log.Println("Connecting to Redis...")

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis connect error: " + err.Error())
	}

	log.Println("Connected to Redis")

}

func GetRedisClient() *redis.Client {
	return rdb
}

func CloseRedis() {
	if rdb != nil {
		_ = rdb.Close()
	}
}
