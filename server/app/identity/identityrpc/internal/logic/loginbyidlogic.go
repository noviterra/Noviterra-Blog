package logic

import (
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/identity/identityrpc/internal/svc"
)

type loginByIDIdentityRepository interface {
	GetUserPasswordByID(id string) (string, error)
}

type loginByIdentityCommonUtils interface {
	CheckPasswordFormat(password string) bool
}

type loginByIdentityCrypt interface {
	CheckPasswords(password, hashPassword string) (bool, error)
}

type loginByIDIdentityResponse interface {
	SuccessResponse(resp *identityrpc.Response, msg string) (*identityrpc.Response, error)
	FailResponse(resp *identityrpc.Response, msg string, status int32, code int64) (*identityrpc.Response, error)
	FailResponseWithErr(resp *identityrpc.Response, err error) (*identityrpc.Response, error)
}
type LoginByIDLogic struct {
	ctx                context.Context
	identityRepository loginByIDIdentityRepository
	commonUtils        loginByIdentityCommonUtils
	crypt              loginByIdentityCrypt
	identityResponse   loginByIDIdentityResponse
}

func NewLoginByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByIDLogic {
	return &LoginByIDLogic{
		ctx:                ctx,
		identityRepository: svcCtx.IdentityRepository,
		commonUtils:        svcCtx.CommonUtils,
		crypt:              svcCtx.Crypt,
		identityResponse:   svcCtx.IdentityResponse,
	}
}

// 通过账号登录
func (l *LoginByIDLogic) LoginByID(in *identityrpc.LoginByIDRequest) (*identityrpc.Response, error) {
	// todo: add your logic here and delete this line
	// todo: 后期增加布隆过滤器
	resp := new(identityrpc.Response)
	// 密码格式判断
	pwdAvailable := l.commonUtils.CheckPasswordFormat(in.Password)
	if !pwdAvailable {
		return l.identityResponse.FailResponse(resp, "密码格式输入有误！", errcode.InvalidParam, errcode.PasswordInvalidFormat)
	}

	// 根据账号返回密码
	hashPwd, err := l.identityRepository.GetUserPasswordByID(in.ID)
	if err != nil {
		return l.identityResponse.FailResponse(resp, "数据库操作异常！", errcode.InvalidParam, errcode.PasswordInvalidFormat)
	}
	if hashPwd == "" {
		return l.identityResponse.FailResponse(resp, "账号输入有误！", errcode.InvalidParam, errcode.PasswordInvalidFormat)
	}

	// 密码对比
	ok, err := l.crypt.CheckPasswords(in.Password, hashPwd)
	if err != nil {
		return l.identityResponse.FailResponse(resp, "密码加密失败！", errcode.InvalidParam, errcode.PasswordInvalidFormat)
	}
	if !ok {
		return l.identityResponse.FailResponse(resp, "密码输入有误！", errcode.InvalidParam, errcode.PasswordInvalidFormat)
	}

	return l.identityResponse.SuccessResponse(resp, "登录成功！")
}
