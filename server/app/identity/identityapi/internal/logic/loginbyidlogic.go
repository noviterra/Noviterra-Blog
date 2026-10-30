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

type loginByIDIdentitytModel interface {
	LoginByID(ctx context.Context, in *identityrpc.LoginByIDRequest, opts ...grpc.CallOption) (*identityrpc.Response, error)
	GenerateToken(ctx context.Context, in *identityrpc.GenerateTokenRequest, opts ...grpc.CallOption) (*identityrpc.GenerateTokenResponse, error)
}

type loginByIDServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
	GetSkipTokenInConf() string
}
type LoginByIDLogic struct {
	ctx                     context.Context
	identityModel           loginByIDIdentitytModel
	serviceTokenInterceptor loginByIDServiceTokenInterceptor
}

func NewLoginByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByIDLogic {
	return &LoginByIDLogic{
		ctx:                     ctx,
		identityModel:           svcCtx.IdentityModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *LoginByIDLogic) LoginByID(req *types.LoginByIDRequest) (resp *types.LoginResponse, err error) {
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

	rep, err := l.identityModel.LoginByID(ctx, &identityrpc.LoginByIDRequest{
		ID:       req.ID,
		Password: req.Password,
	})

	if err != nil {
		return failLoginResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Code != 0 {
		return failLoginResponse(resp, rep.Message, int(rep.Status), int(rep.Code))
	}

	// 重置metadata上下文数据 删除skip-token 增加user-id
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		delete(md, "skip-token")
		md.Set("service-token", l.serviceTokenInterceptor.GetServiceTokenInConf())
		md.Set("user-id", req.ID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// 生产唯一凭证---Token
	rep2, err := l.identityModel.GenerateToken(ctx, &identityrpc.GenerateTokenRequest{
		ID: req.ID,
	})
	if err != nil {
		return failLoginResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep2.Resp.Code != 0 {
		return failLoginResponse(resp, rep2.Resp.Message, int(rep2.Resp.Status), int(rep2.Resp.Code))
	}

	return successLoginResponse(resp, rep2)
}
