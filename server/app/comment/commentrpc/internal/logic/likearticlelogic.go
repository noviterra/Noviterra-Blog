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
	likeArticleCommentResponse interface {
		SuccessResponse(resp *commentrpc.Response, msg string) (*commentrpc.Response, error)
		FailResponse(resp *commentrpc.Response, msg string, status int32, code int64) (*commentrpc.Response, error)
		FailResponseWithErr(resp *commentrpc.Response, err error) (*commentrpc.Response, error)
	}

	likeArticleCommentValidate interface {
		ValidateLikeComment(in validate.InputLikeComment, field string) (validate.OutputLikeComment, error)
	}

	likeArticleCommentCache interface {
		Like(ctx context.Context, articleID, userID string) (bool, int64, error)
		Unlike(ctx context.Context, articleID, userID string) (bool, int64, error)
		PushSyncEvent(ctx context.Context, event model.SyncEvent) error
	}
)

type LikeArticleLogic struct {
	ctx             context.Context
	commentResponse likeArticleCommentResponse
	commentValidate likeArticleCommentValidate
	commentCache    likeArticleCommentCache
}

func NewLikeArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeArticleLogic {
	return &LikeArticleLogic{
		ctx:             ctx,
		commentResponse: svcCtx.CommentResponse,
		commentValidate: svcCtx.CommentValidate,
		commentCache:    svcCtx.CommentCache,
	}
}

// 对文章的点赞操作
func (l *LikeArticleLogic) LikeArticle(in *commentrpc.LikeArticleRequest) (*commentrpc.Response, error) {
	// todo: add your logic here and delete this line
	resp := new(commentrpc.Response)
	// 参数校验
	likeInfo, err := l.commentValidate.ValidateLikeComment(validate.InputLikeComment{
		ArticleID: in.ArticleID,
		UserID:    in.OperatorID,
		Action:    in.Action,
	}, "like")
	if err != nil {
		return l.commentResponse.FailResponseWithErr(resp, err)
	}

	switch likeInfo.Action {
	case "like":
		return l.LikeHandler(resp, likeInfo.ArticleID, likeInfo.UserID)
	case "unlike":
		return l.UnLikeHandler(resp, likeInfo.ArticleID, likeInfo.UserID)
	default:
		return l.commentResponse.FailResponse(resp, "无效的操作！", errcode.InvalidParam, errcode.Failed)
	}
}

func (l *LikeArticleLogic) LikeHandler(resp *commentrpc.Response, articleID, userID string) (*commentrpc.Response, error) {
	result, _, err := l.commentCache.Like(l.ctx, articleID, userID)
	if err != nil {
		return l.commentResponse.FailResponse(resp, "系统繁忙，请稍后重试！", errcode.InternalError, errcode.Failed)
	}
	if !result {
		return l.commentResponse.FailResponse(resp, "你已经点过赞了！", errcode.InvalidParam, errcode.Failed)
	}

	// 异步推入同步队列 不阻塞用户响应
	go l.pushSync(articleID, userID, "like")

	return l.commentResponse.SuccessResponse(resp, "点赞成功！")
}

func (l *LikeArticleLogic) UnLikeHandler(resp *commentrpc.Response, articleID, userID string) (*commentrpc.Response, error) {
	result, _, err := l.commentCache.Unlike(l.ctx, articleID, userID)
	if err != nil {
		return l.commentResponse.FailResponse(resp, "系统繁忙，请稍后重试！", errcode.InternalError, errcode.Failed)
	}
	if !result {
		return l.commentResponse.FailResponse(resp, "你还未点赞！", errcode.InvalidParam, errcode.Failed)
	}

	// 异步推入同步队列 不阻塞用户响应
	go l.pushSync(articleID, userID, "unlike")

	return l.commentResponse.SuccessResponse(resp, "取消点赞成功！")
}

func (l *LikeArticleLogic) pushSync(articleID, userID, action string) {
	_ = l.commentCache.PushSyncEvent(context.Background(), model.SyncEvent{
		ArticleID: articleID,
		UserID:    userID,
		Action:    action,
	})
}
