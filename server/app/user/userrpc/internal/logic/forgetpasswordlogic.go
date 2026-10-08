package logic

import (
	"context"
	"fmt"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ForgetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewForgetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForgetPasswordLogic {
	return &ForgetPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ForgetPasswordLogic) ForgetPassword(in *userrpc.ForgetPasswordRequest) (*userrpc.ForgetPasswordResponse, error) {
	// todo: add your logic here and delete this line
	// 校验数据
	// 密码校验
	passwordAvailable := l.svcCtx.CommonUtils.CheckPasswordFormat(in.NewPassword)
	if !passwordAvailable {
		return &userrpc.ForgetPasswordResponse{
			Code:    1,
			Message: "密码格式有误，需要同时出现大小写字母、特殊字符、数字！",
		}, nil
	}
	if in.NewPassword != in.VerifPassword {
		return &userrpc.ForgetPasswordResponse{
			Code:    1,
			Message: "两次密码输入不一致，请确认后重新输入！",
		}, nil
	}

	// 验证码获取与校验
	realVerification, err := l.svcCtx.RDB.GetVerificationCode(l.ctx, in.Email, "forget_password")
	if err != nil {
		return &userrpc.ForgetPasswordResponse{
			Code:    1,
			Message: "缓存操作异常！",
		}, err
	}
	if realVerification == "" {
		return &userrpc.ForgetPasswordResponse{
			Code:    1,
			Message: "无效的验证码！",
		}, nil
	}

	if in.VerificationCode != realVerification {
		fmt.Println(in.VerificationCode, realVerification)
		return &userrpc.ForgetPasswordResponse{
			Code:    1,
			Message: "验证码错误！",
		}, nil
	}

	// 最后再进行仓库操作
	// 根据用户ID得到对应的邮箱  邮箱与ID强绑定并且唯一
	email, err := l.svcCtx.DB.GetUserEmailByID(in.ID)
	if err != nil {
		return &userrpc.ForgetPasswordResponse{
			Code:    1,
			Message: "数据库操作异常！",
		}, err
	}
	if email == "" {
		return &userrpc.ForgetPasswordResponse{
			Code:    1,
			Message: "账号不存在！",
		}, nil
	}

	if email != in.Email {
		return &userrpc.ForgetPasswordResponse{
			Code:    1,
			Message: "邮箱非该账号的绑定邮箱，请检查自己的账号与邮箱信息！",
		}, nil
	}

	hashPwd, err := l.svcCtx.Crypt.HashPassword(in.NewPassword)
	if err != nil {
		return &userrpc.ForgetPasswordResponse{
			Code:    1,
			Message: "密码加密失败！",
		}, err
	}

	// 更新账号信息
	err = l.svcCtx.DB.UpdateUsersAccount(model.UserAccount{
		ID:       in.ID,
		Email:    in.Email,
		Password: hashPwd,
	})
	if err != nil {
		return &userrpc.ForgetPasswordResponse{
			Code:    1,
			Message: "数据库操作异常！",
		}, err
	}

	// 操作完成后删除缓存
	err = l.svcCtx.RDB.DeleteVerificationCode(l.ctx, in.Email, "forget_password")
	if err != nil {
		l.svcCtx.Logger.Error("验证码缓存清除失败！")
	}

	return &userrpc.ForgetPasswordResponse{
		Code:    0,
		Message: "密码重置成功！",
	}, nil
}
