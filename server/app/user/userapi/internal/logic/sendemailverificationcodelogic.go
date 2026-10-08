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

type SendEmailVerificationCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendEmailVerificationCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendEmailVerificationCodeLogic {
	return &SendEmailVerificationCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendEmailVerificationCodeLogic) SendEmailVerificationCode(req *types.SendEmailVerificationCodeRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			resp.Status = http.StatusBadRequest
			resp.Code = 1
			resp.Message = TypeName[errs[0].Field()] + "格式错误!"
		}
		return resp, nil
	}

	rep, err := l.svcCtx.UserModel.SendVerificationCode(l.ctx, &userrpc.SendVerificationCodeRequest{
		Email: req.Email,
		Type:  req.Type,
	})

	if err != nil { // 有err时 rep为nil （go-zero特性）
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "ERROR:" + err.Error()
		return resp, nil
	}
	if rep.Code != 0 { // 无err但是业务失败
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = rep.Message
		return resp, nil
	}

	// 业务正常结束
	resp.Status = http.StatusOK
	resp.Code = 0
	resp.Message = fmt.Sprintf("验证码发送成功：Email:%s|Type:%s", req.Email, req.Type)

	return resp, nil
}
