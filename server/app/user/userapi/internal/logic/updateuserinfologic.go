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

type updateUserInfoUserModel interface {
	UpdateSelfInfo(ctx context.Context, in *userrpc.UpdateUserInfoRequest, opts ...grpc.CallOption) (*userrpc.Response, error)
}

type updateUserInfoAuthMiddleware interface {
	GetUserIDFromContext(ctx context.Context) string
}

type updateUserServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
}

type UpdateUserInfoLogic struct {
	ctx                     context.Context
	userModel               updateUserInfoUserModel
	authMiddleware          updateUserInfoAuthMiddleware
	serviceTokenInterceptor updateUserServiceTokenInterceptor
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		ctx:                     ctx,
		userModel:               svcCtx.UserModel,
		authMiddleware:          svcCtx.AuthMiddleware,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *UpdateUserInfoLogic) UpdateUserInfo(req *types.UpdateUserInfoRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	// 从上下文中获取用户ID 经过了鉴权中间件获取
	id := l.authMiddleware.GetUserIDFromContext(l.ctx)
	if id == "" {
		return failResponse(resp, "未登录，身份验证失败！", errcode.Unauthorized, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", id,
	)

	rep, err := l.userModel.UpdateSelfInfo(ctx, &userrpc.UpdateUserInfoRequest{
		ID:               id,
		UserName:         req.Username,
		Avatar:           req.Avatar,
		Age:              int64(req.Age),
		Gender:           int64(req.Gender),
		SelfIntroduction: req.SelfIntroduction,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Code != 0 {
		return failResponse(resp, rep.Message, errcode.InvalidParam, int(rep.Code))
	}

	return successResponse(resp, "文章信息更新成功！")
}
