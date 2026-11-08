package logic

import (
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"
)

type blackUserIdentityRepository interface {
	CheckIDExists(ctx context.Context, id string) (bool, error)
	UpdateUsersAccount(ctx context.Context,
		userAccount model.UserAccount) error
	GetUserRole(ctx context.Context, id string) (int, error)
}

type BlackUserLogic struct {
	ctx                context.Context
	identityRepository blackUserIdentityRepository
}

func NewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlackUserLogic {
	return &BlackUserLogic{
		ctx:                ctx,
		identityRepository: svcCtx.IdentityRepository,
	}
}

// 管理员对用户操作
func (l *BlackUserLogic) BlackUser(in *userrpc.BlackUserRequest) (*userrpc.Response, error) {
	// todo: add your logic here and delete this line
	// 校验操作者ID与用户ID的有效性
	operatorRole, err := l.identityRepository.GetUserRole(l.ctx, in.Operator)
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

	idExist, err := l.identityRepository.CheckIDExists(l.ctx, in.ID)
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}
	if !idExist {
		return &userrpc.Response{
			Code:    1,
			Message: "要封禁的用户ID不存在！",
		}, nil
	}

	// 修改数据库字段
	err = l.identityRepository.UpdateUsersAccount(l.ctx, model.UserAccount{
		ID:     in.ID,
		Status: 1,
	})
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}

	return &userrpc.Response{
		Code:    0,
		Message: "用户账号封禁完成！",
	}, nil
}
