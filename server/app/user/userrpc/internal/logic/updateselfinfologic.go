package logic

import (
	"Noviterra/server/utils/common"
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"
)

type updateInfoUserRepository interface {
	UpdateUsersInfo(userInfo model.UserInfo) error
}

type updateInfoIdentityRepository interface {
	CheckIDExists(id string) (bool, error)
}

type updateInfoCommonUtils interface {
	ValidateUserInfo(in *common.ValidateUserInfoInput) (*common.ValidateUserInfoOutput, error)
}

type UpdateSelfInfoLogic struct {
	ctx                context.Context
	identityRespotiory updateInfoIdentityRepository
	userRepository     updateInfoUserRepository
	commonUtils        updateInfoCommonUtils
}

func NewUpdateSelfInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSelfInfoLogic {
	return &UpdateSelfInfoLogic{
		ctx:                ctx,
		identityRespotiory: svcCtx.IdentityRepository,
		userRepository:     svcCtx.UserRepository,
		commonUtils:        svcCtx.CommonUtils,
	}
}

func (l *UpdateSelfInfoLogic) UpdateSelfInfo(in *userrpc.UpdateUserInfoRequest) (*userrpc.Response, error) {
	// todo: add your logic here and delete this line
	// 校验用户ID是否存在
	// ：即使鉴权中间件做了对用户的校验 但是在rpc层仍需要校验数据 防止恶意请求绕过api层直接进入rpc层
	idExist, err := l.identityRespotiory.CheckIDExists(in.ID)
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "数据库操作异常！",
		}, err
	}
	if !idExist {
		return &userrpc.Response{
			Code:    1,
			Message: "用户账号不存在！",
		}, nil
	}

	userInfo, err := l.commonUtils.ValidateUserInfo(&common.ValidateUserInfoInput{
		Age:              int(in.Age),
		Name:             in.UserName,
		Gender:           int(in.Gender),
		SelfIntroduction: in.SelfIntroduction,
	})
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, err
	}

	// 更新用户数据
	err = l.userRepository.UpdateUsersInfo(model.UserInfo{
		ID:               in.ID,
		Name:             userInfo.Name,
		Age:              userInfo.Age,
		Gender:           userInfo.Gender,
		Avatar:           in.Avatar,
		SelfIntroduction: userInfo.SelfIntroduction,
	})
	if err != nil {
		return &userrpc.Response{
			Code:    1,
			Message: "",
		}, nil
	}

	// 缓存更新or删除

	return &userrpc.Response{
		Code:    0,
		Message: "用户信息更新成功！",
	}, nil
}
