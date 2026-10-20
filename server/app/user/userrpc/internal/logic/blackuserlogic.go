package logic

import (
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"
)

type blackUserIdentityRepository interface {
	CheckIDExists(id string) (bool, error)
	UpdateUsersAccount(userAccount model.UserAccount) error
	GetUserRole(id string) (int, error)
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
	// todo：绕过api层获取管理员账号后之间进入rpc层怎么解决？ 设置拦截器 api层注入用户信息与服务凭证 作为rpc是否继续操作下去的密钥
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
			Message: "要封禁的用户ID不存在！",
		}, nil
	}

	// 修改数据库字段
	err = l.identityRepository.UpdateUsersAccount(model.UserAccount{
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
