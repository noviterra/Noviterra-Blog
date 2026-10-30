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
	sendVerificationCodeNoticeModel interface {
		SendVerificationCode(ctx context.Context, in *noticerpc.SendVerificationCodeRequest, opts ...grpc.CallOption) (*noticerpc.Response, error)
	}

	sendVerificaitonCodeServiceTokenInterceptor interface {
		GetServiceTokenInConf() string
		GetSkipTokenInConf() string
	}
)

type SendVerificationCodeLogic struct {
	ctx                     context.Context
	noticeModel             sendVerificationCodeNoticeModel
	serviceTokenInterceptor sendVerificaitonCodeServiceTokenInterceptor
}

// 发送辅助身份验证的邮箱验证码
func NewSendVerificationCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendVerificationCodeLogic {
	return &SendVerificationCodeLogic{
		ctx:                     ctx,
		noticeModel:             svcCtx.NoticeModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *SendVerificationCodeLogic) SendVerificationCode(req *types.SendVerificationCodeRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"skip-token", l.serviceTokenInterceptor.GetSkipTokenInConf(),
	)
	rep, err := l.noticeModel.SendVerificationCode(ctx, &noticerpc.SendVerificationCodeRequest{
		Email: req.Email,
		Type:  req.Type,
	})

	if err != nil { // 有err时 rep为nil （go-zero特性）
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Code != 0 { // 无err但是业务失败
		return failResponse(resp, rep.Message, int(rep.Status), int(rep.Code))
	}

	return successResponse(resp, "验证码发送成功！")
}

func successResponse(resp *types.Response, msg string) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}

	resp.Code = errcode.Success
	resp.Message = msg
	resp.Status = errcode.RequestSuccess

	return resp, nil
}

func failResponse(resp *types.Response, msg string, status, code int) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}

	resp.Code = code
	resp.Status = status
	resp.Message = msg

	return resp, nil
}

func failResponseWithErr(resp *types.Response, err error, status, code int) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}

	resp.Code = code
	resp.Status = status
	resp.Message = "服务失败：" + err.Error()

	return resp, nil
}
