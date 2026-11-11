package logic

import (
	"context"

	"Noviterra/server/app/identity/identityrpc/internal/svc"
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"

	"go.uber.org/zap"
)

type (
	forgerPasswordCommonUtils interface {
		ValidatePasswordFormat(password string) (string, error)
	}
	forgerPasswordIdentityCache interface {
		GetVerificationCode(ctx context.Context, email, t string) (string, error)
		DeleteVerificationCode(ctx context.Context, email, t string) error
	}
	forgetPasswordIdentityRepository interface {
		GetUserEmailByID(ctx context.Context, id string) (string, error)
		UpdateUsersAccount(ctx context.Context, account model.UserAccount) error
	}
	forgetPasswordIdentiryBcrypt interface {
		HashPassword(password string) (string, error)
	}
	forgetPasswordIdentityResponse interface {
		SuccessResponse(resp *identityrpc.Response, msg string) (*identityrpc.Response, error)
		FailResponse(resp *identityrpc.Response, msg string, status int32, code int64) (*identityrpc.Response, error)
		FailResponseWithErr(resp *identityrpc.Response, err error) (*identityrpc.Response, error)
	}
	forgetPasswordCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}
	forgetPasswordLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type ForgetPasswordLogic struct {
	ctx                 context.Context
	commonUtils         forgerPasswordCommonUtils
	noticeCache         forgerPasswordIdentityCache
	identityRespository forgetPasswordIdentityRepository
	crypt               forgetPasswordIdentiryBcrypt
	identityResponse    forgetPasswordIdentityResponse
	ctxdata             forgetPasswordCtxData
	logger              forgetPasswordLogger
}

func NewForgetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForgetPasswordLogic {
	return &ForgetPasswordLogic{
		ctx:                 ctx,
		commonUtils:         svcCtx.IdentityValidate,
		noticeCache:         svcCtx.NoticeCache,
		identityRespository: svcCtx.IdentityRepository,
		crypt:               svcCtx.Crypt,
		identityResponse:    svcCtx.IdentityResponse,
		ctxdata:             svcCtx.Ctxdata,
		logger:              svcCtx.Logger,
	}
}

// 忘记密码请求
func (l *ForgetPasswordLogic) ForgetPassword(in *identityrpc.ForgetPasswordRequest) (*identityrpc.Response, error) {
	resp := new(identityrpc.Response)

	// 校验数据
	// 用户ID校验
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		l.logger.Error("上下文获取操作用户ID失败")
		return l.identityResponse.FailResponse(resp, "用户信息校验失败！",
			errcode.InvalidParam, errcode.GetAuthInfoFail)
	}

	// 需要保证操作用户与被操作用户是同一个
	if id != in.ID {
		l.logger.Error("请求发起者ID与被操作ID不同",
			zap.Strings("user-id||operator-id", []string{id, in.ID}),
		)
		return l.identityResponse.FailResponse(resp, "无法操作他人账号！",
			errcode.InvalidParam, errcode.IdentityOperationOutOfBounds)
	}

	// 密码校验
	_, err := l.commonUtils.ValidatePasswordFormat(in.NewPassword)
	if err != nil {
		l.logger.Error("校验密码格式失败",
			zap.String("user-info", id),
			zap.Error(err),
		)
		return l.identityResponse.FailResponse(resp, "密码格式有误，需要同时出现大小写字母、特殊字符、数字！",
			errcode.InvalidParam, errcode.PasswordInvalidFormat)
	}
	if in.NewPassword != in.VerifPassword {
		l.logger.Error("两次输入的密码不一致",
			zap.String("user-info", id),
		)
		return l.identityResponse.FailResponse(resp, "两次密码输入不一致，请确认后重新输入！",
			errcode.InvalidParam, errcode.PasswordNotMatch)
	}

	// 验证码获取与校验
	realVerification, err := l.noticeCache.GetVerificationCode(l.ctx, in.Email, "forget_password")
	if err != nil {
		l.logger.Error("缓存操作失败：查询验证码失败",
			zap.Strings("user-info", []string{id, in.Email}),
			zap.Error(err),
		)
		return l.identityResponse.FailResponse(resp, "缓存操作异常！",
			errcode.InvalidParam, errcode.GetAuthVerificationCodeFail)
	}
	if realVerification == "" {
		l.logger.Error("读取到的验证码为空",
			zap.Strings("user-info", []string{id, in.Email}),
		)
		return l.identityResponse.FailResponse(resp, "验证码已失效！",
			errcode.InvalidParam, errcode.VerificationCodeExpired)
	}

	if in.VerificationCode != realVerification {
		l.logger.Error("输入验证码与缓存验证码不同",
			zap.Strings("user-info", []string{id, in.Email}),
		)
		return l.identityResponse.FailResponse(resp, "验证码错误！",
			errcode.InvalidParam, errcode.VerificationCodeInvalid)
	}

	// 最后再进行仓库操作
	// 根据用户ID得到对应的邮箱  邮箱与ID强绑定并且唯一
	email, err := l.identityRespository.GetUserEmailByID(l.ctx, in.ID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据用户ID查询用户邮箱失败",
			zap.String("user-info", id),
			zap.Error(err),
		)
		return l.identityResponse.FailResponse(resp, "数据库操作异常！",
			errcode.InternalError, errcode.GetAuthInfoFail)
	}
	if email == "" {
		l.logger.Error("根据用户ID查询到的用户邮箱为空",
			zap.String("user-info", id),
		)
		return l.identityResponse.FailResponse(resp, "输入的账号不存在！ ",
			errcode.InvalidParam, errcode.AccountNotFound)
	}

	if email != in.Email {
		l.logger.Error("传参邮箱与用户绑定邮箱不同",
			zap.Strings("user-info", []string{id, email}),
			zap.String("param-email", in.Email),
		)
		return l.identityResponse.FailResponse(resp, "邮箱非该账号的绑定邮箱，请检查自己的账号与邮箱信息！",
			errcode.InternalError, errcode.EmailMismatch)
	}

	hashPwd, err := l.crypt.HashPassword(in.NewPassword)
	if err != nil {
		l.logger.Error("密码加密失败",
			zap.Strings("user-info", []string{id, email}),
			zap.Error(err),
		)
		return l.identityResponse.FailResponse(resp, "密码加密失败！",
			errcode.InternalError, errcode.PasswordHashFail)
	}

	// 更新账号信息
	err = l.identityRespository.UpdateUsersAccount(l.ctx, model.UserAccount{
		ID:       in.ID,
		Email:    in.Email,
		Password: hashPwd,
	})
	if err != nil {
		l.logger.Error("数据库操作失败：用户密码更新失败",
			zap.Strings("user-info", []string{id, email}),
			zap.Error(err),
		)
		return l.identityResponse.FailResponse(resp, "数据库操作异常！",
			errcode.InternalError, errcode.UpdatePasswordFail)
	}

	// 操作完成后删除缓存
	err = l.noticeCache.DeleteVerificationCode(l.ctx, in.Email, "forget_password")
	if err != nil { // 缓存清除失败不给用户返回
		l.logger.Error("缓存操作失败：删除验证码失败",
			zap.Strings("user-info", []string{id, email}),
			zap.Error(err),
		)
	}

	l.logger.Info("用户通过邮箱验证重置密码成功",
		zap.Strings("user-info", []string{id, email}),
	)
	return l.identityResponse.SuccessResponse(resp, "密码重置成功！")
}
