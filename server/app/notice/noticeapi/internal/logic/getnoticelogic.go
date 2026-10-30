// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/notice/noticerpc/types/noticerpc"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/notice/noticeapi/internal/svc"
	"Noviterra/server/app/notice/noticeapi/internal/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	getNoticeModel interface {
		GetNotice(ctx context.Context, in *noticerpc.GetNoticeRequest, opts ...grpc.CallOption) (*noticerpc.GetNoticeResponse, error)
	}

	getNoticeServiceTokenInterceptor interface {
		GetServiceTokenInConf() string
		GetSkipTokenInConf() string
	}
)

type GetNoticeLogic struct {
	ctx                     context.Context
	noticeModel             getNoticeModel
	serviceTokenInterceptor getNoticeServiceTokenInterceptor
}

// 得到公告
func NewGetNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNoticeLogic {
	return &GetNoticeLogic{
		ctx:                     ctx,
		noticeModel:             svcCtx.NoticeModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *GetNoticeLogic) GetNotice() (resp *types.GetNoticeResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.GetNoticeResponse)

	// 注入rpc服务凭证
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"skip-token", l.serviceTokenInterceptor.GetSkipTokenInConf(),
	)

	rep, err := l.noticeModel.GetNotice(ctx, &noticerpc.GetNoticeRequest{})
	if err != nil {
		return l.FailGetNoticeResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return l.FailGetNoticeResponse(resp, rep.Resp.Message, int(rep.Resp.Status), int(rep.Resp.Code))
	}

	return l.SuccessGetNoticeResponse(resp, rep)
}

func (l *GetNoticeLogic) SuccessGetNoticeResponse(resp *types.GetNoticeResponse, rep *noticerpc.GetNoticeResponse) (*types.GetNoticeResponse, error) {
	if resp == nil {
		resp = new(types.GetNoticeResponse)
	}

	for _, notice := range rep.Notices {
		var n types.Notice
		n.Title = notice.Title
		n.Message = notice.Message
		n.CreatedTime = notice.CreatedAt
		n.Publisher = notice.Publisher

		resp.Notices = append(resp.Notices, n)
	}
	resp.Resp = types.Response{
		Status:  errcode.RequestSuccess,
		Code:    errcode.Success,
		Message: "公告加载完成！",
	}

	return resp, nil
}

func (l *GetNoticeLogic) FailGetNoticeResponse(resp *types.GetNoticeResponse, msg string, status, code int) (*types.GetNoticeResponse, error) {
	if resp == nil {
		resp = new(types.GetNoticeResponse)
	}

	resp.Notices = nil
	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "公告加载失败：" + msg,
	}

	return resp, nil
}

func (l *GetNoticeLogic) FailGetNoticeResponseWithErr(resp *types.GetNoticeResponse, err error, status, code int) (*types.GetNoticeResponse, error) {
	if resp == nil {
		resp = new(types.GetNoticeResponse)
	}

	resp.Notices = nil
	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "公告加载失败：" + err.Error(),
	}

	return resp, nil
}
