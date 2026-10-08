// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"net/http"

	"Noviterra/server/app/user/userapi/internal/svc"
	"Noviterra/server/app/user/userapi/internal/types"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/go-playground/validator/v10"
	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshAccessTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshAccessTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshAccessTokenLogic {
	return &RefreshAccessTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshAccessTokenLogic) RefreshAccessToken(req *types.RefreshAccessTokenRequest) (resp *types.RefreshAccessTokenResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.RefreshAccessTokenResponse)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			resp.Resp.Message = TypeName[errs[0].Field()] + "格式错误！"
			resp.Resp.Status = http.StatusInternalServerError
			resp.Resp.Code = 1
		}
		return resp, nil
	}

	// 根据RefreshToken刷新AccessToken RefreshToken也刷新
	rep, err := l.svcCtx.UserModel.RefreshAccessToken(l.ctx, &userrpc.RefreshAccessTokenARequest{
		RefreshToken: req.RefreshToken,
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

	resp.Resp.Status = http.StatusOK
	resp.Resp.Code = 0
	resp.Resp.Message = "Token刷新成功！"
	resp.AccessToken = rep.AccessToken
	resp.RefreshToken = rep.RefreshToken
	return resp, nil
}
