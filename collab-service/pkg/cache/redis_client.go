package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func InitRedis(addr, password string, db int) {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     addr,     // "localhost:6379"
		Password: password, // "" nếu không set pass
		DB:       db,       // default 0
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := Rdb.Ping(ctx).Err(); err != nil {
		panic("Redis connect error: " + err.Error())
	}

}
