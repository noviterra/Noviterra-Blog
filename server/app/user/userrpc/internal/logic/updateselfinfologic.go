package logic

import (
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"

	"go.uber.org/zap"
)

type (
	updateInfoUserRepository interface {
		UpdateUsersInfo(userInfo model.UserInfo) error
	}
	updateInfoIdentityRepository interface {
		CheckIDExists(ctx context.Context, id string) (bool, error)
	}
	updateInfoUserValidate interface {
		ValidateUserInfo(in *validate.InputValidateUserInfo) (*validate.OutputValidateUserInfo, error)
	}
	updateInfoUserResponse interface {
		FailResponse(resp *userrpc.Response,
			msg string, status int32, code int64) (*userrpc.Response, error)
		FailResponseWithErr(resp *userrpc.Response,
			err error) (*userrpc.Response, error)
		SuccessResponse(resp *userrpc.Response,
			msg string) (*userrpc.Response, error)
	}
	updateInfoLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type UpdateSelfInfoLogic struct {
	ctx                context.Context
	identityRespotiory updateInfoIdentityRepository
	userRepository     updateInfoUserRepository
	userValidate       updateInfoUserValidate
	userResponse       updateInfoUserResponse
	logger             updateInfoLogger
}

func NewUpdateSelfInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSelfInfoLogic {
	return &UpdateSelfInfoLogic{
		ctx:                ctx,
		identityRespotiory: svcCtx.IdentityRepository,
		userRepository:     svcCtx.UserRepository,
		userValidate:       svcCtx.UserValidate,
		userResponse:       svcCtx.UserResponse,
		logger:             svcCtx.Logger,
	}
}

func (l *UpdateSelfInfoLogic) UpdateSelfInfo(in *userrpc.UpdateUserInfoRequest) (*userrpc.Response, error) {
	resp := new(userrpc.Response)

	// 校验用户ID是否存在
	// ：即使鉴权中间件做了对用户的校验 但是在rpc层仍需要校验数据 防止恶意请求绕过api层直接进入rpc层
	idExist, err := l.identityRespotiory.CheckIDExists(l.ctx, in.ID)
	if err != nil {
		l.logger.Error("数据库操作失败：校验用户ID是否存在失败",
			zap.Strings("user-info", []string{in.ID}),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.InterCallToolFail)
	}
	if !idExist {
		l.logger.Error("用户账号不存在",
			zap.Strings("user-info", []string{in.ID}),
		)
		return l.userResponse.FailResponse(resp, "用户账号不存在！",
			errcode.NotFound, errcode.GetUserInfoFail)
	}

	validateInput := &validate.InputValidateUserInfo{
		Age:              int(in.Age),
		Name:             in.UserName,
		Gender:           int(in.Gender),
		SelfIntroduction: in.SelfIntroduction,
	}
	userInfo, err := l.userValidate.ValidateUserInfo(validateInput)
	if err != nil {
		l.logger.Error("校验InputValidateUserInfo参数失败",
			zap.Strings("user-info", []string{in.ID}),
			zap.Any("validate", validateInput),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "服务失败："+err.Error(),
			errcode.InvalidParam, errcode.ValidateUserRequestParamFail)
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
		l.logger.Error("数据库操作失败：用户信息更新失败",
			zap.Strings("user-info", []string{in.ID}),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.InterCallToolFail)
	}

	// 缓存更新or删除

	l.logger.Info("用户信息更新成功",
		zap.Strings("user-info", []string{in.ID}),
	)
	return l.userResponse.SuccessResponse(resp, "用户信息更新成功！")
}
