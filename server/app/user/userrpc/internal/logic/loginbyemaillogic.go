package logic

import (
	"context"

	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginByEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByEmailLogic {
	return &LoginByEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginByEmailLogic) LoginByEmail(in *userrpc.LoginByEmailRequest) (*userrpc.LoginByEmailResponse, error) {
	// todo: add your logic here and delete this line
	// 验证码获取与校验
	verificationCode, err := l.svcCtx.RDB.GetVerificationCode(l.ctx, in.Email, "login")
	if err != nil {
		l.Logger.Error("验证码缓存异常！")
		return &userrpc.LoginByEmailResponse{
			Code:    1,
			Message: "验证码缓存异常，请联系管理员！",
		}, nil
	}
	if verificationCode == "" {
		return &userrpc.LoginByEmailResponse{
			Code:    1,
			Message: "无效的验证码，请检查邮箱或验证码是否正确！",
		}, nil
	}
	if verificationCode != in.VerificationCode {
		return &userrpc.LoginByEmailResponse{
			Code:    1,
			Message: "验证码输入有误！",
		}, nil
	}
	l.svcCtx.Logger.Info("验证码校验通过！")

	// 检验邮箱是否存在
	emailExist, err := l.svcCtx.DB.CheckEmailExists(in.Email)
	if err != nil {
		l.Logger.Error("数据库异常！")
		return &userrpc.LoginByEmailResponse{
			Code:    1,
			Message: "数据库异常，请联系管理员！",
		}, err
	}
	if !emailExist {
		return &userrpc.LoginByEmailResponse{
			Code:    1,
			Message: "邮箱输入有误！",
		}, nil
	}
	l.svcCtx.Logger.Info("邮箱校验通过！")

	// 获取用户ID
	id, err := l.svcCtx.DB.GetUserIDByEmail(in.Email)
	if err != nil {
		return &userrpc.LoginByEmailResponse{
			Code:    1,
			Message: "数据库异常！",
		}, err
	}
	if id == "" {
		return &userrpc.LoginByEmailResponse{
			Code:    1,
			Message: "账号信息异常，请联系管理员！",
		}, nil
	}

	// 信息校验完成后删除验证码信息
	err = l.svcCtx.RDB.DeleteVerificationCode(l.ctx, in.Email, "login")
	if err != nil {
		l.svcCtx.Logger.Error("验证码缓存清除失败！")
	}

	return &userrpc.LoginByEmailResponse{
		Code:    0,
		Message: "登录成功！",
		ID:      id,
	}, nil
}
