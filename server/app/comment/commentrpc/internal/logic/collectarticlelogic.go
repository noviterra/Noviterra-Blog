package logic

import (
	"context"

	"Noviterra/server/app/comment/commentrpc/internal/svc"
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"

	"go.uber.org/zap"
)

type (
	collectArticleCommentValidate interface {
		ValidateLikeComment(in validate.InputLikeComment, field string) (validate.OutputLikeComment, error)
	}

	collentArticleCommentResponse interface {
		SuccessResponse(resp *commentrpc.Response, msg string) (*commentrpc.Response, error)
		FailResponse(resp *commentrpc.Response, msg string, status int32, code int64) (*commentrpc.Response, error)
		FailResponseWithErr(resp *commentrpc.Response, err error) (*commentrpc.Response, error)
	}

	collectArticleCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	collectArticleCommentCache interface {
		Collect(ctx context.Context, articleID int64, userID string) (bool, int64, error)
		Uncollect(ctx context.Context, articleID int64, userID string) (bool, int64, error)
		PushSyncEvent(ctx context.Context, event model.SyncEvent) error
	}

	collectArticleLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type CollectArticleLogic struct {
	ctx             context.Context
	commentValidate collectArticleCommentValidate
	commentResponse collentArticleCommentResponse
	ctxdata         collectArticleCtxData
	commentCache    collectArticleCommentCache
	logger          collectArticleLogger
}

func NewCollectArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectArticleLogic {
	return &CollectArticleLogic{
		ctx:             ctx,
		commentValidate: svcCtx.CommentValidate,
		commentResponse: svcCtx.CommentResponse,
		ctxdata:         svcCtx.CtxData,
		commentCache:    svcCtx.CommentCache,
		logger:          svcCtx.Logger,
	}
}

// 对文章的收藏操作
func (l *CollectArticleLogic) CollectArticle(in *commentrpc.CollectArticleRequest) (*commentrpc.Response, error) {
	// todo: add your logic here and delete this line
	resp := new(commentrpc.Response)
	// 身份校验
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		l.logger.Error("上下文获取操作用户ID失败")
		return l.commentResponse.FailResponse(resp, "信息校验失败！",
			errcode.Forbidden, errcode.CommentUserInfoInvalid)
	}

	// 参数校验
	collectArticlereq, err := l.commentValidate.ValidateLikeComment(validate.InputLikeComment{
		ArticleID: in.ArticleID,
		UserID:    id,
		Action:    in.Action,
	}, "collect")
	if err != nil {
		l.logger.Error("校验收藏文章参数失败",
			zap.Int64("article-id", in.ArticleID),
			zap.String("user-info", id),
			zap.String("action", in.Action),
			zap.Error(err),
		)
		return l.commentResponse.FailResponseWithErr(resp, err)
	}
	switch collectArticlereq.Action {
	case "collect":
		return l.collectHandler(resp, collectArticlereq.ArticleID, collectArticlereq.UserID)
	case "uncollect":
		return l.unCollectHandler(resp, collectArticlereq.ArticleID, collectArticlereq.UserID)
	default:
		l.logger.Error("无效的收藏操作",
			zap.Int64("article-id", collectArticlereq.ArticleID),
			zap.String("user-info", collectArticlereq.UserID),
			zap.String("action", collectArticlereq.Action),
		)
		return l.commentResponse.FailResponse(resp, "无效的操作！",
			errcode.InternalError, errcode.CommentActionInvalid)
	}
}

func (l *CollectArticleLogic) collectHandler(resp *commentrpc.Response, articleID int64, userID string) (*commentrpc.Response, error) {
	result, _, err := l.commentCache.Collect(l.ctx, articleID, userID)
	if err != nil {
		l.logger.Error("缓存操作失败：收藏文章失败",
			zap.String("user-info", userID),
			zap.Int64("article-id", articleID),
			zap.Error(err),
		)
		return l.commentResponse.FailResponse(resp, "系统繁忙，收藏失败！",
			errcode.InternalError, errcode.CollectArticleFail)
	}
	if !result {
		l.logger.Error("用户重复收藏文章",
			zap.String("user-info", userID),
			zap.Int64("article-id", articleID),
		)
		return l.commentResponse.FailResponse(resp, "你已经收藏过了！",
			errcode.InvalidParam, errcode.ArticleAlreadyCollected)
	}

	// 将操作推入消息队列
	go func() {
		if err := l.commentCache.PushSyncEvent(context.Background(), model.SyncEvent{
			ArticleID: articleID,
			UserID:    userID,
			Action:    "collect",
		}); err != nil {
			l.logger.Error("缓存操作失败：写入收藏同步事件失败",
				zap.String("user-info", userID),
				zap.Int64("article-id", articleID),
				zap.Error(err),
			)
		}
	}()

	l.logger.Info("用户收藏文章成功",
		zap.String("user-info", userID),
		zap.Int64("article-id", articleID),
	)
	return l.commentResponse.SuccessResponse(resp, "收藏成功！")
}

func (l *CollectArticleLogic) unCollectHandler(resp *commentrpc.Response, articleID int64, userID string) (*commentrpc.Response, error) {
	result, _, err := l.commentCache.Uncollect(l.ctx, articleID, userID)
	if err != nil {
		l.logger.Error("缓存操作失败：取消收藏文章失败",
			zap.String("user-info", userID),
			zap.Int64("article-id", articleID),
			zap.Error(err),
		)
		return l.commentResponse.FailResponse(resp, "系统繁忙，取消收藏失败！",
			errcode.InternalError, errcode.UncollectArticleFail)
	}
	if !result {
		l.logger.Error("用户还未收藏文章",
			zap.String("user-info", userID),
			zap.Int64("article-id", articleID),
		)
		return l.commentResponse.FailResponse(resp, "你还未收藏！",
			errcode.InvalidParam, errcode.ArticleNotCollected)
	}

	// 将操作推入消息队列
	go func() {
		if err := l.commentCache.PushSyncEvent(context.Background(), model.SyncEvent{
			ArticleID: articleID,
			UserID:    userID,
			Action:    "uncollect",
		}); err != nil {
			l.logger.Error("缓存操作失败：写入取消收藏同步事件失败",
				zap.String("user-info", userID),
				zap.Int64("article-id", articleID),
				zap.Error(err),
			)
		}
	}()

	l.logger.Info("用户取消收藏文章成功",
		zap.String("user-info", userID),
		zap.Int64("article-id", articleID),
	)
	return l.commentResponse.SuccessResponse(resp, "取消收藏成功！")
}
