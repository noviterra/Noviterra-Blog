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
	commentArticleCommentModel interface {
		CreateComment(ctx context.Context, in *commentrpc.CommentRequest, opts ...grpc.CallOption) (*commentrpc.CommentResponse, error)
	}

	commentArticleCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	commentArticleServiceTokenInterceptor interface {
		GetServiceTokenInConf() string
	}
)

type CommentArticleLogic struct {
	ctx                     context.Context
	commentModel            commentArticleCommentModel
	ctxdata                 collectArticleCtxData
	serviceTokenInterceptor commentArticleServiceTokenInterceptor
}

// 对一篇文章进行评论
func NewCommentArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentArticleLogic {
	return &CommentArticleLogic{
		ctx:                     ctx,
		commentModel:            svcCtx.CommentModel,
		ctxdata:                 svcCtx.CtxData,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *CommentArticleLogic) CommentArticle(req *types.CommentArticleRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	// 身份校验
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		return failResponse(resp, "信息校验失败！", errcode.InvalidParam, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", id,
	)

	rep, err := l.commentModel.CreateComment(ctx, &commentrpc.CommentRequest{
		Address:   req.Comment.Address,
		ParentID:  req.Comment.ParentID,
		SpeakerID: req.Comment.SpeakerID,
		Content:   req.Comment.Content,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return failResponse(resp, rep.Resp.Message, int(rep.Resp.Status), int(rep.Resp.Code))
	}

	return successResponse(resp, rep.Resp.Message)
}
