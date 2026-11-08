package cache

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/common"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ArticleCache struct {
	rdb             *redis.Client
	typesConversion *common.TypesConversion
}

func NewArticleCache(redisConfig RedisConfig,
	typesConversion *common.TypesConversion) *ArticleCache {
	return &ArticleCache{
		rdb:             NewRDB(redisConfig),
		typesConversion: typesConversion,
	}
}

// AddArticleToRank 将文章加入到排行榜中
func (ac *ArticleCache) AddArticleToRank(ctx context.Context, articleID int64) error {
	pipe := ac.rdb.Pipeline()

	keyView := "rank:article:view"
	keyLike := "rank:article:like"
	keyCollect := "rank:article:collect"

	cmdView := pipe.ZAdd(ctx, keyView, redis.Z{Score: 0, Member: articleID})
	cmdLike := pipe.ZAdd(ctx, keyLike, redis.Z{Score: 0, Member: articleID})
	cmdCollect := pipe.ZAdd(ctx, keyCollect, redis.Z{Score: 0, Member: articleID})

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	// 逐个检查命令错误 精准到某个ZAdd命令
	for _, cmd := range []*redis.IntCmd{cmdView, cmdLike, cmdCollect} {
		if err := cmd.Err(); err != nil {
			return fmt.Errorf("文章排行榜插入失败，failed：%d，err：%s", articleID, err.Error())
		}
	}
	return nil
}

// HSetArticle 将article信息存入缓存哈希表中
func (ac *ArticleCache) HSetArticle(ctx context.Context, article *model.ArticleInfo) error {
	key := fmt.Sprintf("%d:%s", article.ArticleID, "info")

	// 将tags序列化
	tagsJSON, err := json.Marshal(article.Tags)
	if err != nil {
		return fmt.Errorf("tags序列化失败：%w", err)
	}
	err = ac.rdb.HSet(ctx, key,
		"article_id", article.ArticleID,
		"publisher_id", article.PublisherID,
		"title", article.Title,
		"cover", article.Cover,
		"content", article.Content,
		"tags", tagsJSON,
		"information", article.Information,
		"created_at", article.CreatedAt,
		"updated_at", article.UpdatedAt,
	).Err()
	if err != nil {
		return fmt.Errorf("缓存设置文章信息失败：%w", err)
	}

	// 时间设置为1小时+随机数 防止缓存雪崩
	ttl := 3600 + 1
	if _, err := ac.rdb.Expire(ctx, key, time.Duration(ttl)*time.Second).Result(); err != nil {
		return err
	}
	return nil
}

// HGetArticle 根据文章ID获取存储在缓存哈希表中的文章数据
func (ac *ArticleCache) HGetArticle(ctx context.Context, articleID int64) (*model.Article, error) {
	key := fmt.Sprintf("%d:%s", articleID, "info")
	vs, err := ac.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("获取缓存数据失败：%w", err)
	}

	// 缓存未命中
	if len(vs) == 0 {
		return nil, nil
	}

	article := new(model.Article)

	article.PublisherID = vs["publisher_id"]
	article.Title = vs["title"]
	article.Content = vs["content"]
	article.Information = vs["information"]
	article.Cover = vs["cover"]

	// 其他非string类型 转换+赋值
	if tagStr, ok := vs["tags"]; ok && tagStr != "" {
		if err := json.Unmarshal([]byte(tagStr), &article.Tags); err != nil {
			return nil, fmt.Errorf("解析tags失败: %w", err)
		}
	}
	//string转为int64
	article.ArticleID, err = ac.typesConversion.StringConversionInt64(vs["article_id"])
	if err != nil {
		return nil, fmt.Errorf("解析article_id失败：%w", err)
	}
	article.CreatedAt, err = ac.typesConversion.StringConversionInt64(vs["created_at"])
	if err != nil {
		return nil, fmt.Errorf("解析createdAt失败：%w", err)
	}
	article.UpdatedAt, err = ac.typesConversion.StringConversionInt64(vs["updated_at"])
	if err != nil {
		return nil, fmt.Errorf("解析updatedAt失败：%w", err)
	}

	return article, nil
}
