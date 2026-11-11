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
	updateInfoIdentityValidate interface {
		ValidateUserID(id string) (string, error)
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
	updateInfoUserCache interface {
		DelUserInfo(ctx context.Context, id string) error
	}
	updateInfoCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}
)

type UpdateSelfInfoLogic struct {
	ctx                context.Context
	identityRespotiory updateInfoIdentityRepository
	userRepository     updateInfoUserRepository
	userValidate       updateInfoUserValidate
	userResponse       updateInfoUserResponse
	logger             updateInfoLogger
	userCache          updateInfoUserCache
	ctxdata            updateInfoCtxData
	identityValidate   updateInfoIdentityValidate
}

func NewUpdateSelfInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSelfInfoLogic {
	return &UpdateSelfInfoLogic{
		ctx:                ctx,
		identityRespotiory: svcCtx.IdentityRepository,
		userRepository:     svcCtx.UserRepository,
		userValidate:       svcCtx.UserValidate,
		userResponse:       svcCtx.UserResponse,
		logger:             svcCtx.Logger,
		userCache:          svcCtx.UserCache,
		ctxdata:            svcCtx.Ctxdata,
		identityValidate:   svcCtx.IdentityValidate,
	}
}

func (l *UpdateSelfInfoLogic) UpdateSelfInfo(in *userrpc.UpdateUserInfoRequest) (*userrpc.Response, error) {
	resp := new(userrpc.Response)

	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		l.logger.Error("上下文获取操作用户ID失败")
		return l.userResponse.FailResponse(resp, "身份校验失败！",
			errcode.Unauthorized, errcode.GetContextUserIDFail)
	}

	// 参数校验
	userID, err := l.identityValidate.ValidateUserID(in.ID)
	if err != nil {
		l.logger.Error("校验用户ID参数失败",
			zap.String("validate", in.ID),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "信息校验失败！",
			errcode.InvalidParam, errcode.ValidateUserInfoFail)
	}
	if id != userID {
		l.logger.Error("请求发起者ID与被操作ID不同",
			zap.Strings("user-id||operator-id", []string{id, userID}),
		)
		return l.userResponse.FailResponse(resp, "信息校验失败！",
			errcode.Forbidden, errcode.UserOperationOutOfBounds)
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
			zap.Strings("user-info", []string{userID}),
			zap.Any("validate", validateInput),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "服务失败："+err.Error(),
			errcode.InvalidParam, errcode.ValidateUserRequestParamFail)
	}

	// 校验用户ID是否存在
	idExist, err := l.identityRespotiory.CheckIDExists(l.ctx, userID)
	if err != nil {
		l.logger.Error("数据库操作失败：校验用户ID是否存在失败",
			zap.Strings("user-info", []string{userID}),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.GetUserInfoFail)
	}
	if !idExist {
		l.logger.Error("用户账号不存在",
			zap.Strings("user-info", []string{userID}),
		)
		return l.userResponse.FailResponse(resp, "用户账号不存在！",
			errcode.NotFound, errcode.UserNotFound)
	}

	// 更新用户数据
	err = l.userRepository.UpdateUsersInfo(model.UserInfo{
		ID:               userID,
		Name:             userInfo.Name,
		Age:              userInfo.Age,
		Gender:           userInfo.Gender,
		Avatar:           in.Avatar,
		SelfIntroduction: userInfo.SelfIntroduction,
	})
	if err != nil {
		l.logger.Error("数据库操作失败：用户信息更新失败",
			zap.Strings("user-info", []string{userID}),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.UpdateUserInfoFail)
	}

	// 删除缓存
	err = l.userCache.DelUserInfo(l.ctx, userID)
	if err != nil {
		l.logger.Error("缓存操作失败：删除用户缓存信息失败",
			zap.Strings("user-info", []string{userID}),
			zap.Error(err),
		)
	}

	l.logger.Info("用户信息更新成功",
		zap.Strings("user-info", []string{userID}),
	)
	return l.userResponse.SuccessResponse(resp, "用户信息更新成功！")
}
