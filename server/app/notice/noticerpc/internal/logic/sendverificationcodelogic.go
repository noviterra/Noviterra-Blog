package logic

import (
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/notice/noticerpc/internal/svc"
	"Noviterra/server/app/notice/noticerpc/types/noticerpc"
)

type (
	sendVerificaitonCodeIdentityRepository interface {
		CheckEmailExists(ctx context.Context, email string) (bool, error)
	}

	sendVerificationCodeNoticeResponse interface {
		SuccessResponse(resp *noticerpc.Response, msg string) (*noticerpc.Response, error)
		FailResponse(resp *noticerpc.Response, msg string, status int32, code int64) (*noticerpc.Response, error)
		FailResponseWithErr(resp *noticerpc.Response, err error) (*noticerpc.Response, error)
	}

	sendVerificationCodeNoticeCache interface {
		LimitVerificationCode(ctx context.Context, email string) (bool, error)
		StoreVerificationCode(ctx context.Context, email string, t string, verificatioCode string) error
	}

	sendVerificationCodeCommonUtils interface {
		GenerateVerificationCode() (string, error)
	}

	sendVerificationEmailSender interface {
		TextEmailEditor(verificationCode string, t string) (string, string)
		SendTextEmails(emails []string, emailSubject, emailBody string) error
	}
)

type SendVerificationCodeLogic struct {
	ctx                context.Context
	identityRepository sendVerificaitonCodeIdentityRepository
	noticeResponse     sendVerificationCodeNoticeResponse
	noticeCache        sendVerificationCodeNoticeCache
	commonUtils        sendVerificationCodeCommonUtils
	emailSender        sendVerificationEmailSender
}

func NewSendVerificationCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendVerificationCodeLogic {
	return &SendVerificationCodeLogic{
		ctx:                ctx,
		identityRepository: svcCtx.IdentityRepository,
		noticeResponse:     svcCtx.NoticeResponse,
		noticeCache:        svcCtx.NoticeCache,
		emailSender:        svcCtx.EmailSender,
		commonUtils:        svcCtx.CommonUtils,
	}
}

// 身份验证需要的邮箱验证码验证
func (l *SendVerificationCodeLogic) SendVerificationCode(in *noticerpc.SendVerificationCodeRequest) (*noticerpc.Response, error) {
	// todo: add your logic here and delete this line
	resp := new(noticerpc.Response)
	// 检查邮箱是否存在
	emailExist, err := l.identityRepository.CheckEmailExists(l.ctx, in.Email)
	if err != nil {
		return l.noticeResponse.FailResponse(resp, "数据库操作异常！", errcode.InvalidParam, errcode.Failed)
	}
	// 对于注册 需要保证邮箱不存在 即没有注册过账号
	if in.Type == "register" {
		if emailExist {
			return l.noticeResponse.FailResponse(resp, "邮箱已经注册过账号！", errcode.InvalidParam, errcode.Failed)
		}
	} else {
		// 对于其他则需要保证邮箱已经注册过账号
		if !emailExist {
			return l.noticeResponse.FailResponse(resp, "邮箱尚未注册账号！", errcode.InvalidParam, errcode.Failed)
		}
	}

	// 缓存查询 限制一分钟内只可以发送一次
	ok, err := l.noticeCache.LimitVerificationCode(l.ctx, in.Email)
	if err != nil {
		return l.noticeResponse.FailResponse(resp, "验证码发送失败！", errcode.InvalidParam, errcode.Failed)
	}
	if !ok {
		return l.noticeResponse.FailResponse(resp, "验证码发送频繁！", errcode.InvalidParam, errcode.Failed)
	}

	// 生成验证码
	verificationCode, err := l.commonUtils.GenerateVerificationCode()
	if err != nil {
		return l.noticeResponse.FailResponse(resp, "验证码生成失败！", errcode.InvalidParam, errcode.Failed)
	}
	// 存入缓存中 五分钟有效
	err = l.noticeCache.StoreVerificationCode(l.ctx, in.Email, in.Type, verificationCode)
	if err != nil { // 存入失败则后续的验证工作也无法进行 直接报错
		return l.noticeResponse.FailResponse(resp, "邮件发送异常，请重新发送验证码！", errcode.InvalidParam, errcode.Failed)
	}

	// 发送验证码
	emailSubject, emailBody := l.emailSender.TextEmailEditor(verificationCode, in.Type)
	if emailSubject == "" || emailBody == "" {
		return l.noticeResponse.FailResponse(resp, "验证码发送失败！", errcode.InvalidParam, errcode.Failed)
	}
	err = l.emailSender.SendTextEmails([]string{in.Email}, emailSubject, emailBody)
	if err != nil {
		return l.noticeResponse.FailResponse(resp, "验证码发送失败！", errcode.InvalidParam, errcode.Failed)

	}

	return l.noticeResponse.SuccessResponse(resp, "验证码发送成功！")
}
