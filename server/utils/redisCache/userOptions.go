package rediscache

import (
	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/types/userrpc"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func (r *RedisCache) LimitVerificationCode(ctx context.Context, email string) (bool, error) {
	// 以邮箱为唯一标识 功能不作为标识
	// SetNX原子操作 并发安全
	key := email
	ok, err := r.rdb.SetNX(ctx, key, "1", time.Minute).Result()
	if err != nil {
		return false, err
	}
	// ok == true 表示创建成功 即key不存在
	return ok, nil
}

func (r *RedisCache) StoreVerificationCode(ctx context.Context, email, t, verificationCode string) error {
	// 以邮箱和验证码功能为标识 区分不同业务
	key := fmt.Sprintf("%s:%s", email, t)
	err := r.rdb.Set(ctx, key, verificationCode, 5*time.Minute).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisCache) GetVerificationCode(ctx context.Context, email, t string) (string, error) {
	key := fmt.Sprintf("%s:%s", email, t)
	value, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

func (r *RedisCache) DeleteVerificationCode(ctx context.Context, email, t string) error {
	key := fmt.Sprintf("%s:%s", email, t)
	err := r.rdb.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

// 用户信息存入Redis-Hash类型中
func (r *RedisCache) HSetUserInfo(ctx context.Context, userInfo *model.UserInfo) error {
	key := fmt.Sprintf("%s:%s", userInfo.ID, "info")
	err := r.rdb.HSet(ctx, key,
		"id", userInfo.ID,
		"email", userInfo.Email,
		"name", userInfo.Name,
		"age", userInfo.Age,
		"gender", userInfo.Gender,
		"avatar", userInfo.Avatar,
		"selfIntroduction", userInfo.SelfIntroduction,
		"createdAt", userInfo.CreatedAt.Unix(),
		"updatedAt", userInfo.UpdatedAt.Unix(),
	).Err()
	if err != nil {
		return err
	}

	// 设置过期时间 180s
	ttl := 180
	if _, err = r.rdb.Expire(ctx, key, time.Duration(ttl)*time.Second).Result(); err != nil {
		return err
	}

	return nil
}

// 在缓存中查询用户信息
func (r *RedisCache) HGetUserInfo(ctx context.Context, id string) (*userrpc.UserInfoResponse_Data, error) {
	key := fmt.Sprintf("%s:%s", id, "info")
	vs, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// 缓存未命中
	if len(vs) == 0 {
		return nil, nil
	}

	userInfo := new(userrpc.UserInfoResponse_Data)

	userInfo.ID = vs["id"]
	userInfo.Email = vs["email"]
	userInfo.UserName = vs["name"]
	userInfo.Avatar = vs["avatar"]
	userInfo.SelfIntroduction = vs["selfIntroduction"]

	age, err := strconv.ParseInt(vs["age"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("年龄字段解析失败：%v", err)
	}
	gender, err := strconv.ParseInt(vs["gender"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("性别字段解析失败：%v", err)
	}
	userInfo.Age = age
	userInfo.Gender = gender

	createdTs, err := strconv.ParseInt(vs["createdAt"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("账号注册时间字段解析失败：%v", err)
	}
	updatedAt, err := strconv.ParseInt(vs["updatedAt"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("账号更新时间字段解析失败：%v", err)
	}
	userInfo.CreateAt = createdTs
	userInfo.UpdatedAt = updatedAt

	return userInfo, nil
}
