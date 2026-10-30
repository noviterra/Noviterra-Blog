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

type unBlackListUserModel interface {
	UnBlackUser(ctx context.Context, in *userrpc.UnBlackUserRequest, opts ...grpc.CallOption) (*userrpc.Response, error)
}

type unBlackListUserAuthMiddleware interface {
	GetUserIDFromContext(ctx context.Context) string
	GetUserRoleFromContext(ctx context.Context) int
}

type unBlackListUserServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
}

type UnblacklistUserLogic struct {
	ctx                     context.Context
	userModel               unBlackListUserModel
	authMiddleware          unBlackListUserAuthMiddleware
	serviceTokenInterceptor unBlackListUserServiceTokenInterceptor
}

func NewUnblacklistUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnblacklistUserLogic {
	return &UnblacklistUserLogic{
		ctx:                     ctx,
		userModel:               svcCtx.UserModel,
		authMiddleware:          svcCtx.AuthMiddleware,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *UnblacklistUserLogic) UnblacklistUser(req *types.UnblacklistUserRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	// 通过鉴权中间件 从上下文中获取操作人员的ID与身份
	id := l.authMiddleware.GetUserIDFromContext(l.ctx)
	role := l.authMiddleware.GetUserRoleFromContext(l.ctx)
	if role == -1 || id == "" {
		return failResponse(resp, "未登录，身份验证失败！", errcode.Unauthorized, errcode.Failed)
	}
	if role < 1 {
		return failResponse(resp, "权限不足，无法解封用户！", errcode.Unauthorized, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", id,
	)

	rep, err := l.userModel.UnBlackUser(ctx, &userrpc.UnBlackUserRequest{
		ID:       req.ID,
		Operator: id,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Code != 0 {
		return failResponse(resp, rep.Message, errcode.InvalidParam, int(rep.Code))
	}

	return successResponse(resp, "用户解封成功！")
}
