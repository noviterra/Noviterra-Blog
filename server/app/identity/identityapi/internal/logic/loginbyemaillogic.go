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

type loginByEmailIndetityModel interface {
	LoginByEmail(ctx context.Context, in *identityrpc.LoginByEmailRequest, opts ...grpc.CallOption) (*identityrpc.LoginByEmailResponse, error)
	GenerateToken(ctx context.Context, in *identityrpc.GenerateTokenRequest, opts ...grpc.CallOption) (*identityrpc.GenerateTokenResponse, error)
}

type loginByEmailServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
	GetSkipTokenInConf() string
}
type LoginByEmailLogic struct {
	ctx                     context.Context
	identityModel           loginByEmailIndetityModel
	serviceTokenInterceptor loginByEmailServiceTokenInterceptor
}

func NewLoginByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByEmailLogic {
	return &LoginByEmailLogic{
		ctx:                     ctx,
		identityModel:           svcCtx.IdentityModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *LoginByEmailLogic) LoginByEmail(req *types.LoginByEmailRequest) (resp *types.LoginResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.LoginResponse)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failLoginResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failLoginResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"skip-token", l.serviceTokenInterceptor.GetSkipTokenInConf(),
	)

	rep, err := l.identityModel.LoginByEmail(ctx, &identityrpc.LoginByEmailRequest{
		Email:            req.Email,
		VerificationCode: req.VerificationCode,
	})

	if err != nil {
		return failLoginResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return failLoginResponse(resp, rep.Resp.Message, int(rep.Resp.Status), int(rep.Resp.Code))
	}
	// 将用户ID传入到metadata中 前提是先将skip-token删除
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		delete(md, "skip-token")
		md.Set("user-id", rep.ID)
		md.Set("service-token", l.serviceTokenInterceptor.GetServiceTokenInConf())
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	// 生成唯一Token
	rep2, err := l.identityModel.GenerateToken(ctx, &identityrpc.GenerateTokenRequest{
		ID: rep.ID,
	})
	if err != nil {
		return failLoginResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep2.Resp.Code != 0 {
		return failLoginResponse(resp, rep2.Resp.Message, int(rep2.Resp.Status), int(rep2.Resp.Code))
	}

	return successLoginResponse(resp, rep2)
}

func successLoginResponse(resp *types.LoginResponse, tokenRep *identityrpc.GenerateTokenResponse) (*types.LoginResponse, error) {
	if resp == nil {
		resp = new(types.LoginResponse)
	}

	resp.Token.RefreshToken = tokenRep.Token.RefreshToken
	resp.Token.AccessToken = tokenRep.Token.AccessToken

	resp.Resp = types.Response{
		Status:  errcode.RequestSuccess,
		Code:    errcode.Success,
		Message: "登录成功！",
	}

	return resp, nil
}

func failLoginResponse(resp *types.LoginResponse, msg string, status, code int) (*types.LoginResponse, error) {
	if resp == nil {
		resp = new(types.LoginResponse)
	}

	resp.Token = types.TokenData{}

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: msg,
	}

	return resp, nil
}

func failLoginResponseWithErr(resp *types.LoginResponse, err error, status, code int) (*types.LoginResponse, error) {
	if resp == nil {
		resp = new(types.LoginResponse)
	}

	resp.Token = types.TokenData{}

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "服务失败：" + err.Error(),
	}

	return resp, nil
}
