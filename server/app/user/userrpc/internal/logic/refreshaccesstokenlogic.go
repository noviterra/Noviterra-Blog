package logic

import (
	"context"
	"time"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshAccessTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshAccessTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshAccessTokenLogic {
	return &RefreshAccessTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshAccessTokenLogic) RefreshAccessToken(in *userrpc.RefreshAccessTokenARequest) (*userrpc.RefreshAccessTokenResponse, error) {
	// todo: add your logic here and delete this line
	// 解析RefreshToken
	claim, err := l.svcCtx.Authentication.ParseRefreshToken(in.RefreshToken)
	if err != nil {
		return &userrpc.RefreshAccessTokenResponse{
			Code:    1,
			Message: "Token刷新失败！",
		}, err
	}

	idExist, err := l.svcCtx.DB.CheckIDExists(claim.ID)
	if err != nil {
		return &userrpc.RefreshAccessTokenResponse{
			Code:    1,
			Message: "数据库校验异常！",
		}, err
	}
	if !idExist {
		return &userrpc.RefreshAccessTokenResponse{
			Code:    1,
			Message: "无效的用户信息！",
		}, nil
	}

	// 生成Token
	accessToken, refreshToken, err := l.svcCtx.Authentication.GenerateDoubleToken(claim.ID)
	if err != nil {
		return &userrpc.RefreshAccessTokenResponse{
			Code:    1,
			Message: "Token刷新失败！",
		}, err
	}

	// Token存储
	err = l.svcCtx.DB.UpdateUserToken(model.UserRefreshToken{
		ID:           claim.ID,
		RefreshToken: refreshToken,
		UpdatedAt:    time.Now(),
		// ExpireAt:  time.Now().Add(time.Duration(l.svcCtx.Authentication.RefreshExpire) * time.Second),  // 过期时间不重置
	})
	if err != nil {
		return &userrpc.RefreshAccessTokenResponse{
			Code:    1,
			Message: "Token存储失败！",
		}, err
	}

	return &userrpc.RefreshAccessTokenResponse{
		Code:         0,
		Message:      "Token刷新成功！",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
