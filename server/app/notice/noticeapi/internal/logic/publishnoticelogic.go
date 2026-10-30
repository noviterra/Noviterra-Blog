// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/notice/noticerpc/types/noticerpc"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/notice/noticeapi/internal/svc"
	"Noviterra/server/app/notice/noticeapi/internal/types"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	publishNoticeModel interface {
		PublishNotice(ctx context.Context, in *noticerpc.PublishNoticeRequest, opts ...grpc.CallOption) (*noticerpc.Response, error)
	}

	publishNoticeCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
		GetUserRoleFromContext(ctx context.Context) int
	}

	publishNoticeServiceTokenInterceptor interface {
		GetServiceTokenInConf() string
	}
)

type PublishNoticeLogic struct {
	ctx                     context.Context
	noticeModel             publishNoticeModel
	ctxdata                 publishNoticeCtxData
	serviceTokenInterceptor publishNoticeServiceTokenInterceptor
}

// 发布公告
func NewPublishNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishNoticeLogic {
	return &PublishNoticeLogic{
		ctx:                     ctx,
		noticeModel:             svcCtx.NoticeModel,
		ctxdata:                 svcCtx.CtxData,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *PublishNoticeLogic) PublishNotice(req *types.PublishStationNoticeRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	// 发布公告要求：只要超级管理员 即role=2
	// 从中间件上下文中得到用户ID与权限大小
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	role := l.ctxdata.GetUserRoleFromContext(l.ctx)
	if id == "" || role < 0 {
		return failResponse(resp, "身份信息校验失败！", errcode.InvalidParam, errcode.Failed)
	}
	if role < 2 {
		return failResponse(resp, "权限不足！", errcode.Forbidden, errcode.Failed)
	}

	// 注入到metadata中
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", id,
	)

	rep, err := l.noticeModel.PublishNotice(ctx, &noticerpc.PublishNoticeRequest{
		Title:   req.Title,
		Message: req.Message,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Code != 0 {
		return failResponse(resp, rep.Message, int(rep.Status), int(rep.Code))
	}

	return successResponse(resp, rep.Message)
}
