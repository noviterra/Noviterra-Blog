// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"net/http"

	"Noviterra/server/app/user/userapi/internal/svc"
	"Noviterra/server/app/user/userapi/internal/types"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/go-playground/validator/v10"
	"github.com/zeromicro/go-zero/core/logx"
)

type UnblacklistUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnblacklistUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnblacklistUserLogic {
	return &UnblacklistUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnblacklistUserLogic) UnblacklistUser(req *types.UnblacklistUserRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			resp.Message = TypeName[errs[0].Field()] + "格式错误！"
			resp.Status = http.StatusBadRequest
			resp.Code = 1
			return resp, nil
		}
	}

	// 通过鉴权中间件 从上下文中获取操作人员的ID与身份
	id := l.svcCtx.Middleware.GetUserIDFromContext(l.ctx)
	if id == "" {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "账号为空！"
		return resp, nil
	}
	role := l.svcCtx.Middleware.GetUserRoleFromContext(l.ctx)
	if role == -1 {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "未登录，身份验证失败！"
		return resp, nil
	}
	if role < 1 {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "权限不足，无法解封用户！"
		return resp, nil
	}

	//
	rep, err := l.svcCtx.UserModel.UnBlackUser(l.ctx, &userrpc.UnBlackUserRequest{
		ID:       req.ID,
		Operator: id,
	})
	if err != nil {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "ERROR:" + err.Error()
		return resp, nil
	}
	if rep.Code != 0 {
		resp.Code = 1
		resp.Status = http.StatusBadRequest
		resp.Message = rep.Message
		return resp, nil
	}

	resp.Code = 0
	resp.Message = rep.Message
	resp.Status = http.StatusOK

	return resp, nil
}
