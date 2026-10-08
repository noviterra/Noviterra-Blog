package logic

import (
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateSelfInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSelfInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSelfInfoLogic {
	return &UpdateSelfInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateSelfInfoLogic) UpdateSelfInfo(in *userrpc.UpdateUserInfoRequest) (*userrpc.UpdateUserInfoResponse, error) {
	// todo: add your logic here and delete this line
	// 校验用户ID是否存在
	// ：即使鉴权中间件做了对用户的校验 但是在rpc层仍需要校验数据 防止恶意请求绕过api层直接进入rpc层
	idExist, err := l.svcCtx.DB.CheckIDExists(in.ID)
	if err != nil {
		return &userrpc.UpdateUserInfoResponse{
			Code:    1,
			Message: "数据库操作异常！",
		}, err
	}
	if !idExist {
		return &userrpc.UpdateUserInfoResponse{
			Code:    1,
			Message: "用户账号不存在！",
		}, nil
	}

	userNameErr := l.svcCtx.CommonUtils.ValidateUserName(in.UserName)
	if userNameErr != nil {
		return &userrpc.UpdateUserInfoResponse{
			Code:    1,
			Message: "",
		}, userNameErr
	}

	ageErr := l.svcCtx.CommonUtils.ValidateAge(int(in.Age))
	if ageErr != nil {
		return &userrpc.UpdateUserInfoResponse{
			Code:    1,
			Message: "",
		}, ageErr
	}

	genderErr := l.svcCtx.CommonUtils.ValidateGender(int(in.Gender))
	if genderErr != nil {
		return &userrpc.UpdateUserInfoResponse{
			Code:    1,
			Message: "",
		}, genderErr
	}

	selfIntroductionErr := l.svcCtx.CommonUtils.ValidateSelfIntroduction(in.SelfIntroduction)
	if selfIntroductionErr != nil {
		return &userrpc.UpdateUserInfoResponse{
			Code:    1,
			Message: "",
		}, selfIntroductionErr
	}

	// 更新用户数据
	err = l.svcCtx.DB.UpdateUsersInfo(model.UserInfo{
		ID:               in.ID,
		Name:             in.UserName,
		Age:              int(in.Age),
		Gender:           int(in.Gender),
		Avatar:           in.Avatar,
		SelfIntroduction: in.SelfIntroduction,
	})
	if err != nil {
		return &userrpc.UpdateUserInfoResponse{
			Code:    1,
			Message: "",
		}, nil
	}

	return &userrpc.UpdateUserInfoResponse{
		Code:    0,
		Message: "用户信息更新成功！",
	}, nil
}
