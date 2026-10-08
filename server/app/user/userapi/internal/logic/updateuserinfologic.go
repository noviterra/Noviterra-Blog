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

type UpdateUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserInfoLogic) UpdateUserInfo(req *types.UpdateUserInfoRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			resp.Status = http.StatusBadRequest
			resp.Code = 1
			resp.Message = TypeName[errs[0].Field()] + "格式错误！"
			return resp, nil
		}
	}

	// 从上下文中获取用户ID 经过了鉴权中间件获取
	ID := l.svcCtx.Middleware.GetUserIDFromContext(l.ctx)
	if ID == "" {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "错误的请求，用户ID为空！"
	}

	rep, err := l.svcCtx.UserModel.UpdateSelfInfo(l.ctx, &userrpc.UpdateUserInfoRequest{
		ID:               ID,
		UserName:         req.Username,
		Avatar:           req.Avatar,
		Age:              int64(req.Age),
		Gender:           int64(req.Gender),
		SelfIntroduction: req.SelfIntroduction,
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

	resp.Status = http.StatusOK
	resp.Code = 0
	resp.Message = rep.Message

	return resp, nil
}
