package cache

import (
	"Noviterra/server/app/model"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	_ "embed"
)

// 使用go:embed将lua文件内容编译时嵌入字符串变量

//go:embed Lua/like.lua
var likeLua string

//go:embed Lua/unlike.lua
var unlikeLua string

//go:embed Lua/collect.lua
var collectLua string

//go:embed Lua/uncollect.lua
var uncollectLua string

type CommentCache struct {
	rdb          *redis.Client
	likeSHA      string // 预加载like.lua脚本SHA
	unlikeSHA    string // 预加载unlike.lua脚本SHA
	collectSHA   string
	uncollectSHA string
	loadOnce     sync.Once // 脚本只加载一次
}

func NewCommentCache(cacheConfig RedisConfig) *CommentCache {
	return &CommentCache{
		rdb: NewRDB(cacheConfig),
	}
}

// LoadScripts Lua脚本导入 服务启动时调用一次
func (cc *CommentCache) LoadScripts(ctx context.Context) error {
	var loadErr error
	cc.loadOnce.Do(func() {
		cc.likeSHA, loadErr = cc.rdb.ScriptLoad(ctx, likeLua).Result()
		if loadErr != nil {
			return
		}
		cc.unlikeSHA, loadErr = cc.rdb.ScriptLoad(ctx, unlikeLua).Result()
		if loadErr != nil {
			return
		}

		cc.collectSHA, loadErr = cc.rdb.ScriptLoad(ctx, collectLua).Result()
		if loadErr != nil {
			return
		}

		cc.uncollectSHA, loadErr = cc.rdb.ScriptLoad(ctx, uncollectLua).Result()
		if loadErr != nil {
			return
		}
	})
	return loadErr
}

func (cc *CommentCache) Ping(ctx context.Context) error {
	return cc.rdb.Ping(ctx).Err()
}

