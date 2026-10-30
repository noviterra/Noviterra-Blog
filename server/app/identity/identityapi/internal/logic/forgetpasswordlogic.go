// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/identity/identityapi/internal/svc"
	"Noviterra/server/app/identity/identityapi/internal/types"
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/utils/errcode"
	"context"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type forgetPasswordIdentityModel interface {
	ForgetPassword(ctx context.Context, in *identityrpc.ForgetPasswordRequest, opts ...grpc.CallOption) (*identityrpc.Response, error)
}

type forgetPasswordServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
}
type ForgetPasswordLogic struct {
	ctx                     context.Context
	indetityModel           forgetPasswordIdentityModel
	serviceTokenInterceptor forgetPasswordServiceTokenInterceptor
}

func NewForgetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForgetPasswordLogic {
	return &ForgetPasswordLogic{
		ctx:                     ctx,
		indetityModel:           svcCtx.IdentityModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *ForgetPasswordLogic) ForgetPassword(req *types.ForgetPasswordRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", req.ID,
	)

	rep, err := l.indetityModel.ForgetPassword(ctx, &identityrpc.ForgetPasswordRequest{
		ID:               req.ID,
		Email:            req.Email,
		VerificationCode: req.VerificationCode,
		NewPassword:      req.NewPassword,
		VerifPassword:    req.VerifPassowrd,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Code != 0 {
		return failResponse(resp, rep.Message, errcode.InvalidParam, int(rep.Code))
	}

	return successResponse(resp, "密码重置成功！")
}

func successResponse(resp *types.Response, msg string) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}

	resp.Code = errcode.Success
	resp.Message = msg
	resp.Status = errcode.RequestSuccess

	return resp, nil
}

func failResponse(resp *types.Response, msg string, status, code int) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}

	resp.Code = code
	resp.Status = status
	resp.Message = msg

	return resp, nil
}

func failResponseWithErr(resp *types.Response, err error, status, code int) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}

	resp.Code = code
	resp.Status = status
	resp.Message = "服务失败：" + err.Error()

	return resp, nil
}
