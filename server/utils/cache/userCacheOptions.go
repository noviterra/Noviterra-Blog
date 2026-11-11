package cache

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/common"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type cacheTypeConversion interface {
	StringConversionTime(data string) (time.Time, error)
}

type UserCache struct {
	rdb            *redis.Client
	commonUtils    *common.CommonUtils
	typeConversion cacheTypeConversion
}

func NewUserCache(redisConfig RedisConfig,
	commonUtils *common.CommonUtils,
	typeConversion cacheTypeConversion) *UserCache {
	return &UserCache{
		rdb:            NewRDB(redisConfig),
		commonUtils:    commonUtils,
		typeConversion: typeConversion,
	}
}

// 用户信息存入Redis-Hash类型中
func (uc *UserCache) HSetUserInfo(ctx context.Context, userInfo *model.UserInfo) error {
	key := fmt.Sprintf("%s:%s", userInfo.ID, "info")
	err := uc.rdb.HSet(ctx, key,
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
	if _, err = uc.rdb.Expire(ctx, key, time.Duration(ttl)*time.Second).Result(); err != nil {
		return err
	}

	return nil
}

// 在缓存中查询用户信息
func (uc *UserCache) HGetUserInfo(ctx context.Context, id string) (*model.UserInfo, error) {
	key := fmt.Sprintf("%s:%s", id, "info")
	vs, err := uc.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// 缓存未命中
	if len(vs) == 0 {
		return nil, nil
	}

	userInfo := new(model.UserInfo)

	userInfo.ID = vs["id"]
	userInfo.Email = vs["email"]
	userInfo.Name = vs["name"]
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
	userInfo.Age = int(age)
	userInfo.Gender = int(gender)

	// string转int64类型
	createdTs, err := uc.typeConversion.StringConversionTime(vs["createdAt"])
	if err != nil {
		return nil, fmt.Errorf("账号注册时间字段解析失败：%v", err)
	}
	updatedAt, err := uc.typeConversion.StringConversionTime(vs["updatedAt"])
	if err != nil {
		return nil, fmt.Errorf("账号更新时间字段解析失败：%v", err)
	}
	userInfo.CreatedAt = createdTs
	userInfo.UpdatedAt = updatedAt

	return userInfo, nil
}

// DelUserInfo 删除用户在缓存中的数据
func (uc *UserCache) DelUserInfo(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s:%s", id, "info")
	return uc.rdb.Del(ctx, key).Err()
}
