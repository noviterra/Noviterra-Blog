package logic

import (
	"context"

	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSelfInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSelfInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSelfInfoLogic {
	return &GetSelfInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户个人操作
func (l *GetSelfInfoLogic) GetSelfInfo(in *userrpc.GetUserInfoRequest) (*userrpc.UserInfoResponse, error) {
	// todo: add your logic here and delete this line
	// 查看缓存中是否有用户数据
	cacheUser, err := l.svcCtx.RDB.HGetUserInfo(l.ctx, in.ID)
	if err != nil {
		return &userrpc.UserInfoResponse{
			Code:    1,
			Message: "缓存异常，请联系管理员",
		}, err
	}
	// 缓存命中
	if cacheUser != nil {
		return &userrpc.UserInfoResponse{
			Code:    0,
			Message: "缓存命中，得到用户信息！",
			Data:    cacheUser,
		}, nil
	}

	// 缓存未命中 查库
	userInfo, err := l.svcCtx.DB.GetUserInfoByID(in.ID)
	if err != nil {
		return &userrpc.UserInfoResponse{
			Code:    1,
			Message: "数据库异常，请联系管理员！",
		}, err
	}
	if userInfo == nil {
		return &userrpc.UserInfoResponse{
			Code:    1,
			Message: "无账号相关信息，请联系管理员！",
		}, nil
	}

	// 将用户信息写入缓存中
	err = l.svcCtx.RDB.HSetUserInfo(l.ctx, userInfo)
	if err != nil {
		l.svcCtx.Logger.Error("用户信息写入缓存失败！")
	}

	return &userrpc.UserInfoResponse{
		Code:    0,
		Message: "检索到用户信息！",
		Data: &userrpc.UserInfoResponse_Data{
			ID:               userInfo.ID,
			Email:            userInfo.Email,
			UserName:         userInfo.Name,
			Avatar:           userInfo.Avatar,
			Age:              int64(userInfo.Age),
			Gender:           int64(userInfo.Gender),
			SelfIntroduction: userInfo.SelfIntroduction,
			CreateAt:         userInfo.CreatedAt.Unix(),
			UpdatedAt:        userInfo.UpdatedAt.Unix(),
		},
	}, nil
}
