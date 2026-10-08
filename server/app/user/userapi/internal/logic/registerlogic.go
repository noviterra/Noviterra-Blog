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

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	//fmt.Println("======server/app/user/api/registerlogic.go======")
	resp = new(types.Response)

	// binding tag验证
	if err := Validate.Struct(*req); err != nil {
		// 解析错误类型
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			resp.Code = 1
			resp.Status = http.StatusBadRequest
			resp.Message = TypeName[errs[0].Field()] + "格式错误!"
		}
		return resp, nil
	}

	rep, err := l.svcCtx.UserModel.Register(l.ctx, &userrpc.RegisterRequest{
		ID:               req.ID,
		Password:         req.Password,
		Email:            req.Email,
		VerifPassword:    req.VerifPassword,
		VerificationCode: req.VerificationCode,
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
	resp.Message = fmt.Sprintf("账号注册成功：ID:%s|Email:%s", req.ID, req.Email)

	return resp, nil
}
