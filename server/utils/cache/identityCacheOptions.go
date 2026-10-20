package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type IdentityCache struct {
	rdb *redis.Client
}

func NewIdentityCache(redisConfig RedisConfig) *IdentityCache {
	return &IdentityCache{
		rdb: NewRDB(redisConfig),
	}
}

func (ic *IdentityCache) LimitVerificationCode(ctx context.Context, email string) (bool, error) {
	// 以邮箱为唯一标识 功能不作为标识
	// SetNX原子操作 并发安全
	key := email
	ok, err := ic.rdb.SetNX(ctx, key, "1", time.Minute).Result()
	if err != nil {
		return false, err
	}
	// ok == true 表示创建成功 即key不存在
	return ok, nil
}

func (ic *IdentityCache) StoreVerificationCode(ctx context.Context, email, t, verificationCode string) error {
	// 以邮箱和验证码功能为标识 区分不同业务
	key := fmt.Sprintf("%s:%s", email, t)
	err := ic.rdb.Set(ctx, key, verificationCode, 5*time.Minute).Err()
	if err != nil {
		return err
	}
	return nil
}

func (ic *IdentityCache) GetVerificationCode(ctx context.Context, email, t string) (string, error) {
	key := fmt.Sprintf("%s:%s", email, t)
	value, err := ic.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

func (ic *IdentityCache) DeleteVerificationCode(ctx context.Context, email, t string) error {
	key := fmt.Sprintf("%s:%s", email, t)
	err := ic.rdb.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}
