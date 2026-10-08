// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"net/http"

	"Noviterra/server/app/user/userapi/internal/svc"
	"Noviterra/server/app/user/userapi/internal/types"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetUserInfoLogic {
	return &BatchGetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchGetUserInfoLogic) BatchGetUserInfo(req *types.BatchUserInfoRequest) (resp *types.BatchUserInfoResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.BatchUserInfoResponse)

	role := l.svcCtx.Middleware.GetUserRoleFromContext(l.ctx)
	if role == -1 {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = "未登录，身份验证失败！"
		return resp, nil
	}

	if role == 0 {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = "权限不足，无法查看用户信息！"
		return resp, nil
	}

	if req.Size < 5 || req.Size > 15 {
		req.Size = 10
		resp.Resp.Message = "每次批查询个数限制在5~15之间！"
	}

	if req.Page < 1 {
		req.Page = 1
		resp.Resp.Message = "页数不能小于1！"
	}

	rep, err := l.svcCtx.UserModel.GetMoreUsersInfo(l.ctx, &userrpc.GetMoreUsersInfoRequest{
		Size: int64(req.Size),
		Page: int64(req.Page),
	})
	if err != nil {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = "ERROR:" + err.Error()
		return resp, nil
	}
	if rep.Code != 0 {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = rep.Message
		return resp, nil
	}

	// 用户信息字段映射
	resp.Users = make([]*types.UserInfoResponse, 0, req.Size)

	for _, userInfo := range rep.UsersInfo {
		user := &types.UserInfoResponse{
			ID:               userInfo.Data.ID,
			Role:             int(userInfo.Data.Role),
			Username:         userInfo.Data.UserName,
			Email:            userInfo.Data.Email,
			Avatar:           userInfo.Data.Avatar,
			Gender:           int(userInfo.Data.Gender),
			Age:              int(userInfo.Data.Age),
			SelfIntroduction: userInfo.Data.SelfIntroduction,
		}
		resp.Users = append(resp.Users, user)
	}
	resp.Resp.Code = 0
	resp.Resp.Status = http.StatusOK
	resp.Resp.Message = "用户信息列表刷新成功！"

	return resp, nil
}
