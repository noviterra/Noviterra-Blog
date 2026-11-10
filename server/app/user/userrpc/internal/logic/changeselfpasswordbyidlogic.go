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
	changePwdByIDUserRepository interface {
		GetUserPassword(id string) (string, error)
	}
	changePwdByIDIdentityRepository interface {
		UpdateUsersAccount(ctx context.Context,
			userAccount model.UserAccount) error
	}
	changePwdByIDCrypt interface {
		CheckPasswords(password, hashPassword string) (bool, error)
		HashPassword(password string) (string, error)
	}
	changePwdByIDIdentityValidate interface {
		ValidatePasswordFormat(password string) (string, error)
	}
	changePwdByIDUserResponse interface {
		FailResponse(resp *userrpc.Response,
			msg string, status int32, code int64) (*userrpc.Response, error)
		FailResponseWithErr(resp *userrpc.Response,
			err error) (*userrpc.Response, error)
		SuccessResponse(resp *userrpc.Response,
			msg string) (*userrpc.Response, error)
	}
	changePwdByIDLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type ChangeSelfPasswordByIDLogic struct {
	ctx                context.Context
	userRepository     changePwdByIDUserRepository
	identityRepository changePwdByIDIdentityRepository
	crypt              changePwdByIDCrypt
	identityValidate   changePwdByIDIdentityValidate
	userResponse       changePwdByIDUserResponse
	logger             changePwdByIDLogger
}

func NewChangeSelfPasswordByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeSelfPasswordByIDLogic {
	return &ChangeSelfPasswordByIDLogic{
		ctx:                ctx,
		identityRepository: svcCtx.IdentityRepository,
		userRepository:     svcCtx.UserRepository,
		crypt:              svcCtx.Crypt,
		identityValidate:   svcCtx.IdentityValidate,
		userResponse:       svcCtx.UserResponse,
		logger:             svcCtx.Logger,
	}
}

func (l *ChangeSelfPasswordByIDLogic) ChangeSelfPasswordByID(in *userrpc.ChangePasswordRequest) (*userrpc.Response, error) {
	resp := new(userrpc.Response)

	// 根据用户ID返回Password 为空则表示账号不存在
	hashPwd, err := l.userRepository.GetUserPassword(in.ID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据用户ID查询用户密码失败",
			zap.Strings("user-info", []string{in.ID}),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.GetUserInfoFail)
	}
	if hashPwd == "" {
		l.logger.Error("根据用户ID查询到的用户密码为空，用户账号不存在",
			zap.Strings("user-info", []string{in.ID}),
		)
		return l.userResponse.FailResponse(resp, "用户ID不存在！",
			errcode.NotFound, errcode.GetUserInfoFail)
	}

	samePwd, err := l.crypt.CheckPasswords(in.OldPassword, hashPwd)
	if err != nil {
		l.logger.Error("密码校验失败：原始密码校验操作失败",
			zap.Strings("user-info", []string{in.ID}),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.InterCallToolFail)
	}
	if !samePwd {
		l.logger.Error("原始密码输入错误",
			zap.Strings("user-info", []string{in.ID}),
		)
		return l.userResponse.FailResponse(resp, "原始密码输入错误！",
			errcode.InvalidParam, errcode.ValidateUserInfoFail)
	}

	password, err := l.identityValidate.ValidatePasswordFormat(in.NewPassword)
	if err != nil {
		l.logger.Error("校验新密码格式失败",
			zap.Strings("user-info", []string{in.ID}),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "服务失败："+err.Error(),
			errcode.InvalidParam, errcode.ValidateUserRequestParamFail)
	}

	// 密码加密
	hashPwd1, err := l.crypt.HashPassword(password)
	if err != nil {
		l.logger.Error("密码加密失败",
			zap.Strings("user-info", []string{in.ID}),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.InterCallToolFail)
	}

	err = l.identityRepository.UpdateUsersAccount(l.ctx, model.UserAccount{
		ID:       in.ID,
		Password: hashPwd1,
	})
	if err != nil {
		l.logger.Error("数据库操作失败：用户密码更新失败",
			zap.Strings("user-info", []string{in.ID}),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.InterCallToolFail)
	}

	l.logger.Info("用户通过ID校验修改账号密码成功",
		zap.Strings("user-info", []string{in.ID}),
	)
	return l.userResponse.SuccessResponse(resp, "密码修改成功！")
}
