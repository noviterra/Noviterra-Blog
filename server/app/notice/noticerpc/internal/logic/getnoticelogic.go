package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/notice/noticerpc/internal/svc"
	"Noviterra/server/app/notice/noticerpc/types/noticerpc"
)

type (
	getNoticeRepository interface {
		GetNoticeInfo(ctx context.Context) ([]model.NoticeLookDTO, error)
	}

	getNoticeResponse interface {
		SuccessGetNoticeResponse(resp *noticerpc.GetNoticeResponse, dtos []model.NoticeLookDTO) (*noticerpc.GetNoticeResponse, error)
		FailGetNoticeResponse(resp *noticerpc.GetNoticeResponse, msg string, status int32, code int64) (*noticerpc.GetNoticeResponse, error)
		FailGetNoticeResponseWithErr(resp *noticerpc.GetNoticeResponse, err error) (*noticerpc.GetNoticeResponse, error)
	}
)

type GetNoticeLogic struct {
	ctx              context.Context
	noticeRepository getNoticeRepository
	noticeResponse   getNoticeResponse
}

func NewGetNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNoticeLogic {
	return &GetNoticeLogic{
		ctx:              ctx,
		noticeRepository: svcCtx.NoticeRepository,
		noticeResponse:   svcCtx.NoticeResponse,
	}
}

// 展示公告
func (l *GetNoticeLogic) GetNotice(in *noticerpc.GetNoticeRequest) (*noticerpc.GetNoticeResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(noticerpc.GetNoticeResponse)

	dtos, err := l.noticeRepository.GetNoticeInfo(l.ctx)
	if err != nil {
		return l.noticeResponse.FailGetNoticeResponse(resp, "数据库操作异常！", errcode.InternalError, errcode.Failed)
	}

	return l.noticeResponse.SuccessGetNoticeResponse(resp, dtos)
}
