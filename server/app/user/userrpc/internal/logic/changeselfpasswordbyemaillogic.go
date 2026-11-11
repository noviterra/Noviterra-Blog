package logic

import (
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"go.uber.org/zap"
)

type (
	changePwdByEmailUserResponse interface {
		FailResponse(resp *userrpc.Response,
			msg string, status int32, code int64) (*userrpc.Response, error)
		FailResponseWithErr(resp *userrpc.Response,
			err error) (*userrpc.Response, error)
		SuccessResponse(resp *userrpc.Response,
			msg string) (*userrpc.Response, error)
	}
	changePwdByEmailNoticeCache interface {
		GetVerificationCode(ctx context.Context, email, t string) (string, error)
	}
	changePwdByEmailIdentityRepository interface {
		GetUserEmailByID(ctx context.Context, id string) (string, error)
		UpdateUsersAccount(ctx context.Context, userAccount model.UserAccount) error
	}
	changePwdByEmailUserValidate interface {
		ValidateChangePwdByEmail(in validate.InputChangePwdByEmail) (validate.OutputChangePwdByEmail, error)
	}
	changePwdByEmailCrypt interface {
		HashPassword(password string) (string, error)
	}
	changePwdByEmailCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}
	changePwdByEmailLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type ChangeSelfPasswordByEmailLogic struct {
	ctx            context.Context
	noticeCache    changePwdByEmailNoticeCache
	userRepository changePwdByEmailIdentityRepository
	userValidate   changePwdByEmailUserValidate
	crypt          changePwdByEmailCrypt
	userResponse   changePwdByEmailUserResponse
	ctxdata        changePwdByEmailCtxData
	logger         changePwdByEmailLogger
}

func NewChangeSelfPasswordByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeSelfPasswordByEmailLogic {
	return &ChangeSelfPasswordByEmailLogic{
		ctx:            ctx,
		userRepository: svcCtx.IdentityRepository,
		noticeCache:    svcCtx.NoticeCache,
		crypt:          svcCtx.Crypt,
		userValidate:   svcCtx.UserValidate,
		userResponse:   svcCtx.UserResponse,
		ctxdata:        svcCtx.Ctxdata,
		logger:         svcCtx.Logger,
	}
}

func (l *ChangeSelfPasswordByEmailLogic) ChangeSelfPasswordByEmail(in *userrpc.ChangePasswordByEmailRequest) (*userrpc.Response, error) {
	resp := new(userrpc.Response)
	// 用户信息校验
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		l.logger.Error("从上下文中获取用户信息失败")
		return l.userResponse.FailResponse(resp, "信息校验失败！",
			errcode.Unauthorized, errcode.GetContextUserIDFail)
	}

	// 参数校验
	i := validate.InputChangePwdByEmail{
		ID:               in.ID,
		Email:            in.Email,
		VerificationCode: in.VerificationCode,
		NewPassword:      in.NewPassword,
	}
	param, err := l.userValidate.ValidateChangePwdByEmail(i)
	if err != nil {
		l.logger.Error("校验InputChangePwdByEmail参数失败",
			zap.Strings("user-info", []string{id}),
			zap.Any("validate", i), zap.Error(err))
		return l.userResponse.FailResponse(resp, "服务失败："+err.Error(),
			errcode.InvalidParam, errcode.ValidateUserRequestParamFail)
	}

	if id != param.ID { // 账号非本人操作
		l.logger.Error("请求发起者ID与被操作ID不同",
			zap.Strings("user-id||operator-id", []string{id, param.ID}))
		return l.userResponse.FailResponse(resp, "信息校验失败！",
			errcode.Forbidden, errcode.UserOperationOutOfBounds)
	}

	// 根据用户ID得到邮箱
	userEmail, err := l.userRepository.GetUserEmailByID(l.ctx, id)
	if err != nil {
		l.logger.Error("数据库操作：根据用户ID查询用户邮箱操作失败",
			zap.Strings("user-info", []string{id}),
			zap.Error(err))
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.GetUserInfoFail)
	}
	if userEmail != param.Email {
		l.logger.Error("传参的邮箱与用户绑定的邮箱不同",
			zap.Strings("user-info", []string{id, userEmail}),
			zap.String("param-email", param.Email))
		return l.userResponse.FailResponse(resp, "信息校验失败！",
			errcode.Forbidden, errcode.UserEmailMismatch)
	}

	// 验证码的获取与校验
	realVerificationCode, err := l.noticeCache.GetVerificationCode(l.ctx, userEmail,
		"change_password")
	if err != nil {
		l.logger.Error("缓存操作失败：查询验证码操作失败",
			zap.Strings("user-info", []string{id, userEmail}),
			zap.String("email-type", "change password"), zap.Error(err))
		return l.userResponse.FailResponse(resp, "验证码获取失败！",
			errcode.InternalError, errcode.GetVerificationCodeFail)
	}
	if realVerificationCode == "" {
		l.logger.Error("读取到的验证码为空",
			zap.Strings("user-info", []string{id, userEmail}),
		)
		return l.userResponse.FailResponse(resp, "验证码不存在，请检查邮箱是否正确！",
			errcode.InvalidParam, errcode.VerificationCodeNotFound)
	}
	if realVerificationCode != param.VerificationCode {
		l.logger.Error("输入验证码与邮件发送验证码不同",
			zap.Strings("user-info", []string{id, userEmail}),
			zap.String("input-verification", param.VerificationCode),
		)
		return l.userResponse.FailResponse(resp, "验证码输入有误！",
			errcode.InvalidParam, errcode.ValidateVerificationCodeFail)
	}

	// 密码加密
	hashPwd, err := l.crypt.HashPassword(in.NewPassword)
	if err != nil {
		l.logger.Error("密码加密失败",
			zap.Strings("user-info", []string{id, userEmail}),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.HashPasswordFail)
	}

	// 更新数据库
	err = l.userRepository.UpdateUsersAccount(l.ctx, model.UserAccount{
		ID:       in.ID,
		Password: hashPwd,
	})
	if err != nil {
		l.logger.Error("数据库操作失败：用户密码更新失败",
			zap.Strings("user-info", []string{id, userEmail}),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.UserUpdatePasswordFail)
	}

	l.logger.Info("用户通过邮箱辅助账号密码修改成功",
		zap.Strings("user-info", []string{id, userEmail}),
	)
	return l.userResponse.SuccessResponse(resp, "密码修改成功！")
}
