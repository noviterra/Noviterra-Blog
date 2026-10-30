package logic

import (
	"Noviterra/server/app/identity/identityrpc/internal/svc"
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"
)

type forgerPasswordCommonUtils interface {
	CheckPasswordFormat(password string) bool
}

type forgerPasswordIdentityCache interface {
	GetVerificationCode(ctx context.Context, email, t string) (string, error)
	DeleteVerificationCode(ctx context.Context, email, t string) error
}

type forgetPasswordIdentityRepository interface {
	GetUserEmailByID(id string) (string, error)
	UpdateUsersAccount(account model.UserAccount) error
}

type forgetPasswordIdentiryBcrypt interface {
	HashPassword(password string) (string, error)
}

type forgetPasswordIdentityResponse interface {
	SuccessResponse(resp *identityrpc.Response, msg string) (*identityrpc.Response, error)
	FailResponse(resp *identityrpc.Response, msg string, status int32, code int64) (*identityrpc.Response, error)
	FailResponseWithErr(resp *identityrpc.Response, err error) (*identityrpc.Response, error)
}

type forgetPasswordCtxData interface {
	GetUserIDFromContext(ctx context.Context) string
}
type ForgetPasswordLogic struct {
	ctx                 context.Context
	commonUtils         forgerPasswordCommonUtils
	noticeCache         forgerPasswordIdentityCache
	identityRespository forgetPasswordIdentityRepository
	crypt               forgetPasswordIdentiryBcrypt
	identityResponse    forgetPasswordIdentityResponse
	ctxdata             forgetPasswordCtxData
}

func NewForgetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForgetPasswordLogic {
	return &ForgetPasswordLogic{
		ctx:                 ctx,
		commonUtils:         svcCtx.CommonUtils,
		noticeCache:         svcCtx.NoticeCache,
		identityRespository: svcCtx.IdentityRepository,
		crypt:               svcCtx.Crypt,
		identityResponse:    svcCtx.IdentityResponse,
		ctxdata:             svcCtx.Ctxdata,
	}
}

// 忘记密码请求
func (l *ForgetPasswordLogic) ForgetPassword(in *identityrpc.ForgetPasswordRequest) (*identityrpc.Response, error) {
	// todo: add your logic here and delete this line
	resp := new(identityrpc.Response)
	// 校验数据
	// 用户ID校验
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		return l.identityResponse.FailResponse(resp, "用户信息校验失败！", errcode.InvalidParam, errcode.Failed)
	}

	// 需要保证操作用户与被操作用户是同一个
	if id != in.ID {
		return l.identityResponse.FailResponse(resp, "无法操作他人账号！", errcode.InvalidParam, errcode.Failed)
	}

	// 密码校验
	passwordAvailable := l.commonUtils.CheckPasswordFormat(in.NewPassword)
	if !passwordAvailable {
		return l.identityResponse.FailResponse(resp, "密码格式有误，需要同时出现大小写字母、特殊字符、数字！", errcode.InvalidParam, errcode.PasswordInvalidFormat)
	}
	if in.NewPassword != in.VerifPassword {
		return l.identityResponse.FailResponse(resp, "两次密码输入不一致，请确认后重新输入！", errcode.InvalidParam, errcode.Failed)
	}

	// 验证码获取与校验
	realVerification, err := l.noticeCache.GetVerificationCode(l.ctx, in.Email, "forget_password")
	if err != nil {
		return l.identityResponse.FailResponse(resp, "缓存操作异常！", errcode.InvalidParam, errcode.Failed)
	}
	if realVerification == "" {
		return l.identityResponse.FailResponse(resp, "验证码已失效！", errcode.InvalidParam, errcode.Failed)
	}

	if in.VerificationCode != realVerification {
		return l.identityResponse.FailResponse(resp, "验证码错误！", errcode.InvalidParam, errcode.Failed)
	}

	// 最后再进行仓库操作
	// 根据用户ID得到对应的邮箱  邮箱与ID强绑定并且唯一
	email, err := l.identityRespository.GetUserEmailByID(in.ID)
	if err != nil {
		return l.identityResponse.FailResponse(resp, "数据库操作异常！", errcode.InternalError, errcode.Failed)
	}
	if email == "" {
		return l.identityResponse.FailResponse(resp, "输入的账号不存在！ ", errcode.InvalidParam, errcode.Failed)
	}

	if email != in.Email {
		return l.identityResponse.FailResponse(resp, "邮箱非该账号的绑定邮箱，请检查自己的账号与邮箱信息！", errcode.InternalError, errcode.Failed)
	}

	hashPwd, err := l.crypt.HashPassword(in.NewPassword)
	if err != nil {
		return l.identityResponse.FailResponse(resp, "密码加密失败！", errcode.InternalError, errcode.Failed)
	}

	// 更新账号信息
	err = l.identityRespository.UpdateUsersAccount(model.UserAccount{
		ID:       in.ID,
		Email:    in.Email,
		Password: hashPwd,
	})
	if err != nil {
		return l.identityResponse.FailResponse(resp, "数据库操作异常！", errcode.InternalError, errcode.Failed)
	}

	// 操作完成后删除缓存
	err = l.noticeCache.DeleteVerificationCode(l.ctx, in.Email, "forget_password")
	if err != nil { // 缓存清除失败不给用户返回

	}

	return l.identityResponse.SuccessResponse(resp, "密码重置成功！")
}
