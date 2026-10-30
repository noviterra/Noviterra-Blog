// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/article/articleapi/internal/svc"
	"Noviterra/server/app/article/articleapi/internal/types"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
	"Noviterra/server/utils/errcode"
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type deleteArticleServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
}

type deleteArticleCtxData interface {
	GetUserIDFromContext(ctx context.Context) string
	GetUserRoleFromContext(ctx context.Context) int
}

type deleteArticleModel interface {
	DeleteArticle(ctx context.Context, in *articlerpc.DeleteArticleRequest, opts ...grpc.CallOption) (*articlerpc.Response, error)
}
type DeleteArticleLogic struct {
	ctx                     context.Context
	articleModel            deleteArticleModel
	ctxdata                 deleteArticleCtxData
	serviceTokenInterceptor deleteArticleServiceTokenInterceptor
}

func NewDeleteArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteArticleLogic {
	return &DeleteArticleLogic{
		ctx:                     ctx,
		ctxdata:                 svcCtx.CtxData,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
		articleModel:            svcCtx.ArticleModel,
	}
}

func (l *DeleteArticleLogic) DeleteArticle(req *types.DeleteArticleRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	// 从上下文中获取用户ID
	publisherID := l.ctxdata.GetUserIDFromContext(l.ctx)
	if publisherID == "" {
		return failResponse(resp, "用户ID为空！", errcode.InvalidParam, errcode.Failed)
	}

	// 获取用户身份权限
	role := l.ctxdata.GetUserRoleFromContext(l.ctx)
	if role != 1 { // 1为管理员
		return failResponse(resp, "用户身份权限不足！", errcode.InvalidParam, errcode.Failed)
	}

	// 用户信息与服务凭证注入metadata中
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", publisherID,
	)

	rep, err := l.articleModel.DeleteArticle(ctx, &articlerpc.DeleteArticleRequest{
		ArticleID: req.ArticleID,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Code != 0 {
		return failResponse(resp, rep.Message, errcode.InvalidParam, int(rep.Code))
	}

	return successResponse(resp, fmt.Sprintf("文章：%s删除成功！", req.ArticleID))
}

// 封装Response的对外导出 不绑定结构体之下

func failResponse(resp *types.Response, msg string, status, code int) (*types.Response, error) {
	resp.Status = status
	resp.Code = code
	resp.Message = msg

	return resp, nil
}

func failResponseWithErr(resp *types.Response, err error, status, code int) (*types.Response, error) {
	resp.Status = status
	resp.Code = code
	resp.Message = "服务错误：" + err.Error()

	return resp, nil
}

// 多个logic调用 msg自定义传入
func successResponse(resp *types.Response, msg string) (*types.Response, error) {
	resp.Status = errcode.RequestSuccess
	resp.Code = errcode.Success
	resp.Message = msg

	return resp, nil
}
