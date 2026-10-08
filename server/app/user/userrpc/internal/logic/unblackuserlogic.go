package logic

import (
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnBlackUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnBlackUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnBlackUserLogic {
	return &UnBlackUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnBlackUserLogic) UnBlackUser(in *userrpc.UnBlackUserRequest) (*userrpc.UnBlackUserResponse, error) {
	operatorRole, err := l.svcCtx.DB.GetUserRole(in.Operator)
	if err != nil {
		return &userrpc.UnBlackUserResponse{
			Code:    1,
			Message: "",
		}, err
	}
	if operatorRole != 1 {
		return &userrpc.UnBlackUserResponse{
			Code:    1,
			Message: "操作者权限不足！",
		}, nil
	}

	idExist, err := l.svcCtx.DB.CheckIDExists(in.ID)
	if err != nil {
		return &userrpc.UnBlackUserResponse{
			Code:    1,
			Message: "",
		}, err
	}
	if !idExist {
		return &userrpc.UnBlackUserResponse{
			Code:    1,
			Message: "要解封的用户ID不存在！",
		}, nil
	}

	// 修改数据库字段
	// todo： q1 解封字段无法更新
	err = l.svcCtx.DB.UpdateUsersAccount(model.UserAccount{
		ID:     in.ID,
		Status: 0,
	})
	if err != nil {
		return &userrpc.UnBlackUserResponse{
			Code:    1,
			Message: "",
		}, err
	}

	return &userrpc.UnBlackUserResponse{
		Code:    0,
		Message: "用户账号解封完成！",
	}, nil
}
