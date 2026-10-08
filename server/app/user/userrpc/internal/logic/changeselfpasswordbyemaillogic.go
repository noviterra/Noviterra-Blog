package logic

import (
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeSelfPasswordByEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeSelfPasswordByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeSelfPasswordByEmailLogic {
	return &ChangeSelfPasswordByEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChangeSelfPasswordByEmailLogic) ChangeSelfPasswordByEmail(in *userrpc.ChangePasswordByEmailRequest) (*userrpc.ChangePasswordByEmailResponse, error) {
	// todo: add your logic here and delete this line
	// 验证码的获取与校验
	realVerificationCode, err := l.svcCtx.RDB.GetVerificationCode(l.ctx, in.Email, "change_password")
	if err != nil {
		return &userrpc.ChangePasswordByEmailResponse{
			Code:    1,
			Message: "",
		}, err
	}
	if realVerificationCode == "" {
		return &userrpc.ChangePasswordByEmailResponse{
			Code:    1,
			Message: "无效的验证码！",
		}, nil
	}
	if realVerificationCode != in.VerificationCode {
		return &userrpc.ChangePasswordByEmailResponse{
			Code:    1,
			Message: "验证码输入有误！",
		}, nil
	}

	// 根据ID获得用户邮箱
	email, err := l.svcCtx.DB.GetUserEmailByID(in.ID)
	if err != nil {
		return &userrpc.ChangePasswordByEmailResponse{
			Code:    1,
			Message: "",
		}, err
	}
	if email == "" {
		return &userrpc.ChangePasswordByEmailResponse{
			Code:    1,
			Message: "用户账号不存在！",
		}, nil
	}

	// 核对填写的邮箱
	if email != in.Email {
		return &userrpc.ChangePasswordByEmailResponse{
			Code:    1,
			Message: "绑定邮箱输入有误，请检查账号的绑定邮箱和输入的邮箱是否一致！",
		}, nil
	}

	// 密码校验
	passwordAvailable := l.svcCtx.CommonUtils.CheckPasswordFormat(in.NewPassword)
	if !passwordAvailable {
		return &userrpc.ChangePasswordByEmailResponse{
			Code:    1,
			Message: "密码格式错误，请检查后重试！",
		}, nil
	}

	// 密码加密
	hashPwd, err := l.svcCtx.Crypt.HashPassword(in.NewPassword)
	if err != nil {
		return &userrpc.ChangePasswordByEmailResponse{
			Code:    1,
			Message: "",
		}, err
	}

	// 更新数据库
	err = l.svcCtx.DB.UpdateUsersAccount(model.UserAccount{
		ID:       in.ID,
		Password: hashPwd,
	})
	if err != nil {
		return &userrpc.ChangePasswordByEmailResponse{
			Code:    1,
			Message: "",
		}, err
	}

	return &userrpc.ChangePasswordByEmailResponse{
		Code:    0,
		Message: "密码修改成功！",
	}, nil
}