// 统计某个键对应的值 int64
// 文章点赞、收藏个数
func (cc *CommentCache) mustGetCount(ctx context.Context, key string) (int64, error) {
	count, err := cc.rdb.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

// IsLiked 查询用户是否点赞
func (cc *CommentCache) IsLiked(ctx context.Context, articleID, userID string) (bool, error) {
	return cc.rdb.SIsMember(ctx, likeSetKey(articleID), userID).Result()
}

// IsCollect 查询用户是否收藏
func (cc *CommentCache) IsCollect(ctx context.Context, articleID, userID string) (bool, error) {
	return cc.rdb.SIsMember(ctx, collectSetKey(articleID), userID).Result()
}

// GetLikeCount 查询文章点赞总数
func (cc *CommentCache) GetLikeCount(ctx context.Context, articleID string) (int64, error) {
	return cc.mustGetCount(ctx, likeCntKey(articleID))
}

// GetCollectCount 查询文章收藏总数
func (cc *CommentCache) GetCollectCount(ctx context.Context, articleID string) (int64, error) {
	return cc.mustGetCount(ctx, collectCntKey(articleID))
}

// GetViewCount 查询文章新增的浏览量
func (cc *CommentCache) GetViewCount(ctx context.Context, articleID string) (int64, error) {
	return cc.mustGetCount(ctx, viewCntKey(articleID))
}

// GetUserLikedArticles 获取用户点赞过的所有文章ID
func (cc *CommentCache) GetUserLikedArticles(ctx context.Context, userID string) ([]string, error) {
	return cc.rdb.SMembers(ctx, userLikeKey(userID)).Result()
}

// GetUserCollectArticles 获取用户收藏列表
func (cc *CommentCache) GetUserCollectArticles(ctx context.Context, userID string) ([]string, error) {
	return cc.rdb.SMembers(ctx, userCollectKey(userID)).Result()
}

// BatchGetArticleStatus 批量查询用户对多篇文章的点赞、收藏状态（文章列表使用） 返回map[string][]bool 0、1分别表示点赞与收藏
func (cc *CommentCache) BatchGetArticleStatus(ctx context.Context,
	articleIDs []string, userID string) (map[string][]bool, error) {

	if len(articleIDs) == 0 {
		return map[string][]bool{}, nil
	}

	pipe := cc.rdb.Pipeline()
	length := len(articleIDs)
	likeCmds := make([]*redis.BoolCmd, length)
	collectCmds := make([]*redis.BoolCmd, length)
	for i, id := range articleIDs {
		likeCmds[i] = pipe.SIsMember(ctx, likeSetKey(id), userID)
		collectCmds[i] = pipe.SIsMember(ctx, collectSetKey(id), userID)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("批量获取用户点赞文章ID操作失败：%w", err)
	}

	result := make(map[string][]bool, len(articleIDs))
	for i, id := range articleIDs {
		result[id] = []bool{likeCmds[i].Val(), collectCmds[i].Val()}
	}
	return result, nil
}

// BatchGetArticleView 批量得到文章的新增浏览量
func (cc *CommentCache) BatchGetArticleView(ctx context.Context,
	articleIDs []string) (map[string]int64, error) {
	pipe := cc.rdb.Pipeline()
	viewCmds := make([]*redis.StringCmd, len(articleIDs))
	for i, id := range articleIDs {
		viewCmds[i] = pipe.Get(ctx, viewCntKey(id))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(articleIDs))
	for i, id := range articleIDs {
		result[id], _ = viewCmds[i].Int64()
	}

	return result, nil
}

// BatchDeleteArticleViewSet 删除已经同步过浏览量的计数统计
func (cc *CommentCache) BatchDeleteArticleViewSet(ctx context.Context,
	articleIDs []string) error {
	if len(articleIDs) == 0 {
		return nil
	}
	keys := make([]string, len(articleIDs))
	for i, id := range articleIDs {
		keys[i] = viewCntKey(id)
	}

	return cc.rdb.Del(ctx, keys...).Err()
}

// ScanViewKeys 扫描全部article:view:count:{article_id} 即文章浏览量的key 返回去重后的articleID列表
// 因为浏览量统计 在redis中并没有设置article-user-set 因此不能像点赞、收藏那样直接得到有值、值变化的articleID 因此需要直接从redis中scan出全部的keys
func (cc *CommentCache) ScanViewKeys(ctx context.Context) ([]string, error) {
	articleIDSet := make(map[string]struct{})
	// cursor下标 记录当前遍历到的位置 再次为0时表示遍历了一圈
	var cursor uint64
	for {
		// count:100 是提示值（hint） 不是返回的key数量 而是往返的哈希桶
		keys, nextCursor, err := cc.rdb.Scan(ctx, cursor, viewCntKey("*"), 100).Result()
		if err != nil {
			return nil, fmt.Errorf("扫描文章浏览量keys失败：%w", err)
		}
		for _, key := range keys {
			// 从key中准确提取出article_id
			articleID := strings.TrimPrefix(key, "article:view:count:")
			articleIDSet[articleID] = struct{}{}
		}
		cursor = nextCursor
		if cursor == 0 { // 0表示遍历已经结束
			break
		}
	}
	articleIDs := make([]string, 0, len(articleIDSet))
	for id := range articleIDSet {
		articleIDs = append(articleIDs, id)
	}

	return articleIDs, nil
}

// Collect 收藏操作
func (cc *CommentCache) Collect(ctx context.Context, articleID, userID string) (bool, int64, error) {
	result, err := cc.rdb.EvalSha(ctx, cc.collectSHA,
		[]string{ // KEY
			collectSetKey(articleID),
			collectCntKey(articleID),
			userCollectKey(userID),
		},
		userID, articleID, // ARGV
	).Int64()
	if err != nil {
		return false, 0, fmt.Errorf("收藏操作失败：%w", err)
	}

	count, err := cc.GetCollectCount(ctx, articleID)
	if err != nil {
		return false, 0, fmt.Errorf("获取收藏总数失败：%w", err)
	}

	return result == 1, count, nil
}

// Uncollect 收藏操作
func (cc *CommentCache) Uncollect(ctx context.Context, articleID, userID string) (bool, int64, error) {
	result, err := cc.rdb.EvalSha(ctx, cc.uncollectSHA,
		[]string{ // KEY
			collectSetKey(articleID),
			collectCntKey(articleID),
			userCollectKey(userID),
		},
		userID, articleID, // ARGV
	).Int64()
	if err != nil {
		return false, 0, fmt.Errorf("收藏操作失败：%w", err)
	}

	count, err := cc.GetCollectCount(ctx, articleID)
	if err != nil {
		return false, 0, fmt.Errorf("获取收藏总数失败：%w", err)
	}

	return result == 1, count, nil
}

// Like 点赞
// bool：false为点赞失败 true为点赞成功 int返回的为点赞之后的最新点赞总数
func (cc *CommentCache) Like(ctx context.Context, articleID, userID string) (bool, int64, error) {
	result, err := cc.rdb.EvalSha(ctx, cc.likeSHA,
		[]string{
			likeSetKey(articleID),
			likeCntKey(articleID),
			userLikeKey(userID),
		},
		userID, articleID,
	).Int64()
	if err != nil {
		return false, 0, fmt.Errorf("缓存点赞操作失败：%w", err)
	}

	count, err := cc.mustGetCount(ctx, likeCntKey(articleID))
	if err != nil {
		return false, 0, fmt.Errorf("获取点赞总数失败：%w", err)
	}

	return result == 1, count, nil
}

// Unlike 取消点赞
// bool：false为取消失败 true为取消成功 int返回取消点赞之后的最新点赞总数
func (cc *CommentCache) Unlike(ctx context.Context, articleID, userID string) (bool, int64, error) {
	result, err := cc.rdb.EvalSha(ctx, cc.unlikeSHA,
		[]string{
			likeSetKey(articleID),
			likeCntKey(articleID),
			userLikeKey(userID),
		},
		userID, articleID,
	).Int64()
	if err != nil {
		return false, 0, fmt.Errorf("缓存取消点赞操作失败：%w", err)
	}

	count, err := cc.mustGetCount(ctx, likeCntKey(articleID))
	if err != nil {
		return false, 0, fmt.Errorf("获取点赞总数失败：%w", err)
	}

	return result == 1, count, nil
}

// View 记录文章浏览次数 每次调用原子+1
func (cc *CommentCache) View(ctx context.Context, articleID string) error {
	return cc.rdb.Incr(ctx, viewCntKey(articleID)).Err()
}

// PushSyncEvent 入队 统一使用JSON序列化
func (cc *CommentCache) PushSyncEvent(ctx context.Context, event model.SyncEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("事务入队JSON序列化失败：%w", err)
	}
	return cc.rdb.LPush(ctx, syncQueueKey(), string(payload)).Err()
}

// PeekSyncEvents 查看但不删除同步事件
// 查看与删除两个方法进行处理 全量启动前提是成功执行
// 保障数据安全性 防止数据丢失
// eg：防止执行到第10个同步事件的时候 redis or mysql挂载导致同步事件丢失
// 执行到多少事件删除多少事件
func (cc *CommentCache) PeekSyncEvents(ctx context.Context, count int) ([]model.SyncEvent, error) {
	key := syncQueueKey()
	results, err := cc.rdb.LRange(ctx, key, 0, int64(count-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("PeekSyncEvents：%w", err)
	}
	events := make([]model.SyncEvent, 0, len(results))
	for _, raw := range results {
		var event model.SyncEvent
		if err = json.Unmarshal([]byte(raw), &event); err != nil {
			// 补日志
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

// AckSyncEvents 删除已经成功处理的同步事件
func (cc *CommentCache) AckSyncEvents(ctx context.Context, count int) error {
	key := syncQueueKey()
	return cc.rdb.LTrim(ctx, key, int64(count), -1).Err()
}

// 存在peek-ack并发竞态
