package logic

import (
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"
)

type unBlackUserIdentityRepository interface {
	CheckIDExists(id string) (bool, error)
	UpdateUsersAccount(userAccount model.UserAccount) error
	GetUserRole(id string) (int, error)
}

type UnBlackUserLogic struct {
	ctx                context.Context
	identityRepository unBlackUserIdentityRepository
}

func NewUnBlackUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnBlackUserLogic {
	return &UnBlackUserLogic{
		ctx:                ctx,
		identityRepository: svcCtx.IdentityRepository,
	}
}

func (l *UnBlackUserLogic) UnBlackUser(in *userrpc.UnBlackUserRequest) (*userrpc.Response, error) {
	operatorRole, err := l.identityRepository.GetUserRole(in.Operator)
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}
	if operatorRole != 1 {
		return &userrpc.Response{
			Code:    1,
			Message: "操作者权限不足！",
		}, nil
	}

	idExist, err := l.identityRepository.CheckIDExists(in.ID)
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}
	if !idExist {
		return &userrpc.Response{
			Code:    1,
			Message: "要解封的用户ID不存在！",
		}, nil
	}

	// 修改数据库字段
	// todo： q1 解封字段无法更新
	err = l.identityRepository.UpdateUsersAccount(model.UserAccount{
		ID:     in.ID,
		Status: 0,
	})
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}

	return &userrpc.Response{
		Code:    0,
		Message: "用户账号解封完成！",
	}, nil
}
