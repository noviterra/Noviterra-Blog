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
	collectArticleCommentModel interface {
		CollectArticle(ctx context.Context, in *commentrpc.CollectArticleRequest, opts ...grpc.CallOption) (*commentrpc.Response, error)
	}

	collectArticleCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	collectArticleServiceTokenInterceptor interface {
		GetServiceTokenInConf() string
	}
)

type CollectArticleLogic struct {
	ctx                     context.Context
	commentModel            collectArticleCommentModel
	ctxdata                 collectArticleCtxData
	serviceTokenInterceptor collectArticleServiceTokenInterceptor
}

// 对文章收藏操作
func NewCollectArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectArticleLogic {
	return &CollectArticleLogic{
		ctx:                     ctx,
		commentModel:            svcCtx.CommentModel,
		ctxdata:                 svcCtx.CtxData,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *CollectArticleLogic) CollectArticle(req *types.CollectArticleRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err = Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InternalError, errcode.Failed)
	}

	// 上下文获取用户ID
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		return failResponse(resp, "信息校验失败！", errcode.InvalidParam, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", id,
	)

	rep, err := l.commentModel.CollectArticle(ctx, &commentrpc.CollectArticleRequest{
		ArticleID:  req.ArticleID,
		Action:     req.Action,
		OperatorID: id,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Code != 0 {
		return failResponse(resp, rep.Message, int(rep.Status), int(rep.Code))
	}

	return successResponse(resp, rep.Message)
}

func successResponse(resp *types.Response, msg string) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}
	resp.Code = errcode.Success
	resp.Status = errcode.RequestSuccess
	resp.Message = msg

	return resp, nil
}

func failResponse(resp *types.Response, msg string, status, code int) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}
	resp.Code = code
	resp.Status = status
	resp.Message = msg

	return resp, nil
}

func failResponseWithErr(resp *types.Response, err error, status, code int) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}
	resp.Code = code
	resp.Status = status
	resp.Message = "服务失败：" + err.Error()

	return resp, nil
}
