package cache

import "fmt"

const (
	// 点赞系列keys
	prefixLikeSet   = "article:like:set"   // 某文章点赞用户集合
	prefixLikeCnt   = "article:like:count" // 某文章点赞总数
	prefixUserLikes = "user:like:articles" // 某用户点赞过的文章列表

	// 收藏系列keys
	prefixCollectSet   = "article:collect:set"   // 某文章收藏用户集合
	prefixCollectCnt   = "article:collect:count" // 某文章收藏总数
	prefixUserCollects = "user:collect:articles" // 某用户收藏的文章列表

	// 异步同步keys
	prefixSyncQueue = "sync:queue" // 异步同步队列
)

func likeSetKey(articleID string) string {
	return fmt.Sprintf("%s:%s", prefixLikeSet, articleID)
}

func likeCntKey(articleID string) string {
	return fmt.Sprintf("%s:%s", prefixLikeCnt, articleID)
}

func userLikeKey(userID string) string {
	return fmt.Sprintf("%s:%s", prefixUserLikes, userID)
}

func collectSetKey(articleID string) string {
	return fmt.Sprintf("%s:%s", prefixCollectSet, articleID)
}

func collectCntKey(articleID string) string {
	return fmt.Sprintf("%s:%s", prefixCollectCnt, articleID)
}

func userCollectKey(userID string) string {
	return fmt.Sprintf("%s:%s", prefixUserCollects, userID)
}

func syncQueueKey() string {
	return prefixSyncQueue
}
