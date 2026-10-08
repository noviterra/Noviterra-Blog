package logic

import (
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeSelfPasswordByIDLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeSelfPasswordByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeSelfPasswordByIDLogic {
	return &ChangeSelfPasswordByIDLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChangeSelfPasswordByIDLogic) ChangeSelfPasswordByID(in *userrpc.ChangePasswordRequest) (*userrpc.ChangePasswordResponse, error) {
	// todo: add your logic here and delete this line
	// 根据用户ID返回Password 为空则表示账号不存在
	hashPwd, err := l.svcCtx.DB.GetUserPassword(in.ID)
	if err != nil {
		return &userrpc.ChangePasswordResponse{
			Code:    1,
			Message: "",
		}, err
	}
	if hashPwd == "" {
		return &userrpc.ChangePasswordResponse{
			Code:    1,
			Message: "用户ID不存在！",
		}, nil
	}

	samePwd, err := l.svcCtx.Crypt.CheckPasswords(in.OldPassword, hashPwd)
	if err != nil {
		return &userrpc.ChangePasswordResponse{
			Code:    1,
			Message: "",
		}, err
	}
	if !samePwd {
		return &userrpc.ChangePasswordResponse{
			Code:    1,
			Message: "原始密码输入错误！",
		}, nil
	}

	passwordAvailable := l.svcCtx.CommonUtils.CheckPasswordFormat(in.NewPassword)
	if !passwordAvailable {
		return &userrpc.ChangePasswordResponse{
			Code:    1,
			Message: "密码格式错误，请检查后重试！",
		}, nil
	}

	// 密码加密
	hashPwd1, err := l.svcCtx.Crypt.HashPassword(in.NewPassword)
	if err != nil {
		return &userrpc.ChangePasswordResponse{
			Code:    1,
			Message: "",
		}, err
	}

	err = l.svcCtx.DB.UpdateUsersAccount(model.UserAccount{
		ID:       in.ID,
		Password: hashPwd1,
	})
	if err != nil {
		return &userrpc.ChangePasswordResponse{
			Code:    1,
			Message: "",
		}, err
	}

	return &userrpc.ChangePasswordResponse{
		Code:    0,
		Message: "密码修改成功！",
	}, nil
}
