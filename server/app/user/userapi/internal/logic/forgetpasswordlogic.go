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

type ForgetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewForgetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForgetPasswordLogic {
	return &ForgetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ForgetPasswordLogic) ForgetPassword(req *types.ForgetPasswordRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			resp.Message = TypeName[errs[0].Field()] + "格式错误！"
			resp.Code = 1
			resp.Status = http.StatusBadRequest
		}
		return resp, nil
	}

	// 一键操作
	rep, err := l.svcCtx.UserModel.ForgetPassword(l.ctx, &userrpc.ForgetPasswordRequest{
		ID:               req.ID,
		Email:            req.Email,
		VerificationCode: req.VerificationCode,
		NewPassword:      req.NewPassword,
		VerifPassword:    req.VerifPassowrd,
	})
	if err != nil {
		resp.Code = 1
		resp.Status = http.StatusBadRequest
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
	resp.Status = http.StatusOK
	resp.Message = rep.Message

	return resp, nil
}
