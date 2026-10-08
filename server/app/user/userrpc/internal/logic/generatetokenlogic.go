package logic

import (
	"context"
	"time"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateTokenLogic {
	return &GenerateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GenerateTokenLogic) GenerateToken(in *userrpc.GenerateTokenRequest) (*userrpc.GenerateTokenResponse, error) {
	// todo: add your logic here and delete this line
	accessToken, refreshToken, err := l.svcCtx.Authentication.GenerateDoubleToken(in.ID)
	if err != nil {
		return &userrpc.GenerateTokenResponse{
			Code:    1,
			Message: "Token生成失败！",
		}, err
	}

	err = l.svcCtx.DB.InsertUserToken(model.UserRefreshToken{
		ID:           in.ID,
		RefreshToken: refreshToken,
		CreateAt:     time.Now(),
		UpdatedAt:    time.Now(),
		ExpireAt:     time.Now().Add(time.Duration(l.svcCtx.Authentication.RefreshExpire) * time.Second),
	})
	if err != nil {
		return &userrpc.GenerateTokenResponse{
			Code:    1,
			Message: "Token存储失败！",
		}, err
	}

	return &userrpc.GenerateTokenResponse{
		Code:         0,
		Message:      "Token生成成功！",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
