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

	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("Redis connect error: " + err.Error())
	}

}

func CloseRedis() {
	if rdb != nil {
		_ = rdb.Close()
	}
}

func GetRedisClient() *redis.Client {
	if rdb == nil {
		log.Fatal("Redis client is not initialized. Call InitRedis first.")
	}
	return rdb
}
