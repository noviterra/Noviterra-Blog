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

type batchGetUserInfoCtxData interface {
	GetUserRoleFromContext(ctx context.Context) int
	GetUserIDFromContext(ctx context.Context) string
}

type batchGetUserInfoUserModel interface {
	GetMoreUsersInfo(ctx context.Context, in *userrpc.GetMoreUsersInfoRequest, opts ...grpc.CallOption) (*userrpc.GetMoreUsersInfoResponse, error)
}

type batchGetUserInfoServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
}

type BatchGetUserInfoLogic struct {
	ctx                     context.Context
	ctxdata                 batchGetUserInfoCtxData
	userModel               batchGetUserInfoUserModel
	serviceTokenInterceptor batchGetUserInfoServiceTokenInterceptor
}

func NewBatchGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetUserInfoLogic {
	return &BatchGetUserInfoLogic{
		ctx:                     ctx,
		ctxdata:                 svcCtx.CtxData,
		userModel:               svcCtx.UserModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *BatchGetUserInfoLogic) BatchGetUserInfo(req *types.BatchUserInfoRequest) (resp *types.BatchUserInfoResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.BatchUserInfoResponse)
	// 管理员操作

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return l.failBatchUserInfoResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
	}

	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	role := l.ctxdata.GetUserRoleFromContext(l.ctx)
	if role == -1 || id == "" {
		return l.failBatchUserInfoResponse(resp, "未登录，身份验证失败！", errcode.InvalidParam, errcode.Failed)
	}
	if role < 1 {
		return l.failBatchUserInfoResponse(resp, "权限不足，无法查看用户信息！", errcode.InvalidParam, errcode.Failed)
	}

	// 将用户信息与服务凭证注入到metadata中
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", id,
	)

	rep, err := l.userModel.GetMoreUsersInfo(ctx, &userrpc.GetMoreUsersInfoRequest{
		Size: int64(req.Size),
		Page: int64(req.Page),
	})
	if err != nil {
		return l.failBatchUserInfoResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return l.failBatchUserInfoResponse(resp, rep.Resp.Message, errcode.InvalidParam, int(rep.Resp.Code))
	}

	return l.successBatchUserInfoResponse(resp, rep)
}

func (l *BatchGetUserInfoLogic) failBatchUserInfoResponse(resp *types.BatchUserInfoResponse, msg string, status, code int) (*types.BatchUserInfoResponse, error) {
	if resp == nil {
		resp = new(types.BatchUserInfoResponse)
	}

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: msg,
	}
	resp.Users = nil

	return resp, nil
}

func (l *BatchGetUserInfoLogic) failBatchUserInfoResponseWithErr(resp *types.BatchUserInfoResponse, err error, status, code int) (*types.BatchUserInfoResponse, error) {
	if resp == nil {
		resp = new(types.BatchUserInfoResponse)
	}

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "服务错误：" + err.Error(),
	}
	resp.Users = nil

	return resp, nil
}

func (l *BatchGetUserInfoLogic) successBatchUserInfoResponse(resp *types.BatchUserInfoResponse, rep *userrpc.GetMoreUsersInfoResponse) (*types.BatchUserInfoResponse, error) {
	if resp == nil {
		resp = new(types.BatchUserInfoResponse)
	}

	// 用户信息字段映射
	resp.Users = make([]types.UserInfo, 0)
	for _, userInfo := range rep.UsersInfo {
		user := types.UserInfo{
			ID:               userInfo.ID,
			Role:             int(userInfo.Role),
			Username:         userInfo.UserName,
			Email:            userInfo.Email,
			Avatar:           userInfo.Avatar,
			Gender:           int(userInfo.Gender),
			Age:              int(userInfo.Age),
			SelfIntroduction: userInfo.SelfIntroduction,
		}
		resp.Users = append(resp.Users, user)
	}

	resp.Resp = types.Response{
		Status:  errcode.RequestSuccess,
		Code:    errcode.Success,
		Message: "用户信息获取成功！",
	}

	return resp, nil
}
