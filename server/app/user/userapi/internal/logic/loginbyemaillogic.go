// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"fmt"
	"net/http"

	"Noviterra/server/app/user/userapi/internal/svc"
	"Noviterra/server/app/user/userapi/internal/types"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/go-playground/validator/v10"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginByEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByEmailLogic {
	return &LoginByEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginByEmailLogic) LoginByEmail(req *types.LoginByEmailRequest) (resp *types.LoginResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.LoginResponse)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			resp.Resp.Status = http.StatusBadRequest
			resp.Resp.Code = 1
			resp.Resp.Message = TypeName[errs[0].Field()] + "格式错误！"
		}
		return resp, nil
	}

	rep, err := l.svcCtx.UserModel.LoginByEmail(l.ctx, &userrpc.LoginByEmailRequest{
		Email:            req.Email,
		VerificationCode: req.VerificationCode,
	})

	if err != nil {
		resp.Resp.Code = http.StatusBadRequest
		resp.Resp.Message = "ERROR:" + err.Error()
		return resp, nil
	}
	if rep.Code != 0 {
		resp.Resp.Code = http.StatusBadRequest
		resp.Resp.Message = rep.Message
		return resp, nil
	}

	// 生成唯一Token
	rep2, err := l.svcCtx.UserModel.GenerateToken(l.ctx, &userrpc.GenerateTokenRequest{
		ID: rep.ID,
	})
	if err != nil {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = fmt.Sprintf("Token生成失败！ERROR：%s", err.Error())
	}
	if rep2.Code != 0 {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = rep2.Message
		return resp, nil
	}

	resp.Resp.Status = http.StatusOK
	resp.Resp.Code = 0
	resp.Resp.Message = "登录成功！"
	resp.AccessToken = rep2.AccessToken
	resp.RefreshToken = rep2.RefreshToken
	return resp, nil
}
