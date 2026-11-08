package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/app/notice/noticerpc/internal/svc"
	"Noviterra/server/app/notice/noticerpc/types/noticerpc"
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"
	"context"
)

type (
	publishNoticeResponse interface {
		SuccessResponse(resp *noticerpc.Response, msg string) (*noticerpc.Response, error)
		FailResponse(resp *noticerpc.Response, msg string, status int32, code int64) (*noticerpc.Response, error)
		FailResponseWithErr(resp *noticerpc.Response, err error) (*noticerpc.Response, error)
	}

	publishNoticeCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	publishNoticeIdentityRepository interface {
		GetUserRole(ctx context.Context, id string) (int, error)
	}

	publishNoticeValidate interface {
		ValidateNoticeInfo(in validate.InputValidateNotice) (validate.OutputValidateNotice, error)
	}

	publishNoticeRepository interface {
		PublishNoticeInfo(ctx context.Context, notice model.Notices) error
	}
)

type PublishNoticeLogic struct {
	ctx                context.Context
	ctxdata            publishNoticeCtxData
	noticeResonse      publishNoticeResponse
	identityRepository publishNoticeIdentityRepository
	noticeValidate     publishNoticeValidate
	noticeRepository   publishNoticeRepository
}

func NewPublishNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishNoticeLogic {
	return &PublishNoticeLogic{
		ctx:                ctx,
		ctxdata:            svcCtx.CtxData,
		noticeResonse:      svcCtx.NoticeResponse,
		identityRepository: svcCtx.IdentityRepository,
		noticeValidate:     svcCtx.NoticeValidate,
		noticeRepository:   svcCtx.NoticeRepository,
	}
}

// 发布公告
func (l *PublishNoticeLogic) PublishNotice(in *noticerpc.PublishNoticeRequest) (*noticerpc.Response, error) {
	// todo: add your logic here and delete this line
	resp := new(noticerpc.Response)
	// 传参校验
	notice, err := l.noticeValidate.ValidateNoticeInfo(validate.InputValidateNotice{
		Title:   in.Title,
		Message: in.Message,
	})
	if err != nil {
		return l.noticeResonse.FailResponseWithErr(resp, err)
	}

	// 用户信息校验
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		return l.noticeResonse.FailResponse(resp, "身份信息校验失败！", errcode.InvalidParam, errcode.Failed)
	}
	role, err := l.identityRepository.GetUserRole(l.ctx, id)
	if err != nil {
		return l.noticeResonse.FailResponse(resp, "数据库操作异常！", errcode.InternalError, errcode.Failed)
	}
	if role != 2 {
		return l.noticeResonse.FailResponse(resp, "操作权限不足！", errcode.Forbidden, errcode.Failed)
	}

	// 插入库中
	err = l.noticeRepository.PublishNoticeInfo(l.ctx, model.Notices{
		Title:     notice.Title,
		Message:   notice.Message,
		Publisher: id,
	})
	if err != nil {
		return l.noticeResonse.FailResponse(resp, "数据库操作异常！", errcode.InvalidParam, errcode.Failed)
	}

	return l.noticeResonse.SuccessResponse(resp, "公告发布成功！")
}
