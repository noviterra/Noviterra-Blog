// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/user/userapi/internal/svc"
	"Noviterra/server/app/user/userapi/internal/types"
	"Noviterra/server/app/user/userrpc/types/userrpc"
	"Noviterra/server/utils/errcode"
	"context"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type changPasswordByEmailUserModel interface {
	ChangeSelfPasswordByEmail(ctx context.Context, resp *userrpc.ChangePasswordByEmailRequest, opts ...grpc.CallOption) (*userrpc.Response, error)
}

type changePasswordByEmailCtxData interface {
	GetUserIDFromContext(ctx context.Context) string
}

type changePasswordByEmailServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
}

type ChangePasswordByEmailLogic struct {
	ctx                     context.Context
	userModel               changPasswordByEmailUserModel
	ctxdata                 changePasswordByEmailCtxData
	serviceTokenInterceptor changePasswordByEmailServiceTokenInterceptor
}

func NewChangePasswordByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordByEmailLogic {
	return &ChangePasswordByEmailLogic{
		ctx:                     ctx,
		userModel:               svcCtx.UserModel,
		ctxdata:                 svcCtx.CtxData,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *ChangePasswordByEmailLogic) ChangePasswordByEmail(req *types.ChangePasswordByEmailRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		return failResponse(resp, "未登录，身份验证失败！", errcode.Unauthorized, errcode.Failed)
	}

	// 需保证请求用户与被操作用户是同一个 则传入id
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", id,
	)

	rep, err := l.userModel.ChangeSelfPasswordByEmail(ctx, &userrpc.ChangePasswordByEmailRequest{
		ID:               id,
		Email:            req.Email,
		VerificationCode: req.VerificationCode,
		NewPassword:      req.NewPassword,
	})

	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	if rep.Code != 0 {
		return failResponse(resp, resp.Message, errcode.InvalidParam, errcode.Failed)
	}

	return successResponse(resp, "密码修改成功！")
}
