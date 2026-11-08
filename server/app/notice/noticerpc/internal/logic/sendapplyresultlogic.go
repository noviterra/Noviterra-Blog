package logic

import (
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"
	"context"

	"Noviterra/server/app/notice/noticerpc/internal/svc"
	"Noviterra/server/app/notice/noticerpc/types/noticerpc"
)

type (
	sendApplyResultNoticeResponse interface {
		SuccessResponse(resp *noticerpc.Response, msg string) (*noticerpc.Response, error)
		FailResponse(resp *noticerpc.Response, msg string, status int32, code int64) (*noticerpc.Response, error)
		FailResponseWithErr(resp *noticerpc.Response, err error) (*noticerpc.Response, error)
	}

	sendApplyResultEmailSender interface {
		SendTextEmails(emails []string, emailSubject, emailBody string) error
		TextFriendLinkEmail(link, t string) (string, string)
	}

	sendApplyResultCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	sendApplyResultIdentityRepositity interface {
		GetUserEmailByID(ctx context.Context, email string) (string, error)
	}

	sendApplyResultNoticeValidate interface {
		ValidateFriendLinkEmailResultInfo(in validate.InputValidateFriendLinkResult) (validate.OutputValidateFriendLinkResult, error)
	}
)

type SendApplyResultLogic struct {
	ctx                context.Context
	ctxdata            sendApplyResultCtxData
	emailSender        sendApplyResultEmailSender
	noticeResponse     sendApplyResultNoticeResponse
	identityRepositity sendApplyResultIdentityRepositity
	noticeValidate     sendApplyResultNoticeValidate
}

func NewSendApplyResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendApplyResultLogic {
	return &SendApplyResultLogic{
		ctx:                ctx,
		ctxdata:            svcCtx.CtxData,
		emailSender:        svcCtx.EmailSender,
		noticeResponse:     svcCtx.NoticeResponse,
		identityRepositity: svcCtx.IdentityRepository,
		noticeValidate:     svcCtx.NoticeValidate,
	}
}

// 发送友链申请进度邮件
func (l *SendApplyResultLogic) SendApplyResult(in *noticerpc.SendApplyResultRequest) (*noticerpc.Response, error) {
	// todo: add your logic here and delete this line
	resp := new(noticerpc.Response)
	// 参数校验
	op, err := l.noticeValidate.ValidateFriendLinkEmailResultInfo(validate.InputValidateFriendLinkResult{
		Email:   in.Email,
		Link:    in.Link,
		Process: in.Process,
	})
	if err != nil {
		return l.noticeResponse.FailResponseWithErr(resp, err)
	}

	// 用户校验 操作用户与被操作用户是否相同
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		return l.noticeResponse.FailResponse(resp, "信息校验失败！", errcode.InvalidParam, errcode.Failed)
	}

	// 邮箱输入为空的话 找用户注册账号对应的email
	if op.Email == "" {
		op.Email, err = l.identityRepositity.GetUserEmailByID(l.ctx, id)
		if err != nil {
			// 日志记录
		}
	}

	emailSubject, emailBody := l.emailSender.TextFriendLinkEmail(op.Link, op.Process)
	if emailSubject == "" || emailBody == "" { // 发送进度的验证码 不应该给前端用户返回err 仅在日志中记录就行
		return l.noticeResponse.FailResponse(resp, "信息编辑失败！", errcode.InvalidParam, errcode.Failed)
	}
	err = l.emailSender.SendTextEmails([]string{op.Email}, emailSubject, emailBody)
	if err != nil {
		return l.noticeResponse.FailResponse(resp, "邮件发送失败！", errcode.InternalError, errcode.Failed)
	}

	return l.noticeResponse.SuccessResponse(resp, "邮件发送成功！")
}
