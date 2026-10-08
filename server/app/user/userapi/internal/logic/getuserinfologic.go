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

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo() (resp *types.UserInfoResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.UserInfoResponse)

	userID := l.svcCtx.Middleware.GetUserIDFromContext(l.ctx)
	if userID == "" {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = "未登录，用户ID为空！"
		return resp, nil
	}

	rep, err := l.svcCtx.UserModel.GetSelfInfo(l.ctx, &userrpc.GetUserInfoRequest{
		ID: userID,
	})
	if err != nil {
		resp.Resp.Message = "ERROR:" + err.Error()
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		return resp, nil
	}
	if rep.Code != 0 {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = rep.Message
	}

	resp.Resp.Status = http.StatusOK
	resp.Resp.Code = 0
	resp.Resp.Message = "个人信息校验完毕！"

	resp.ID = rep.Data.ID
	resp.Email = rep.Data.Email
	resp.Avatar = rep.Data.Avatar
	resp.Username = rep.Data.UserName
	resp.Age = int(rep.Data.Age)
	resp.Gender = int(rep.Data.Gender)
	resp.SelfIntroduction = rep.Data.SelfIntroduction

	return resp, nil
}
