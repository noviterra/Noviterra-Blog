// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/notice/noticerpc/types/noticerpc"
	"Noviterra/server/app/site/siterpc/types/siterpc"
	"Noviterra/server/utils/errcode"
	"context"
	"fmt"

	"Noviterra/server/app/site/siteapi/internal/svc"
	"Noviterra/server/app/site/siteapi/internal/types"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	applyFriendLinkSiteModel interface {
		ApplyFriendLink(ctx context.Context, in *siterpc.FriendLinkRequest, opts ...grpc.CallOption) (*siterpc.Response, error)
	}

	applyFriendLinkCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	applyFriendLinkServiceTokenInterceptor interface {
		GetServiceTokenInConf() string
	}

	applyFriendLinkResultNotice interface {
		SendApplyResult(ctx context.Context, in *noticerpc.SendApplyResultRequest, opts ...grpc.CallOption) (*noticerpc.Response, error)
	}
)

type ApplyFriendLinkLogic struct {
	ctx                     context.Context
	siteModel               applyFriendLinkSiteModel
	noticeModel             applyFriendLinkResultNotice
	ctxdata                 applyFriendLinkCtxData
	serviceTokenInterceptor applyFriendLinkServiceTokenInterceptor
}

// 友链申请
func NewApplyFriendLinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyFriendLinkLogic {
	return &ApplyFriendLinkLogic{
		ctx:                     ctx,
		siteModel:               svcCtx.SiteModel,
		noticeModel:             svcCtx.NoticeModel,
		ctxdata:                 svcCtx.CtaData,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *ApplyFriendLinkLogic) ApplyFriendLink(req *types.FriendLinkRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	applicant := l.ctxdata.GetUserIDFromContext(l.ctx)
	if applicant == "" {
		return failResponse(resp, "请先登录账号！", errcode.InvalidParam, errcode.Failed)
	}

	// 服务凭证与用户信息注入到metadata消息中
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", applicant,
	)

	rep, err := l.siteModel.ApplyFriendLink(ctx, &siterpc.FriendLinkRequest{
		RequestInfo: &siterpc.FriendLink{
			Link:            req.RequestInfo.Link,
			LinkName:        req.RequestInfo.LinkName,
			SelfInformation: req.RequestInfo.SelfInformation,
		},
		Email: req.Email,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Code != 0 {
		return failResponse(resp, rep.Message, int(rep.Status), int(rep.Code))
	}

	// 调用邮箱发送信息服务 此服务执行失败不返回fail
	rep2, err := l.noticeModel.SendApplyResult(ctx, &noticerpc.SendApplyResultRequest{
		Email:   req.Email,
		Process: "apply",
		Link:    req.RequestInfo.Link,
	})
	if err != nil {
		fmt.Println("err:", err)
		return successResponse(resp, "申请成功提交成功！")
	}
	if rep2.Code != 0 {
		fmt.Println("rep", rep2.Message)
		return successResponse(resp, "申请提交成功！")
	}

	return successResponse(resp, "申请成功提交，请注意邮箱信息！")
}

// 基础响应体暴露于包内 其他绑定其下结构体

func successResponse(resp *types.Response, msg string) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}

	resp.Message = msg
	resp.Status = errcode.RequestSuccess
	resp.Code = errcode.Success

	return resp, nil
}

func failResponseWithErr(resp *types.Response, err error, status, code int) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}

	resp.Message = "服务失败：" + err.Error()
	resp.Status = status
	resp.Code = code

	return resp, nil
}

func failResponse(resp *types.Response, msg string, status, code int) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}

	resp.Message = msg
	resp.Status = status
	resp.Code = code

	return resp, nil
}
