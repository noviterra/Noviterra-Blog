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
	likeArticleCommentResponse interface {
		SuccessResponse(resp *commentrpc.Response, msg string) (*commentrpc.Response, error)
		FailResponse(resp *commentrpc.Response, msg string, status int32, code int64) (*commentrpc.Response, error)
		FailResponseWithErr(resp *commentrpc.Response, err error) (*commentrpc.Response, error)
	}

	likeArticleCommentValidate interface {
		ValidateLikeComment(in validate.InputLikeComment, field string) (validate.OutputLikeComment, error)
	}

	likeArticleCommentCache interface {
		Like(ctx context.Context, articleID int64, userID string) (bool, int64, error)
		Unlike(ctx context.Context, articleID int64, userID string) (bool, int64, error)
		PushSyncEvent(ctx context.Context, event model.SyncEvent) error
	}

	likeArticleCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	likeArticleLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type LikeArticleLogic struct {
	ctx             context.Context
	commentResponse likeArticleCommentResponse
	commentValidate likeArticleCommentValidate
	commentCache    likeArticleCommentCache
	ctxdata         likeArticleCtxData
	logger          likeArticleLogger
}

func NewLikeArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeArticleLogic {
	return &LikeArticleLogic{
		ctx:             ctx,
		commentResponse: svcCtx.CommentResponse,
		commentValidate: svcCtx.CommentValidate,
		commentCache:    svcCtx.CommentCache,
		ctxdata:         svcCtx.CtxData,
		logger:          svcCtx.Logger,
	}
}

// 对文章的点赞操作
func (l *LikeArticleLogic) LikeArticle(in *commentrpc.LikeArticleRequest) (*commentrpc.Response, error) {
	// todo: add your logic here and delete this line
	resp := new(commentrpc.Response)
	// 身份校验
	userID := l.ctxdata.GetUserIDFromContext(l.ctx)
	if userID == "" {
		l.logger.Error("上下文获取操作用户ID失败")
		return l.commentResponse.FailResponse(resp, "用户信息校验失败！",
			errcode.Unauthorized, errcode.CommentUserInfoInvalid)
	}

	// 参数校验
	likeInfo, err := l.commentValidate.ValidateLikeComment(validate.InputLikeComment{
		ArticleID: in.ArticleID,
		UserID:    in.OperatorID,
		Action:    in.Action,
	}, "like")
	if err != nil {
		l.logger.Error("校验点赞文章参数失败",
			zap.Int64("article-id", in.ArticleID),
			zap.String("user-info", userID),
			zap.String("operator-id", in.OperatorID),
			zap.String("action", in.Action),
			zap.Error(err),
		)
		return l.commentResponse.FailResponseWithErr(resp, err)
	}

	switch likeInfo.Action {
	case "like":
		return l.LikeHandler(resp, likeInfo.ArticleID, likeInfo.UserID)
	case "unlike":
		return l.UnLikeHandler(resp, likeInfo.ArticleID, likeInfo.UserID)
	default:
		l.logger.Error("无效的点赞操作",
			zap.Int64("article-id", likeInfo.ArticleID),
			zap.String("user-info", likeInfo.UserID),
			zap.String("action", likeInfo.Action),
		)
		return l.commentResponse.FailResponse(resp, "无效的操作！",
			errcode.InvalidParam, errcode.CommentActionInvalid)
	}
}

func (l *LikeArticleLogic) LikeHandler(resp *commentrpc.Response, articleID int64, userID string) (*commentrpc.Response, error) {
	result, _, err := l.commentCache.Like(l.ctx, articleID, userID)
	if err != nil {
		l.logger.Error("缓存操作失败：点赞文章失败",
			zap.String("user-info", userID),
			zap.Int64("article-id", articleID),
			zap.Error(err),
		)
		return l.commentResponse.FailResponse(resp, "系统繁忙，请稍后重试！",
			errcode.InternalError, errcode.LikeArticleFail)
	}
	if !result {
		l.logger.Error("用户重复点赞文章",
			zap.String("user-info", userID),
			zap.Int64("article-id", articleID),
		)
		return l.commentResponse.FailResponse(resp, "你已经点过赞了！",
			errcode.InvalidParam, errcode.ArticleAlreadyLiked)
	}

	// 异步推入同步队列 不阻塞用户响应
	go l.pushSync(articleID, userID, "like")

	l.logger.Info("用户点赞文章成功",
		zap.String("user-info", userID),
		zap.Int64("article-id", articleID),
	)
	return l.commentResponse.SuccessResponse(resp, "点赞成功！")
}

func (l *LikeArticleLogic) UnLikeHandler(resp *commentrpc.Response, articleID int64, userID string) (*commentrpc.Response, error) {
	result, _, err := l.commentCache.Unlike(l.ctx, articleID, userID)
	if err != nil {
		l.logger.Error("缓存操作失败：取消点赞文章失败",
			zap.String("user-info", userID),
			zap.Int64("article-id", articleID),
			zap.Error(err),
		)
		return l.commentResponse.FailResponse(resp, "系统繁忙，请稍后重试！",
			errcode.InternalError, errcode.UnlikeArticleFail)
	}
	if !result {
		l.logger.Error("用户还未点赞文章",
			zap.String("user-info", userID),
			zap.Int64("article-id", articleID),
		)
		return l.commentResponse.FailResponse(resp, "你还未点赞！",
			errcode.InvalidParam, errcode.ArticleNotLiked)
	}

	// 异步推入同步队列 不阻塞用户响应
	go l.pushSync(articleID, userID, "unlike")

	l.logger.Info("用户取消点赞文章成功",
		zap.String("user-info", userID),
		zap.Int64("article-id", articleID),
	)
	return l.commentResponse.SuccessResponse(resp, "取消点赞成功！")
}

func (l *LikeArticleLogic) pushSync(articleID int64, userID, action string) {
	err := l.commentCache.PushSyncEvent(context.Background(), model.SyncEvent{
		ArticleID: articleID,
		UserID:    userID,
		Action:    action,
	})
	if err != nil {
		l.logger.Error("缓存操作失败：写入点赞同步事件失败",
			zap.String("user-info", userID),
			zap.Int64("article-id", articleID),
			zap.String("action", action),
			zap.Error(err),
		)
	}
}
