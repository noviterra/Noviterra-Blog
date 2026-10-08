package rediscache

import "github.com/redis/go-redis/v9"

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type RedisCache struct {
	rdb *redis.Client
}

func NewRedisCache(redisConfig RedisConfig) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	return &RedisCache{
		rdb: rdb,
	}
}
