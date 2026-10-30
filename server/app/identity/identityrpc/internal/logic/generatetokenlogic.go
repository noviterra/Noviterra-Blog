package logic

import (
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"
	"fmt"
	"time"

	"Noviterra/server/app/identity/identityrpc/internal/svc"
)

type generateTokenIdentityRespository interface {
	InsertUserToken(userToken model.UserRefreshToken) error
}

type generateTokenIdentityAuthentication interface {
	GenerateDoubleToken(id string) (string, string, error)
	ReturnRefreshTokenExpire() int64
}

type generateTokenIdentityResponse interface {
	SuccessGenerateTokenResponse(resp *identityrpc.GenerateTokenResponse, accessToken, refreshToken string) (*identityrpc.GenerateTokenResponse, error)
	FailGenerateTokenResponse(resp *identityrpc.GenerateTokenResponse, msg string, status int32, code int64) (*identityrpc.GenerateTokenResponse, error)
	FailGenerateTokenResponseWithErr(reso *identityrpc.GenerateTokenResponse, err error) (*identityrpc.GenerateTokenResponse, error)
}

type generateTokenCtxData interface {
	GetUserIDFromContext(ctx context.Context) string
}
type GenerateTokenLogic struct {
	ctx                 context.Context
	identityRespository generateTokenIdentityRespository
	authentication      generateTokenIdentityAuthentication
	identityResponse    generateTokenIdentityResponse
	ctxdata             generateTokenCtxData
}

func NewGenerateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateTokenLogic {
	return &GenerateTokenLogic{
		ctx:                 ctx,
		identityRespository: svcCtx.IdentityRepository,
		authentication:      svcCtx.Authentication,
		identityResponse:    svcCtx.IdentityResponse,
		ctxdata:             svcCtx.Ctxdata,
	}
}

// 生成用户Token
func (l *GenerateTokenLogic) GenerateToken(in *identityrpc.GenerateTokenRequest) (*identityrpc.GenerateTokenResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(identityrpc.GenerateTokenResponse)

	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		return l.identityResponse.FailGenerateTokenResponse(resp, "用户ID校验失败！", errcode.InvalidParam, errcode.Failed)
	}
	if id != in.ID {
		fmt.Println("id", id)
		fmt.Println("in.ID", in.ID)
		return l.identityResponse.FailGenerateTokenResponse(resp, "用户操作异常！", errcode.InvalidParam, errcode.Failed)
	}

	accessToken, refreshToken, err := l.authentication.GenerateDoubleToken(in.ID)
	if err != nil {
		return l.identityResponse.FailGenerateTokenResponse(resp, "Token生成失败！", errcode.InvalidParam, errcode.Failed)
	}

	err = l.identityRespository.InsertUserToken(model.UserRefreshToken{
		ID:           in.ID,
		RefreshToken: refreshToken,
		CreateAt:     time.Now(),
		UpdatedAt:    time.Now(),
		ExpireAt:     time.Now().Add(time.Duration(l.authentication.ReturnRefreshTokenExpire()) * time.Second),
	})
	if err != nil {
		return l.identityResponse.FailGenerateTokenResponse(resp, "Token存储失败！", errcode.InvalidParam, errcode.Failed)
	}

	return l.identityResponse.SuccessGenerateTokenResponse(resp, accessToken, refreshToken)
}
