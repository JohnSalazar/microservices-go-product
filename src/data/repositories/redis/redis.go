package redis_repository

import (
	"time"

	"github.com/JohnSalazar/microservices-go-common/config"
	"github.com/go-redis/redis/v8"
)

func NewRedisClient(config *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:        config.Redis.Addr,
		DB:          config.Redis.Db,
		PoolSize:    config.Redis.PoolSize,
		ReadTimeout: 5 * time.Second,
		DialTimeout: 5 * time.Second,
	})

	return client
}
