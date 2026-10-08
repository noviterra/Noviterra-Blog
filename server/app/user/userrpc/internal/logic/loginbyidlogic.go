package logic

import (
	"context"

	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginByIDLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByIDLogic {
	return &LoginByIDLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginByIDLogic) LoginByID(in *userrpc.LoginByIDRequest) (*userrpc.LoginByIDResponse, error) {
	// todo: add your logic here and delete this line
	// todo: 后期增加布隆过滤器

	// 密码格式判断
	pwdAvailable := l.svcCtx.CommonUtils.CheckPasswordFormat(in.Password)
	if !pwdAvailable {
		return &userrpc.LoginByIDResponse{
			Code:    1,
			Message: "密码格式输入有误！",
		}, nil
	}

	// 根据账号返回密码
	hashPwd, err := l.svcCtx.DB.GetUserPasswordByID(in.ID)
	if err != nil {
		return &userrpc.LoginByIDResponse{
			Code:    1,
			Message: "数据库异常，请联系管理员",
		}, err
	}
	if hashPwd == "" {
		return &userrpc.LoginByIDResponse{
			Code:    1,
			Message: "账号输入有误！",
		}, nil
	}

	// 密码对比
	ok, err := l.svcCtx.Crypt.CheckPasswords(in.Password, hashPwd)
	if err != nil {
		return &userrpc.LoginByIDResponse{
			Code:    1,
			Message: "加密系统异常，请联系管理员！",
		}, err
	}
	if !ok {
		return &userrpc.LoginByIDResponse{
			Code:    1,
			Message: "密码输入有误！",
		}, nil
	}

	return &userrpc.LoginByIDResponse{
		Code:    0,
		Message: "登录成功！",
	}, nil
}
