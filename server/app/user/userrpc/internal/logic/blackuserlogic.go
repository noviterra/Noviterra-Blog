package logic

import (
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BlackUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBlackUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlackUserLogic {
	return &BlackUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员对用户操作
func (l *BlackUserLogic) BlackUser(in *userrpc.BlackUserRequest) (*userrpc.BlackUserResponse, error) {
	// todo: add your logic here and delete this line
	// 校验操作者ID与用户ID的有效性
	// todo：绕过api层获取管理员账号后之间进入rpc层怎么解决？
	operatorRole, err := l.svcCtx.DB.GetUserRole(in.Operator)
	if err != nil {
		return &userrpc.BlackUserResponse{
			Code:    1,
			Message: "",
		}, err
	}
	if operatorRole != 1 {
		return &userrpc.BlackUserResponse{
			Code:    1,
			Message: "操作者权限不足！",
		}, nil
	}

	idExist, err := l.svcCtx.DB.CheckIDExists(in.ID)
	if err != nil {
		return &userrpc.BlackUserResponse{
			Code:    1,
			Message: "",
		}, err
	}
	if !idExist {
		return &userrpc.BlackUserResponse{
			Code:    1,
			Message: "要封禁的用户ID不存在！",
		}, nil
	}

	// 修改数据库字段
	err = l.svcCtx.DB.UpdateUsersAccount(model.UserAccount{
		ID:     in.ID,
		Status: 1,
	})
	if err != nil {
		return &userrpc.BlackUserResponse{
			Code:    1,
			Message: "",
		}, err
	}

	return &userrpc.BlackUserResponse{
		Code:    0,
		Message: "用户账号封禁完成！",
	}, nil
}
