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
)

type UpdateArticleCoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateArticleCoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateArticleCoverLogic {
	return &UpdateArticleCoverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateArticleCoverLogic) UpdateArticleCover(req *types.UpdateArticleCoverRequest) (resp *types.UpdateArticleCoverResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.UpdateArticleCoverResponse)

	if err := Validate.Struct(*req); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) && len(errs) > 0 {
			resp.Resp.Status = http.StatusBadRequest
			resp.Resp.Code = 1
			resp.Resp.Message = TypeName[errs[0].Field()] + "格式错误！"
		}
		return resp, nil
	}

	// 从上下文中获取用户ID
	publisherID := l.svcCtx.Middleware.GetUserIDFromContext(l.ctx)
	if publisherID == "" {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = "用户ID为空！"
		return resp, nil
	}

	// 获取用户身份权限
	role := l.svcCtx.Middleware.GetUserRoleFromContext(l.ctx)
	if role != 1 { // 1为管理员
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = "用户身份权限不足！"
		return resp, nil
	}

	// 调用rpc服务方法 执行核心业务
	rep, err := l.svcCtx.ArticleModel.UpdateArticleCover(l.ctx, &articlerpc.UpdateArticleCoverRequest{
		NewCover:    req.NewCover,
		PublisherID: publisherID,
		ArticleID:   req.ArticleID,
	})
	if err != nil {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = "ERROR:" + err.Error()
		return resp, nil
	}
	if rep.Resp.Code != 0 {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = rep.Resp.Message
		return resp, nil
	}

	resp.Resp.Code = 0
	resp.Resp.Message = rep.Resp.Message
	resp.Resp.Status = http.StatusOK

	resp.Cover = rep.Cover

	return resp, nil
}
