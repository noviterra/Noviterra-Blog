package logic

import (
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/jwtauth"
	"context"
	"time"

	"Noviterra/server/app/identity/identityrpc/internal/svc"
)

type (
	refreshAccessTokenAuthentication interface {
		ParseRefreshToken(tokenStr string) (*jwtauth.UserClaims, error)
		GenerateDoubleToken(id string) (string, string, error)
		ReturnRefreshTokenExpire() int64
	}

	refreshAccessTokenIdentityRepository interface {
		CheckIDExists(ctx context.Context, id string) (bool, error)
		UpdateUserToken(ctx context.Context,
			userToken model.UserRefreshToken) error
	}

	refreshAccessTokenIdentityResponse interface {
		SuccessRefreshAccessTokenResponse(resp *identityrpc.RefreshAccessTokenResponse,
			accessToken, refreshToken string) (*identityrpc.RefreshAccessTokenResponse, error)
		FailRefreshAccessTokenResponse(resp *identityrpc.RefreshAccessTokenResponse,
			msg string, status int32, code int64) (*identityrpc.RefreshAccessTokenResponse, error)
		FailRefreshAccessTokenResponseWithErr(resp *identityrpc.RefreshAccessTokenResponse,
			err error) (*identityrpc.RefreshAccessTokenResponse, error)
	}
)

type RefreshAccessTokenLogic struct {
	ctx                context.Context
	identityRepository refreshAccessTokenIdentityRepository
	authentication     refreshAccessTokenAuthentication
	identityResponse   refreshAccessTokenIdentityResponse
}

func NewRefreshAccessTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshAccessTokenLogic {
	return &RefreshAccessTokenLogic{
		ctx:                ctx,
		identityRepository: svcCtx.IdentityRepository,
		authentication:     svcCtx.Authentication,
		identityResponse:   svcCtx.IdentityResponse,
	}
}

// 刷新用户AccessToken
func (l *RefreshAccessTokenLogic) RefreshAccessToken(in *identityrpc.RefreshAccessTokenARequest) (*identityrpc.RefreshAccessTokenResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(identityrpc.RefreshAccessTokenResponse)
	// 解析RefreshToken
	claim, err := l.authentication.ParseRefreshToken(in.RefreshToken)
	if err != nil {
		return l.identityResponse.FailRefreshAccessTokenResponse(resp, "账号信息校验失败！", errcode.InvalidParam, errcode.Failed)
	}

	idExist, err := l.identityRepository.CheckIDExists(l.ctx, claim.ID)
	if err != nil {
		return l.identityResponse.FailRefreshAccessTokenResponse(resp, "数据库操作异常！", errcode.InvalidParam, errcode.Failed)

	}
	if !idExist {
		return l.identityResponse.FailRefreshAccessTokenResponse(resp, "用户信息异常！", errcode.InvalidParam, errcode.Failed)
	}

	// 生成Token
	accessToken, refreshToken, err := l.authentication.GenerateDoubleToken(claim.ID)
	if err != nil {
		return l.identityResponse.FailRefreshAccessTokenResponse(resp, "Token刷新失败，请重新登录！", errcode.InvalidParam, errcode.Failed)
	}

	// Token存储
	err = l.identityRepository.UpdateUserToken(l.ctx, model.UserRefreshToken{
		ID:           claim.ID,
		RefreshToken: refreshToken,
		UpdatedAt:    time.Now(),
		ExpireAt:     time.Now().Add(time.Duration(l.authentication.ReturnRefreshTokenExpire()) * time.Second),
	})
	if err != nil {
		return l.identityResponse.FailRefreshAccessTokenResponse(resp, "数据库操作异常！", errcode.InvalidParam, errcode.Failed)
	}

	return l.identityResponse.SuccessRefreshAccessTokenResponse(resp, accessToken, refreshToken)
}
