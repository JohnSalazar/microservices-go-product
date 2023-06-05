package redis_repository

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/oceano-dev/microservices-go-common/config"
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
