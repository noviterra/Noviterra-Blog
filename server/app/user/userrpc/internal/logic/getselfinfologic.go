package logic

import (
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"
	"Noviterra/server/utils/errcode"

	"go.uber.org/zap"
)

type (
	getSelfInfoUserCache interface {
		HGetUserInfo(ctx context.Context, id string) (*model.UserInfo, error)
		HSetUserInfo(ctx context.Context, userInfo *model.UserInfo) error
	}
	getSelfInfoUserRepository interface {
		GetUserInfoByID(id string) (*model.UserInfo, error)
	}
	getSelfInfoUserResponse interface {
		SuccessGetSelfInfoResponse(resp *userrpc.UserInfoResponse,
			msg string, data *model.UserInfo) (*userrpc.UserInfoResponse, error)
		FailGetSelfInfoResponse(resp *userrpc.UserInfoResponse,
			msg string, status int32, code int64) (*userrpc.UserInfoResponse, error)
		FailGetSelfInfoResponseWithErr(resp *userrpc.UserInfoResponse,
			err error) (*userrpc.UserInfoResponse, error)
	}
	getSelfInfoLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
	getSelfInfoCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}
	getSelfInfoIdentityValidate interface {
		ValidateUserID(id string) (string, error)
	}
)

type GetSelfInfoLogic struct {
	ctx              context.Context
	userCache        getSelfInfoUserCache
	userRepository   getSelfInfoUserRepository
	userResponse     getSelfInfoUserResponse
	logger           getSelfInfoLogger
	ctxdata          getSelfInfoCtxData
	identityValidate getSelfInfoIdentityValidate
}

func NewGetSelfInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSelfInfoLogic {
	return &GetSelfInfoLogic{
		ctx:              ctx,
		userCache:        svcCtx.UserCache,
		userRepository:   svcCtx.UserRepository,
		userResponse:     svcCtx.UserResponse,
		logger:           svcCtx.Logger,
		ctxdata:          svcCtx.Ctxdata,
		identityValidate: svcCtx.IdentityValidate,
	}
}

// 用户个人操作
func (l *GetSelfInfoLogic) GetSelfInfo(in *userrpc.GetUserInfoRequest) (*userrpc.UserInfoResponse, error) {
	resp := new(userrpc.UserInfoResponse)

	// 只有登录后才可以查看其他人或个人信息
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		l.logger.Error("上下文获取操作用户ID失败")
		return l.userResponse.FailGetSelfInfoResponse(resp, "登陆后解锁完整功能！",
			errcode.Unauthorized, errcode.GetUserInfoFail)
	}

	// 参数校验
	userID, err := l.identityValidate.ValidateUserID(in.ID)
	if err != nil {
		l.logger.Error("校验用户ID参数失败",
			zap.String("validate", in.ID),
			zap.Error(err),
		)
		return l.userResponse.FailGetSelfInfoResponse(resp, "信息校验失败 ",
			errcode.InvalidParam, errcode.ValidateUserInfoFail)
	}

	// 查看缓存中是否有用户数据
	cacheUser, err := l.userCache.HGetUserInfo(l.ctx, userID)
	if err != nil {
		l.logger.Error("缓存操作失败：查询用户缓存信息失败",
			zap.Strings("user-info", []string{userID}),
			zap.Error(err),
		)
		return l.userResponse.FailGetSelfInfoResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.InterCallToolFail)
	}
	// 缓存命中
	if cacheUser != nil {
		l.logger.Info("缓存命中，获取用户信息成功",
			zap.Strings("user-info", []string{userID}),
		)
		return l.userResponse.SuccessGetSelfInfoResponse(resp,
			"用户信息加载完成！", cacheUser)
	}

	// 缓存未命中 查库
	userInfo, err := l.userRepository.GetUserInfoByID(userID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据用户ID查询用户信息失败",
			zap.Strings("user-info", []string{userID}),
			zap.Error(err),
		)
		return l.userResponse.FailGetSelfInfoResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.GetUserInfoFail)
	}
	if userInfo == nil {
		l.logger.Error("根据用户ID查询到的用户信息为空，用户账号不存在",
			zap.Strings("user-info", []string{userID}),
		)
		return l.userResponse.FailGetSelfInfoResponse(resp, "无账号相关信息，请联系管理员！",
			errcode.NotFound, errcode.GetUserInfoFail)
	}

	// 将用户信息写入缓存中
	err = l.userCache.HSetUserInfo(l.ctx, userInfo)
	if err != nil {
		l.logger.Error("缓存操作失败：用户信息写入缓存失败",
			zap.Strings("user-info", []string{userID}),
			zap.Error(err),
		)
	}

	l.logger.Info("检索到用户信息",
		zap.Strings("user-info", []string{userID}),
		zap.Any("inf0", userInfo),
	)
	return l.userResponse.SuccessGetSelfInfoResponse(resp, "用户信息加载完成！", userInfo)
}
