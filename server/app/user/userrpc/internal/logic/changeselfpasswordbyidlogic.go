package logic

import (
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"
)

type changPwdByIDUserRepository interface {
	GetUserPassword(id string) (string, error)
}

type changePwdByIDIdentityRepository interface {
	UpdateUsersAccount(ctx context.Context,
		userAccount model.UserAccount) error
}

type changePwdByIDCrypt interface {
	CheckPasswords(password, hashPassword string) (bool, error)
	HashPassword(password string) (string, error)
}

type changePwdByIDIdentityValidate interface {
	CheckPasswordFormat(password string) bool
}
type ChangeSelfPasswordByIDLogic struct {
	ctx                context.Context
	userRepository     changPwdByIDUserRepository
	identityRepository changePwdByIDIdentityRepository
	crypt              changePwdByIDCrypt
	identityValidate   changePwdByIDIdentityValidate
}

func NewChangeSelfPasswordByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeSelfPasswordByIDLogic {
	return &ChangeSelfPasswordByIDLogic{
		ctx:                ctx,
		identityRepository: svcCtx.IdentityRepository,
		userRepository:     svcCtx.UserRepository,
		crypt:              svcCtx.Crypt,
		identityValidate:   svcCtx.IdentityValidate,
	}
}

func (l *ChangeSelfPasswordByIDLogic) ChangeSelfPasswordByID(in *userrpc.ChangePasswordRequest) (*userrpc.Response, error) {
	// todo: add your logic here and delete this line
	// 根据用户ID返回Password 为空则表示账号不存在
	hashPwd, err := l.userRepository.GetUserPassword(in.ID)
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}
	if hashPwd == "" {
		return &userrpc.Response{
			Code:    1,
			Message: "用户ID不存在！",
		}, nil
	}

	samePwd, err := l.crypt.CheckPasswords(in.OldPassword, hashPwd)
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}
	if !samePwd {
		return &userrpc.Response{
			Code:    1,
			Message: "原始密码输入错误！",
		}, nil
	}

	passwordAvailable := l.identityValidate.CheckPasswordFormat(in.NewPassword)
	if !passwordAvailable {
		return &userrpc.Response{
			Code:    1,
			Message: "密码格式错误，请检查后重试！",
		}, nil
	}

	// 密码加密
	hashPwd1, err := l.crypt.HashPassword(in.NewPassword)
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}

	err = l.identityRepository.UpdateUsersAccount(l.ctx, model.UserAccount{
		ID:       in.ID,
		Password: hashPwd1,
	})
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}

	return &userrpc.Response{
		Code:    0,
		Message: "密码修改成功！",
	}, nil
}
