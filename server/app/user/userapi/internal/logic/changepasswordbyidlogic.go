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

type changePasswordByIDUserModel interface {
	ChangeSelfPasswordByID(ctx context.Context, in *userrpc.ChangePasswordRequest, opts ...grpc.CallOption) (*userrpc.Response, error)
}

type changePasswordByIDAuthMiddleware interface {
	GetUserIDFromContext(ctx context.Context) string
}

type changePasswordByIDServiceTokenIntercptor interface {
	GetServiceTokenInConf() string
}

type ChangePasswordByIDLogic struct {
	ctx                     context.Context
	userModel               changePasswordByIDUserModel
	authMiddleware          changePasswordByIDAuthMiddleware
	serviceTokenInterceptor changePasswordByIDServiceTokenIntercptor
}

func NewChangePasswordByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordByIDLogic {
	return &ChangePasswordByIDLogic{
		ctx:                     ctx,
		userModel:               svcCtx.UserModel,
		authMiddleware:          svcCtx.AuthMiddleware,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *ChangePasswordByIDLogic) ChangePasswordByID(req *types.ChangePasswordByIDRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	id := l.authMiddleware.GetUserIDFromContext(l.ctx)
	if id == "" {
		return failResponse(resp, "未登录，身份验证失败！", errcode.Unauthorized, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", id,
	)

	rep, err := l.userModel.ChangeSelfPasswordByID(ctx, &userrpc.ChangePasswordRequest{
		ID:          id,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	if rep.Code != 0 {
		return failResponse(resp, resp.Message, errcode.InvalidParam, errcode.Failed)
	}

	return successResponse(resp, "密码修改成功！")
}
