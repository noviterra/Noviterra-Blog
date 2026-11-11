package logic

import (
	"context"

	"Noviterra/server/app/identity/identityrpc/internal/svc"
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/utils/errcode"

	"go.uber.org/zap"
)

type (
	loginByIDIdentityRepository interface {
		GetUserPasswordByID(ctx context.Context, id string) (string, error)
	}
	loginByIDentityValidate interface {
		ValidatePasswordFormat(password string) (string, error)
	}
	loginByIDentityCrypt interface {
		CheckPasswords(password, hashPassword string) (bool, error)
	}
	loginByIDIdentityResponse interface {
		SuccessResponse(resp *identityrpc.Response, msg string) (*identityrpc.Response, error)
		FailResponse(resp *identityrpc.Response, msg string, status int32, code int64) (*identityrpc.Response, error)
		FailResponseWithErr(resp *identityrpc.Response, err error) (*identityrpc.Response, error)
	}
	loginByIDLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type LoginByIDLogic struct {
	ctx                context.Context
	identityRepository loginByIDIdentityRepository
	identityValidate   loginByIDentityValidate
	crypt              loginByIDentityCrypt
	identityResponse   loginByIDIdentityResponse
	logger             loginByIDLogger
}

func NewLoginByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByIDLogic {
	return &LoginByIDLogic{
		ctx:                ctx,
		identityRepository: svcCtx.IdentityRepository,
		identityValidate:   svcCtx.IdentityValidate,
		crypt:              svcCtx.Crypt,
		identityResponse:   svcCtx.IdentityResponse,
		logger:             svcCtx.Logger,
	}
}

// 通过账号登录
func (l *LoginByIDLogic) LoginByID(in *identityrpc.LoginByIDRequest) (*identityrpc.Response, error) {
	// todo: 后期增加布隆过滤器
	resp := new(identityrpc.Response)

	// 密码格式判断
	_, err := l.identityValidate.ValidatePasswordFormat(in.Password)
	if err != nil {
		l.logger.Error("校验密码格式失败",
			zap.String("user-info", in.ID),
			zap.Error(err),
		)
		return l.identityResponse.FailResponse(resp, "密码格式输入有误！",
			errcode.InvalidParam, errcode.PasswordInvalidFormat)
	}

	// 根据账号返回密码
	hashPwd, err := l.identityRepository.GetUserPasswordByID(l.ctx, in.ID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据用户ID查询用户密码失败",
			zap.String("user-info", in.ID),
			zap.Error(err),
		)
		return l.identityResponse.FailResponse(resp, "数据库操作异常！",
			errcode.InvalidParam, errcode.GetAuthInfoFail)
	}
	if hashPwd == "" {
		l.logger.Error("根据用户ID查询到的用户密码为空，用户账号不存在",
			zap.String("user-info", in.ID),
		)
		return l.identityResponse.FailResponse(resp, "账号输入有误！",
			errcode.InvalidParam, errcode.AccountNotFound)
	}

	// 密码对比
	ok, err := l.crypt.CheckPasswords(in.Password, hashPwd)
	if err != nil {
		l.logger.Error("密码校验失败",
			zap.String("user-info", in.ID),
			zap.Error(err),
		)
		return l.identityResponse.FailResponse(resp, "密码加密失败！",
			errcode.InvalidParam, errcode.PasswordCheckFail)
	}
	if !ok {
		l.logger.Error("密码输入错误",
			zap.String("user-info", in.ID),
		)
		return l.identityResponse.FailResponse(resp, "密码输入有误！",
			errcode.InvalidParam, errcode.PasswordIncorrect)
	}

	l.logger.Info("用户账号密码登录成功",
		zap.String("user-info", in.ID),
	)
	return l.identityResponse.SuccessResponse(resp, "登录成功！")
}
