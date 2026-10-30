package logic

import (
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/identity/identityrpc/internal/svc"
)

type loginByEmailIdentityCache interface {
	GetVerificationCode(ctx context.Context, email string, t string) (string, error)
	DeleteVerificationCode(ctx context.Context, email, t string) error
}

type loginByEmailIdentityRepository interface {
	CheckEmailExists(email string) (bool, error)
	GetUserIDByEmail(id string) (string, error)
}

type loginByEmailIdentityResponse interface {
	SuccessLoginByEmailResponse(resp *identityrpc.LoginByEmailResponse, id string) (*identityrpc.LoginByEmailResponse, error)
	FailLoginByEmailResponse(resp *identityrpc.LoginByEmailResponse, msg string, status int32, code int64) (*identityrpc.LoginByEmailResponse, error)
	FailLoginByEmailResponseWithErr(resp *identityrpc.LoginByEmailResponse, err error) (*identityrpc.LoginByEmailResponse, error)
}
type LoginByEmailLogic struct {
	ctx                context.Context
	noticeCache        loginByEmailIdentityCache
	identityRepository loginByEmailIdentityRepository
	identityResponse   loginByEmailIdentityResponse
}

func NewLoginByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByEmailLogic {
	return &LoginByEmailLogic{
		ctx:                ctx,
		noticeCache:        svcCtx.NoticeCache,
		identityRepository: svcCtx.IdentityRepository,
		identityResponse:   svcCtx.IdentityResponse,
	}
}

// 通过邮箱验证登录
func (l *LoginByEmailLogic) LoginByEmail(in *identityrpc.LoginByEmailRequest) (*identityrpc.LoginByEmailResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(identityrpc.LoginByEmailResponse)
	// 验证码获取与校验
	verificationCode, err := l.noticeCache.GetVerificationCode(l.ctx, in.Email, "login")
	if err != nil {
		return l.identityResponse.FailLoginByEmailResponse(resp, "获取验证码失败！", errcode.InvalidParam, errcode.Failed)
	}
	if verificationCode == "" {
		return l.identityResponse.FailLoginByEmailResponse(resp, "无效的验证码，请检查邮箱或验证码是否正确！", errcode.InvalidParam, errcode.Failed)
	}
	if verificationCode != in.VerificationCode {
		return l.identityResponse.FailLoginByEmailResponse(resp, "验证码输入有误！", errcode.InvalidParam, errcode.Failed)
	}

	// 检验邮箱是否存在
	emailExist, err := l.identityRepository.CheckEmailExists(in.Email)
	if err != nil {
		return l.identityResponse.FailLoginByEmailResponse(resp, "数据库操作异常！", errcode.InvalidParam, errcode.Failed)
	}
	if !emailExist {
		return l.identityResponse.FailLoginByEmailResponse(resp, "邮箱输入有误！", errcode.InvalidParam, errcode.Failed)
	}

	// 获取用户ID
	id, err := l.identityRepository.GetUserIDByEmail(in.Email)
	if err != nil {
		return l.identityResponse.FailLoginByEmailResponse(resp, "数据库操作异常！", errcode.InvalidParam, errcode.Failed)

	}
	if id == "" {
		return l.identityResponse.FailLoginByEmailResponse(resp, "账号信息异常，请联系管理员！", errcode.InvalidParam, errcode.Failed)

	}

	// 信息校验完成后删除验证码信息
	err = l.noticeCache.DeleteVerificationCode(l.ctx, in.Email, "login")
	if err != nil { // 缓存删除失败但不反馈 仅日志记录

	}

	return l.identityResponse.SuccessLoginByEmailResponse(resp, id)
}
