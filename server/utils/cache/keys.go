package cache

import (
	"fmt"
	"time"
)

const (
	// 点赞系列keys
	prefixLikeSet   = "article:like:set"   // 某文章点赞用户集合
	prefixLikeCnt   = "article:like:count" // 某文章点赞总数
	prefixUserLikes = "user:like:articles" // 某用户点赞过的文章列表

	// 收藏系列keys
	prefixCollectSet   = "article:collect:set"   // 某文章收藏用户集合
	prefixCollectCnt   = "article:collect:count" // 某文章收藏总数
	prefixUserCollects = "user:collect:articles" // 某用户收藏的文章列表

	// 浏览系列keys
	prefixViewCnt = "article:view:count" //  某文章的浏览总数

	// 异步同步keys
	prefixSyncQueue = "sync:queue" // 异步同步队列

	// 评论相关keys
	prefixTopComentList = "comment:top"  // 某文章or站点的顶层评论列表
	prefixTreeComment   = "comment:tree" // 某文章or站点下某个顶评下的完整评论
)

const (
	topCommentListTTL = 10 * time.Minute
	treeCommentTTL    = 15 * time.Minute
)

func likeSetKey(articleID int64) string {
	return fmt.Sprintf("%s:%d", prefixLikeSet, articleID)
}

func likeCntKey(articleID int64) string {
	return fmt.Sprintf("%s:%d", prefixLikeCnt, articleID)
}

func userLikeKey(userID string) string {
	return fmt.Sprintf("%s:%s", prefixUserLikes, userID)
}

func collectSetKey(articleID int64) string {
	return fmt.Sprintf("%s:%d", prefixCollectSet, articleID)
}

func collectCntKey(articleID int64) string {
	return fmt.Sprintf("%s:%d", prefixCollectCnt, articleID)
}

func userCollectKey(userID string) string {
	return fmt.Sprintf("%s:%s", prefixUserCollects, userID)
}

func viewCntKey(articleID int64) string {
	return fmt.Sprintf("%s:%d", prefixViewCnt, articleID)
}

func viewCntkeyPattern() string {
	return fmt.Sprintf("%s:*", prefixViewCnt)
}

func syncQueueKey() string {
	return prefixSyncQueue
}

func topCommentListKey(address int64) string {
	return fmt.Sprintf("%s:%d", prefixTopComentList, address)
}

func treeCommentKey(commentID int64) string {
	return fmt.Sprintf("%s:%d", prefixTreeComment, commentID)
}
