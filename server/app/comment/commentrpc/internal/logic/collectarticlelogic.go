package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"
	"context"

	"Noviterra/server/app/comment/commentrpc/internal/svc"
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"
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
		Collect(ctx context.Context, articleID, userID string) (bool, int64, error)
		Uncollect(ctx context.Context, articleID, userID string) (bool, int64, error)
		PushSyncEvent(ctx context.Context, event model.SyncEvent) error
	}
)

type CollectArticleLogic struct {
	ctx             context.Context
	commentValidate collectArticleCommentValidate
	commentResponse collentArticleCommentResponse
	ctxdata         collectArticleCtxData
	commentCache    collectArticleCommentCache
}

func NewCollectArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectArticleLogic {
	return &CollectArticleLogic{
		ctx:             ctx,
		commentValidate: svcCtx.CommentValidate,
		commentResponse: svcCtx.CommentResponse,
		ctxdata:         svcCtx.CtxData,
		commentCache:    svcCtx.CommentCache,
	}
}

// 对文章的收藏操作
func (l *CollectArticleLogic) CollectArticle(in *commentrpc.CollectArticleRequest) (*commentrpc.Response, error) {
	// todo: add your logic here and delete this line
	resp := new(commentrpc.Response)
	// 身份校验
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		return l.commentResponse.FailResponse(resp, "信息校验失败！", errcode.Forbidden, errcode.Failed)
	}

	// 参数校验
	collectArticlereq, err := l.commentValidate.ValidateLikeComment(validate.InputLikeComment{
		ArticleID: in.ArticleID,
		UserID:    id,
		Action:    in.Action,
	}, "collect")
	if err != nil {
		return l.commentResponse.FailResponseWithErr(resp, err)
	}
	switch collectArticlereq.Action {
	case "collect":
		return l.collectHandler(resp, collectArticlereq.ArticleID, collectArticlereq.UserID)
	case "uncollect":
		return l.unCollectHandler(resp, collectArticlereq.ArticleID, collectArticlereq.UserID)
	default:
		return l.commentResponse.FailResponse(resp, "无效的操作！", errcode.InternalError, errcode.Failed)
	}
}

func (l *CollectArticleLogic) collectHandler(resp *commentrpc.Response, articleID, userID string) (*commentrpc.Response, error) {
	result, _, err := l.commentCache.Collect(l.ctx, articleID, userID)
	if err != nil {
		return l.commentResponse.FailResponse(resp, "系统繁忙，收藏失败！", errcode.InternalError, errcode.Failed)
	}
	if !result {
		return l.commentResponse.FailResponse(resp, "你已经收藏过了！", errcode.InvalidParam, errcode.Failed)
	}

	// 将操作推入消息队列
	go l.commentCache.PushSyncEvent(context.Background(), model.SyncEvent{
		ArticleID: articleID,
		UserID:    userID,
		Action:    "collect",
	})

	return l.commentResponse.SuccessResponse(resp, "收藏成功！")
}

func (l *CollectArticleLogic) unCollectHandler(resp *commentrpc.Response, articleID, userID string) (*commentrpc.Response, error) {
	result, _, err := l.commentCache.Uncollect(l.ctx, articleID, userID)
	if err != nil {
		return l.commentResponse.FailResponse(resp, "系统繁忙，取消收藏失败！", errcode.InternalError, errcode.Failed)
	}
	if !result {
		return l.commentResponse.FailResponse(resp, "你还未收藏！", errcode.InvalidParam, errcode.Failed)
	}

	// 将操作推入消息队列
	go l.commentCache.PushSyncEvent(context.Background(), model.SyncEvent{
		ArticleID: articleID,
		UserID:    userID,
		Action:    "uncollect",
	})

	return l.commentResponse.SuccessResponse(resp, "取消收藏成功！")
}
