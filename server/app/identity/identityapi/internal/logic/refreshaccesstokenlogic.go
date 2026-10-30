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

type refreshTokenIdentityModel interface {
	RefreshAccessToken(ctx context.Context, in *identityrpc.RefreshAccessTokenARequest, opts ...grpc.CallOption) (*identityrpc.RefreshAccessTokenResponse, error)
}

type refreshAccessTokenServiceTokenIntercptor interface {
	GetServiceTokenInConf() string
	GetSkipTokenInConf() string
}
type RefreshAccessTokenLogic struct {
	ctx                     context.Context
	identityModel           refreshTokenIdentityModel
	serviceTokenInterceptor refreshAccessTokenServiceTokenIntercptor
}

func NewRefreshAccessTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshAccessTokenLogic {
	return &RefreshAccessTokenLogic{
		ctx:                     ctx,
		identityModel:           svcCtx.IdentityModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *RefreshAccessTokenLogic) RefreshAccessToken(req *types.RefreshAccessTokenRequest) (resp *types.RefreshAccessTokenResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.RefreshAccessTokenResponse)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return l.failRefreshAccessTokenResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return l.failRefreshAccessTokenResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"skip-token", l.serviceTokenInterceptor.GetSkipTokenInConf(),
	)

	// 根据RefreshToken刷新AccessToken RefreshToken也刷新
	rep, err := l.identityModel.RefreshAccessToken(ctx, &identityrpc.RefreshAccessTokenARequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return l.failRefreshAccessTokenResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return l.failRefreshAccessTokenResponse(resp, rep.Resp.Message, int(rep.Resp.Status), int(rep.Resp.Code))
	}

	return l.successRefreshAccessTokenResponse(resp, rep)
}

func (l *RefreshAccessTokenLogic) successRefreshAccessTokenResponse(resp *types.RefreshAccessTokenResponse,
	rep *identityrpc.RefreshAccessTokenResponse) (*types.RefreshAccessTokenResponse, error) {
	if resp == nil {
		resp = new(types.RefreshAccessTokenResponse)
	}

	resp.Token.RefreshToken = rep.Token.RefreshToken
	resp.Token.AccessToken = rep.Token.AccessToken

	resp.Resp = types.Response{
		Status:  errcode.RequestSuccess,
		Code:    errcode.Success,
		Message: "Token刷新成功！",
	}

	return resp, nil
}

func (l *RefreshAccessTokenLogic) failRefreshAccessTokenResponse(resp *types.RefreshAccessTokenResponse,
	msg string, status, code int) (*types.RefreshAccessTokenResponse, error) {
	if resp == nil {
		resp = new(types.RefreshAccessTokenResponse)
	}

	resp.Token = types.TokenData{}

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: msg,
	}

	return resp, nil
}

func (l *RefreshAccessTokenLogic) failRefreshAccessTokenResponseWithErr(resp *types.RefreshAccessTokenResponse,
	err error, status, code int) (*types.RefreshAccessTokenResponse, error) {
	if resp == nil {
		resp = new(types.RefreshAccessTokenResponse)
	}

	resp.Token = types.TokenData{}

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "服务失败：" + err.Error(),
	}

	return resp, nil
}
