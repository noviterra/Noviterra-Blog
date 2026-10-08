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

type LoginByIDLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByIDLogic {
	return &LoginByIDLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginByIDLogic) LoginByID(req *types.LoginByIDRequest) (resp *types.LoginResponse, err error) {
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

	rep, err := l.svcCtx.UserModel.LoginByID(l.ctx, &userrpc.LoginByIDRequest{
		ID:       req.ID,
		Password: req.Password,
	})

	if err != nil {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = "ERROR:" + err.Error()
		return resp, nil
	}
	if rep.Code != 0 {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = rep.Message
		return resp, nil
	}

	// 生产唯一凭证---Token
	rep2, err := l.svcCtx.UserModel.GenerateToken(l.ctx, &userrpc.GenerateTokenRequest{
		ID: req.ID,
	})
	if err != nil {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = "ERROR:" + err.Error()
		return resp, nil
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
