package logic

import (
	"context"

	"Noviterra/server/app/identity/identityrpc/internal/svc"
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/utils/errcode"

	"go.uber.org/zap"
)

type (
	loginByEmailIdentityCache interface {
		GetVerificationCode(ctx context.Context, email string, t string) (string, error)
		DeleteVerificationCode(ctx context.Context, email, t string) error
	}
	loginByEmailIdentityRepository interface {
		CheckEmailExists(ctx context.Context, email string) (bool, error)
		GetUserIDByEmail(ctx context.Context, id string) (string, error)
	}
	loginByEmailIdentityResponse interface {
		SuccessLoginByEmailResponse(resp *identityrpc.LoginByEmailResponse, id string) (*identityrpc.LoginByEmailResponse, error)
		FailLoginByEmailResponse(resp *identityrpc.LoginByEmailResponse, msg string, status int32, code int64) (*identityrpc.LoginByEmailResponse, error)
		FailLoginByEmailResponseWithErr(resp *identityrpc.LoginByEmailResponse, err error) (*identityrpc.LoginByEmailResponse, error)
	}
	loginByEmailLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type LoginByEmailLogic struct {
	ctx                context.Context
	noticeCache        loginByEmailIdentityCache
	identityRepository loginByEmailIdentityRepository
	identityResponse   loginByEmailIdentityResponse
	logger             loginByEmailLogger
}

func NewLoginByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByEmailLogic {
	return &LoginByEmailLogic{
		ctx:                ctx,
		noticeCache:        svcCtx.NoticeCache,
		identityRepository: svcCtx.IdentityRepository,
		identityResponse:   svcCtx.IdentityResponse,
		logger:             svcCtx.Logger,
	}
}

// 通过邮箱验证登录
func (l *LoginByEmailLogic) LoginByEmail(in *identityrpc.LoginByEmailRequest) (*identityrpc.LoginByEmailResponse, error) {
	resp := new(identityrpc.LoginByEmailResponse)

	// 验证码获取与校验
	verificationCode, err := l.noticeCache.GetVerificationCode(l.ctx, in.Email, "login")
	if err != nil {
		l.logger.Error("缓存操作失败：查询登录验证码失败",
			zap.String("email", in.Email),
			zap.Error(err),
		)
		return l.identityResponse.FailLoginByEmailResponse(resp, "获取验证码失败！",
			errcode.InvalidParam, errcode.GetAuthVerificationCodeFail)
	}
	if verificationCode == "" {
		l.logger.Error("读取到的登录验证码为空",
			zap.String("email", in.Email),
		)
		return l.identityResponse.FailLoginByEmailResponse(resp, "无效的验证码，请检查邮箱或验证码是否正确！",
			errcode.InvalidParam, errcode.VerificationCodeExpired)
	}
	if verificationCode != in.VerificationCode {
		l.logger.Error("输入验证码与缓存验证码不同",
			zap.String("email", in.Email),
		)
		return l.identityResponse.FailLoginByEmailResponse(resp, "验证码输入有误！",
			errcode.InvalidParam, errcode.VerificationCodeInvalid)
	}

	// 检验邮箱是否存在
	emailExist, err := l.identityRepository.CheckEmailExists(l.ctx, in.Email)
	if err != nil {
		l.logger.Error("数据库操作失败：校验邮箱是否存在失败",
			zap.String("email", in.Email),
			zap.Error(err),
		)
		return l.identityResponse.FailLoginByEmailResponse(resp, "数据库操作异常！",
			errcode.InvalidParam, errcode.GetAuthInfoFail)
	}
	if !emailExist {
		l.logger.Error("用户邮箱不存在",
			zap.String("email", in.Email),
		)
		return l.identityResponse.FailLoginByEmailResponse(resp, "邮箱输入有误！",
			errcode.InvalidParam, errcode.EmailNotFound)
	}

	// 获取用户ID
	id, err := l.identityRepository.GetUserIDByEmail(l.ctx, in.Email)
	if err != nil {
		l.logger.Error("数据库操作失败：根据邮箱查询用户ID失败",
			zap.String("email", in.Email),
			zap.Error(err),
		)
		return l.identityResponse.FailLoginByEmailResponse(resp, "数据库操作异常！",
			errcode.InvalidParam, errcode.GetAuthInfoFail)
	}
	if id == "" {
		l.logger.Error("根据邮箱查询到的用户ID为空",
			zap.String("email", in.Email),
		)
		return l.identityResponse.FailLoginByEmailResponse(resp, "账号信息异常，请联系管理员！",
			errcode.InvalidParam, errcode.AccountNotFound)
	}

	// 信息校验完成后删除验证码信息
	err = l.noticeCache.DeleteVerificationCode(l.ctx, in.Email, "login")
	if err != nil { // 缓存删除失败但不反馈 仅日志记录
		l.logger.Error("缓存操作失败：删除登录验证码失败",
			zap.Strings("user-info", []string{id, in.Email}),
			zap.Error(err),
		)
	}

	l.logger.Info("用户通过邮箱验证登录成功",
		zap.Strings("user-info", []string{id, in.Email}),
	)
	return l.identityResponse.SuccessLoginByEmailResponse(resp, id)
}
