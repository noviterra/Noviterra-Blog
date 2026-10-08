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

type ChangePasswordByIDLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePasswordByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordByIDLogic {
	return &ChangePasswordByIDLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePasswordByIDLogic) ChangePasswordByID(req *types.ChangePasswordByIDRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			resp.Code = 1
			resp.Status = http.StatusBadRequest
			resp.Message = TypeName[errs[0].Field()] + "格式错误！"
			return resp, nil
		}
	}

	id := l.svcCtx.Middleware.GetUserIDFromContext(l.ctx)
	if id == "" {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "用户ID为空！"
		return resp, nil
	}

	rep, err := l.svcCtx.UserModel.ChangeSelfPasswordByID(l.ctx, &userrpc.ChangePasswordRequest{
		ID:          id,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "ERROR:" + err.Error()
		return resp, nil
	}

	if rep.Code != 0 {
		resp.Message = rep.Message
		resp.Code = 1
		resp.Status = http.StatusBadRequest
		return resp, nil
	}

	resp.Message = rep.Message
	resp.Code = 0
	resp.Status = http.StatusOK

	return resp, nil
}
