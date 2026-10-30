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

type blackListUserModel interface {
	BlackUser(ctx context.Context, in *userrpc.BlackUserRequest, opts ...grpc.CallOption) (*userrpc.Response, error)
}

type blackListUserAuthMiddleware interface {
	GetUserIDFromContext(ctx context.Context) string
	GetUserRoleFromContext(ctx context.Context) int
}

type blackListUserServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
}

type BlacklistUserLogic struct {
	ctx                     context.Context
	userModel               blackListUserModel
	authMiddleware          blackListUserAuthMiddleware
	serviceTokenInterceptor blackListUserServiceTokenInterceptor
}

func NewBlacklistUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlacklistUserLogic {
	return &BlacklistUserLogic{
		ctx:                     ctx,
		userModel:               svcCtx.UserModel,
		authMiddleware:          svcCtx.AuthMiddleware,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

// 拉黑用户 管理员操作
func (l *BlacklistUserLogic) BlacklistUser(req *types.BlacklistUserRequest) (resp *types.Response, err error) {
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
		return failResponse(resp, "权限不足，无法封禁用户！", errcode.Forbidden, errcode.Failed)
	}

	// metadata数据
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", id,
	)

	rep, err := l.userModel.BlackUser(ctx, &userrpc.BlackUserRequest{
		ID:       req.ID,
		Operator: id,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Code != 0 {
		return failResponse(resp, rep.Message, errcode.InvalidParam, int(rep.Code))
	}

	return successResponse(resp, "用户拉黑成功！")
}

func successResponse(resp *types.Response, msg string) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}
	resp.Code = errcode.Success
	resp.Status = errcode.RequestSuccess
	resp.Message = msg

	return resp, nil
}

func failResponse(resp *types.Response, msg string, status, code int) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}
	resp.Status = status
	resp.Code = code
	resp.Message = msg

	return resp, nil
}

func failResponseWithErr(resp *types.Response, err error, status, code int) (*types.Response, error) {
	if resp == nil {
		resp = new(types.Response)
	}
	resp.Status = status
	resp.Code = code
	resp.Message = "服务错误：" + err.Error()

	return resp, nil
}
