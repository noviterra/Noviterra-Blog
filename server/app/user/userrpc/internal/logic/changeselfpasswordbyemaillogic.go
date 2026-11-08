package logic

import (
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"
)

type changePwdByEmailNoticeCache interface {
	GetVerificationCode(ctx context.Context, email, t string) (string, error)
}

type changePwdByEmailIdentityRepository interface {
	GetUserEmailByID(ctx context.Context, id string) (string, error)
	UpdateUsersAccount(ctx context.Context, userAccount model.UserAccount) error
}

type changePwdByEmailIdentityValidate interface {
	CheckPasswordFormat(password string) bool
}

type changePwdByEmailCrypt interface {
	HashPassword(password string) (string, error)
}

type ChangeSelfPasswordByEmailLogic struct {
	ctx              context.Context
	noticeCache      changePwdByEmailNoticeCache
	userRepository   changePwdByEmailIdentityRepository
	identityValidate changePwdByEmailIdentityValidate
	crypt            changePwdByEmailCrypt
}

func NewChangeSelfPasswordByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeSelfPasswordByEmailLogic {
	return &ChangeSelfPasswordByEmailLogic{
		ctx:              ctx,
		userRepository:   svcCtx.IdentityRepository,
		noticeCache:      svcCtx.NoticeCache,
		crypt:            svcCtx.Crypt,
		identityValidate: svcCtx.IdentityValidate,
	}
}

func (l *ChangeSelfPasswordByEmailLogic) ChangeSelfPasswordByEmail(in *userrpc.ChangePasswordByEmailRequest) (*userrpc.Response, error) {
	// todo: add your logic here and delete this line
	// 验证码的获取与校验
	realVerificationCode, err := l.noticeCache.GetVerificationCode(l.ctx, in.Email, "change_password")
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}
	if realVerificationCode == "" {
		return &userrpc.Response{
			Code:    1,
			Message: "无效的验证码！",
		}, nil
	}
	if realVerificationCode != in.VerificationCode {
		return &userrpc.Response{
			Code:    1,
			Message: "验证码输入有误！",
		}, nil
	}

	// 根据ID获得用户邮箱
	email, err := l.userRepository.GetUserEmailByID(l.ctx, in.ID)
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}
	if email == "" {
		return &userrpc.Response{
			Code:    1,
			Message: "用户账号不存在！",
		}, nil
	}

	// 核对填写的邮箱
	if email != in.Email {
		return &userrpc.Response{
			Code:    1,
			Message: "绑定邮箱输入有误，请检查账号的绑定邮箱和输入的邮箱是否一致！",
		}, nil
	}

	// 密码校验
	passwordAvailable := l.identityValidate.CheckPasswordFormat(in.NewPassword)
	if !passwordAvailable {
		return &userrpc.Response{
			Code:    1,
			Message: "密码格式错误，请检查后重试！",
		}, nil
	}

	// 密码加密
	hashPwd, err := l.crypt.HashPassword(in.NewPassword)
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}

	// 更新数据库
	err = l.userRepository.UpdateUsersAccount(l.ctx, model.UserAccount{
		ID:       in.ID,
		Password: hashPwd,
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
