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

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type updateArticleModel interface {
	UpdateArticle(ctx context.Context, in *articlerpc.UpdateArticleRequest, opts ...grpc.CallOption) (*articlerpc.Response, error)
}

type updateArticleAuthMiddleware interface {
	GetUserIDFromContext(ctx context.Context) string
	GetUserRoleFromContext(ctx context.Context) int
}

type updateArticleServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
}
type UpdateArticleLogic struct {
	ctx                     context.Context
	articleModel            updateArticleModel
	authMiddleware          updateArticleAuthMiddleware
	serviceTokenInterceptor updateArticleServiceTokenInterceptor
}

func NewUpdateArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateArticleLogic {
	return &UpdateArticleLogic{
		ctx:                     ctx,
		articleModel:            svcCtx.ArticleModel,
		authMiddleware:          svcCtx.AuthMiddleware,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *UpdateArticleLogic) UpdateArticle(req *types.UpdateArticleRequest) (resp *types.Response, err error) {
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
	publisherID := l.authMiddleware.GetUserIDFromContext(l.ctx)
	if publisherID == "" {
		return failResponse(resp, "用户ID为空！", errcode.InvalidParam, errcode.Failed)
	}

	// 获取用户身份权限
	role := l.authMiddleware.GetUserRoleFromContext(l.ctx)
	if role < 1 { // 0为用户权限最低
		return failResponse(resp, "用户身份权限不足！", errcode.InvalidParam, errcode.Failed)
	}

	// 信息注入到metadata中
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", publisherID,
	)

	// 调用rpc服务方法 执行核心业务
	rep, err := l.articleModel.UpdateArticle(ctx, &articlerpc.UpdateArticleRequest{
		ArticleID:   req.ArticleID,
		Title:       req.Title,
		Content:     req.Content,
		Tags:        req.Tags,
		Information: req.Information,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Code != 0 {
		return failResponse(resp, rep.Message, errcode.InvalidParam, int(rep.Code))
	}

	return successResponse(resp, "文章更新成功！")
}
