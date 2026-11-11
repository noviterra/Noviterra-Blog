package logic

import (
	"context"
	"time"

	"Noviterra/server/app/identity/identityrpc/internal/svc"
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"

	"go.uber.org/zap"
)

type (
	generateTokenIdentityRespository interface {
		InsertUserToken(ctx context.Context,
			userToken model.UserRefreshToken) error
	}
	generateTokenIdentityAuthentication interface {
		GenerateDoubleToken(id string) (string, string, error)
		ReturnRefreshTokenExpire() int64
	}
	generateTokenIdentityResponse interface {
		SuccessGenerateTokenResponse(resp *identityrpc.GenerateTokenResponse, accessToken, refreshToken string) (*identityrpc.GenerateTokenResponse, error)
		FailGenerateTokenResponse(resp *identityrpc.GenerateTokenResponse, msg string, status int32, code int64) (*identityrpc.GenerateTokenResponse, error)
		FailGenerateTokenResponseWithErr(reso *identityrpc.GenerateTokenResponse, err error) (*identityrpc.GenerateTokenResponse, error)
	}
	generateTokenCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}
	generateTokenLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type GenerateTokenLogic struct {
	ctx                 context.Context
	identityRespository generateTokenIdentityRespository
	authentication      generateTokenIdentityAuthentication
	identityResponse    generateTokenIdentityResponse
	ctxdata             generateTokenCtxData
	logger              generateTokenLogger
}

func NewGenerateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateTokenLogic {
	return &GenerateTokenLogic{
		ctx:                 ctx,
		identityRespository: svcCtx.IdentityRepository,
		authentication:      svcCtx.Authentication,
		identityResponse:    svcCtx.IdentityResponse,
		ctxdata:             svcCtx.Ctxdata,
		logger:              svcCtx.Logger,
	}
}

// 生成用户Token
func (l *GenerateTokenLogic) GenerateToken(in *identityrpc.GenerateTokenRequest) (*identityrpc.GenerateTokenResponse, error) {
	resp := new(identityrpc.GenerateTokenResponse)

	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		l.logger.Error("上下文获取操作用户ID失败")
		return l.identityResponse.FailGenerateTokenResponse(resp, "用户ID校验失败！",
			errcode.InvalidParam, errcode.GetAuthInfoFail)
	}
	if id != in.ID {
		l.logger.Error("请求发起者ID与被操作ID不同",
			zap.Strings("user-id||operator-id", []string{id, in.ID}),
		)
		return l.identityResponse.FailGenerateTokenResponse(resp, "用户操作异常！",
			errcode.InvalidParam, errcode.IdentityOperationOutOfBounds)
	}

	accessToken, refreshToken, err := l.authentication.GenerateDoubleToken(in.ID)
	if err != nil {
		l.logger.Error("生成用户Token失败",
			zap.String("user-info", id),
			zap.Error(err),
		)
		return l.identityResponse.FailGenerateTokenResponse(resp, "Token生成失败！",
			errcode.InvalidParam, errcode.GenerateTokenFail)
	}

	err = l.identityRespository.InsertUserToken(l.ctx, model.UserRefreshToken{
		ID:           in.ID,
		RefreshToken: refreshToken,
		CreateAt:     time.Now(),
		UpdatedAt:    time.Now(),
		ExpireAt:     time.Now().Add(time.Duration(l.authentication.ReturnRefreshTokenExpire()) * time.Second),
	})
	if err != nil {
		l.logger.Error("数据库操作失败：存储用户Token失败",
			zap.String("user-info", id),
			zap.Error(err),
		)
		return l.identityResponse.FailGenerateTokenResponse(resp, "Token存储失败！",
			errcode.InvalidParam, errcode.SaveTokenFail)
	}

	l.logger.Info("生成用户Token成功",
		zap.String("user-info", id),
	)
	return l.identityResponse.SuccessGenerateTokenResponse(resp, accessToken, refreshToken)
}
