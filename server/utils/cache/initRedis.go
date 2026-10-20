package cache

import "github.com/redis/go-redis/v9"

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func NewRDB(redisConfig RedisConfig) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	return rdb
}
