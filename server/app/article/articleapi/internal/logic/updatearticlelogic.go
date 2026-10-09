// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
	"context"
	"errors"
	"net/http"

	"Noviterra/server/app/article/articleapi/internal/svc"
	"Noviterra/server/app/article/articleapi/internal/types"

	"github.com/go-playground/validator/v10"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type UpdateArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateArticleLogic {
	return &UpdateArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateArticleLogic) UpdateArticle(req *types.UpdateArticleRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) && len(errs) > 0 {
			resp.Status = http.StatusBadRequest
			resp.Code = 1
			resp.Message = TypeName[errs[0].Field()] + "格式错误！"
		}
		return resp, nil
	}

	// 从上下文中获取用户ID
	publisherID := l.svcCtx.Middleware.GetUserIDFromContext(l.ctx)
	if publisherID == "" {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "用户ID为空！"
		return resp, nil
	}

	// 获取用户身份权限
	role := l.svcCtx.Middleware.GetUserRoleFromContext(l.ctx)
	if role != 1 { // 1为管理员
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "用户身份权限不足！"
		return resp, nil
	}

	// 信息注入到metadata中
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.svcCtx.Config.Interceptor.ServiceToken,
		"user-id", publisherID,
	)

	// 调用rpc服务方法 执行核心业务
	rep, err := l.svcCtx.ArticleModel.UpdateArticle(ctx, &articlerpc.UpdateArticleRequest{
		ArticleID:   req.ArticleID,
		Title:       req.Title,
		Content:     req.Content,
		Tags:        req.Tags,
		Information: req.Information,
	})
	if err != nil {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "ERROR:" + err.Error()
		return resp, nil
	}
	if rep.Code != 0 {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = rep.Message
		return resp, nil
	}

	resp.Code = 0
	resp.Message = rep.Message
	resp.Status = http.StatusOK

	return resp, nil
}
