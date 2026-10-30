// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/user/userapi/internal/svc"
	"Noviterra/server/app/user/userapi/internal/types"
	"Noviterra/server/app/user/userrpc/types/userrpc"
	"Noviterra/server/utils/errcode"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type getUserInfoAuthMiddleware interface {
	GetUserIDFromContext(ctx context.Context) string
}

type getUserInfoUserModel interface {
	GetSelfInfo(ctx context.Context, in *userrpc.GetUserInfoRequest, opts ...grpc.CallOption) (*userrpc.UserInfoResponse, error)
}

type getUserInfoServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
	GetSkipTokenInConf() string
}
type GetUserInfoLogic struct {
	ctx                     context.Context
	userModel               getUserInfoUserModel
	authMiddleware          getUserInfoAuthMiddleware
	serviceTokenInterceptro getUserInfoServiceTokenInterceptor
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:                     ctx,
		userModel:               svcCtx.UserModel,
		authMiddleware:          svcCtx.AuthMiddleware,
		serviceTokenInterceptro: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *GetUserInfoLogic) GetUserInfo() (resp *types.UserInfoResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.UserInfoResponse)

	userID := l.authMiddleware.GetUserIDFromContext(l.ctx)
	if userID == "" {
		return l.failUserInfoResponse(resp, "未登录，身份验证失败！", errcode.InvalidParam, errcode.Failed)
	}

	// 注入服务凭证
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptro.GetServiceTokenInConf(),
		"skip-token", l.serviceTokenInterceptro.GetSkipTokenInConf(),
	)

	rep, err := l.userModel.GetSelfInfo(ctx, &userrpc.GetUserInfoRequest{
		ID: userID,
	})
	if err != nil {
		return l.failUserInfoResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return l.failUserInfoResponse(resp, rep.Resp.Message, errcode.InvalidParam, int(rep.Resp.Code))
	}

	return l.SuccessUserInfoResponse(resp, rep)
}

func (l *GetUserInfoLogic) failUserInfoResponse(resp *types.UserInfoResponse, msg string, status, code int) (*types.UserInfoResponse, error) {
	if resp == nil {
		resp = new(types.UserInfoResponse)
	}

	resp.Info = types.UserInfo{}

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: msg,
	}

	return resp, nil
}

func (l *GetUserInfoLogic) failUserInfoResponseWithErr(resp *types.UserInfoResponse, err error, status, code int) (*types.UserInfoResponse, error) {
	if resp == nil {
		resp = new(types.UserInfoResponse)
	}

	resp.Info = types.UserInfo{}

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "服务失败：" + err.Error(),
	}

	return resp, nil
}

func (l *GetUserInfoLogic) SuccessUserInfoResponse(resp *types.UserInfoResponse, rep *userrpc.UserInfoResponse) (*types.UserInfoResponse, error) {
	if resp == nil {
		resp = new(types.UserInfoResponse)
	}

	resp.Info.ID = rep.Data.ID
	resp.Info.Age = int(rep.Data.Age)
	resp.Info.Email = rep.Data.Email
	resp.Info.Avatar = rep.Data.Avatar
	resp.Info.Gender = int(rep.Data.Gender)
	resp.Info.SelfIntroduction = rep.Data.SelfIntroduction
	resp.Info.Role = int(rep.Data.Role)
	resp.Info.Username = rep.Data.UserName

	resp.Resp = types.Response{
		Status:  errcode.RequestSuccess,
		Code:    errcode.Success,
		Message: "用户列表加载完成！",
	}

	return resp, nil
}
