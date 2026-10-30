// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/comment/commentapi/internal/svc"
	"Noviterra/server/app/comment/commentapi/internal/types"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	likeArticleCommentModel interface {
		LikeArticle(ctx context.Context, in *commentrpc.LikeArticleRequest, opts ...grpc.CallOption) (*commentrpc.Response, error)
	}

	likeArticleCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	likeArticleServiceTokenInterceptor interface {
		GetServiceTokenInConf() string
	}
)

type LikeArticleLogic struct {
	ctx                     context.Context
	commentModel            likeArticleCommentModel
	ctxdata                 likeArticleCtxData
	serviceTokenInterceptor likeArticleServiceTokenInterceptor
}

// 对文章点赞操作
func NewLikeArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeArticleLogic {
	return &LikeArticleLogic{
		ctx:                     ctx,
		commentModel:            svcCtx.CommentModel,
		ctxdata:                 svcCtx.CtxData,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *LikeArticleLogic) LikeArticle(req *types.LikeArticleRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err = Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InternalError, errcode.Failed)
	}

	// 获取用户ID
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		return failResponse(resp, "信息校验失败！", errcode.InvalidParam, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", id,
	)

	rep, err := l.commentModel.LikeArticle(ctx, &commentrpc.LikeArticleRequest{
		ArticleID:  req.ArticleID,
		Action:     req.Action,
		OperatorID: id,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InternalError, errcode.Failed)
	}
	if rep.Code != 0 {
		return failResponse(resp, rep.Message, int(rep.Status), int(rep.Code))
	}

	return successResponse(resp, rep.Message)
}
