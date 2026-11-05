package woker

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/cache"
	"Noviterra/server/utils/repo"
	"context"
	"fmt"
	"time"
)

type LikeSyncWoker struct {
	commentCache      *cache.CommentCache
	commentRepository *repo.CommentRepository
	batchSize         int           // 每次最多拉取多少条队列数据
	interval          time.Duration // 拉取间隔
	stopCh            chan struct{}
}

func NewLikeSyncWoker(commentCache *cache.CommentCache,
	commentRepository *repo.CommentRepository) *LikeSyncWoker {
	return &LikeSyncWoker{
		commentCache:      commentCache,
		commentRepository: commentRepository,
		batchSize:         200,
		interval:          5 * time.Second,
		stopCh:            make(chan struct{}),
	}
}

// Start 启动后台循环
func (lsw *LikeSyncWoker) Start(ctx context.Context) {
	go lsw.loop(ctx)
}

// Stop 优雅关闭
func (lsw *LikeSyncWoker) Stop() {
	close(lsw.stopCh)
}

// loop 循环操作
func (lsw *LikeSyncWoker) loop(ctx context.Context) {
	ticker := time.NewTicker(lsw.interval)
	defer ticker.Stop()
	// 定时任务使用独立的 Background context，不受外部 cancel 影响
	for {
		select {
		case <-ctx.Done():
			// 退出前做最后一次flush 尽量不丢数据
			lsw.flush(context.Background())
			return
		case <-lsw.stopCh:
			lsw.flush(context.Background())
			return
		case <-ticker.C:
			lsw.flush(context.Background())
		}
	}
}

func (lsw *LikeSyncWoker) flush(ctx context.Context) {
	// 查看（不删除）事件
	events, err := lsw.commentCache.PeekSyncEvents(ctx, lsw.batchSize)
	if err != nil {
		return
	}
	if len(events) == 0 {
		return
	}

	// 处理所有事件 去重聚合、写入DB
	if err := lsw.proccessEvents(ctx, events); err != nil {
		return
	}

	// 全部成功后再进行删除
	if err := lsw.commentCache.AckSyncEvents(ctx, len(events)); err != nil {
		return
	}
}

// proccessEvents 将事件去重聚合后批量处理
// 事件记录的数据包含了点赞系列与收藏系列
// userArticleSet数据结构变为map[string]map[string][]bool bool切片 0为点赞状态 1为收藏状态
func (lsw *LikeSyncWoker) proccessEvents(ctx context.Context, events []model.SyncEvent) error {
	// 去重聚合
	userLikeArticleSet := make(map[string]map[int64]bool) // userID -> set<articleID>
	articleLikeSet := make(map[int64]bool)                // 所有涉及点赞的文章
	userCollectArticleSet := make(map[string]map[int64]bool)
	articleCollectSet := make(map[int64]bool)

	// 区分action
	for _, event := range events {
		if userLikeArticleSet[event.UserID] == nil {
			userLikeArticleSet[event.UserID] = make(map[int64]bool)
		}
		if userCollectArticleSet[event.UserID] == nil {
			userCollectArticleSet[event.UserID] = make(map[int64]bool)
		}
		// 一个事件仅记录一个行为 一个用户对一个文章点赞并收藏为两个事件
		// 不会丢失收藏事件
		if event.Action == "like" || event.Action == "unlike" {
			userLikeArticleSet[event.UserID][event.ArticleID] = true
			articleLikeSet[event.ArticleID] = true
		} else {
			userCollectArticleSet[event.UserID][event.ArticleID] = true
			articleCollectSet[event.ArticleID] = true
		}
	}

	// 批量查询Redis （每个用户与每篇文章只查一次）
	// 查询每篇文章的点赞总数
	articleLikeCounts := make(map[int64]int64, len(articleLikeSet))
	for articleID := range articleLikeSet {
		count, err := lsw.commentCache.GetLikeCount(ctx, articleID)
		if err != nil {
			return fmt.Errorf("获取文章点赞总数失败：%w", err)
		}
		articleLikeCounts[articleID] = count
	}

	// 每个用户点赞的文章列表
	userLikedArticles := make(map[string][]string, len(userLikeArticleSet))
	for userID := range userLikeArticleSet {
		articles, err := lsw.commentCache.GetUserLikedArticles(ctx, userID)
		if err != nil {
			return fmt.Errorf("获取用户点赞文章列表失败：%w", err)
		}
		userLikedArticles[userID] = articles
	}

	// 每篇文章的收藏总数
	articleCollectCount := make(map[int64]int64, len(articleCollectSet))
	for articleID := range articleCollectSet {
		count, err := lsw.commentCache.GetCollectCount(ctx, articleID)
		if err != nil {
			return fmt.Errorf("获取文章收藏总数失败：%w", err)
		}
		articleCollectCount[articleID] = count
	}

	// 每个用户的收藏文章列表
	userCollectArticles := make(map[string][]string, len(userCollectArticleSet))
	for userID := range userCollectArticleSet {
		articles, err := lsw.commentCache.GetUserCollectArticles(ctx, userID)
		if err != nil {
			return fmt.Errorf("获取用户收藏文章列表失败：%w", err)
		}
		userCollectArticles[userID] = articles
	}

	// 查询每篇文章的新增浏览量
	viewArticleIDs, err := lsw.commentCache.ScanViewKeys(ctx)
	if err != nil {
		return fmt.Errorf("扫描浏览量key失败：%w", err)
	}
	var viewCount map[string]int64
	if len(viewCount) > 0 {
		viewCount, err = lsw.commentCache.BatchGetArticleView(ctx, viewArticleIDs)
		if err != nil {
			return fmt.Errorf("获取文章浏览量失败：%w", err)
		}
	}

	// 在一个事务中批量写入DB
	err = lsw.commentRepository.BatchUpsertSyncStatus(ctx, userLikedArticles, articleLikeCounts,
		userCollectArticles, articleCollectCount, viewCount)
	if err != nil {
		return fmt.Errorf("同步DB数据失败：%w", err)
	}

	// 删除文章新增浏览量的记录
	err = lsw.commentCache.BatchDeleteArticleViewSet(ctx, viewArticleIDs)
	if err != nil {
		// 实际是去记录日志而不是返回错误
		return fmt.Errorf("删除文章新增浏览量缓存数据失败：%w", err)
	}

	return nil
}
