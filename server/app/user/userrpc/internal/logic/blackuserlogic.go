package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"
	"context"

	"go.uber.org/zap"
)

type (
	blackUserIdentityRepository interface {
		CheckIDExists(ctx context.Context, id string) (bool, error)
		UpdateUsersAccount(ctx context.Context,
			userAccount model.UserAccount) error
		GetUserRole(ctx context.Context, id string) (int, error)
	}
	blackUserResponse interface {
		FailResponse(resp *userrpc.Response,
			msg string, status int32, code int64) (*userrpc.Response, error)
		FailResponseWithErr(resp *userrpc.Response,
			err error) (*userrpc.Response, error)
		SuccessResponse(resp *userrpc.Response,
			msg string) (*userrpc.Response, error)
	}
	blackUserLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
		Warn(msg string, fields ...zap.Field)
	}
	blackUserCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}
	blackUserValidate interface {
		ValidateBlackUser(in validate.InputBlackUser) (validate.OutputBlackUser, error)
	}
)

type BlackUserLogic struct {
	ctx                context.Context
	identityRepository blackUserIdentityRepository
	userResponse       blackUserResponse
	logger             blackUserLogger
	ctxdata            blackUserCtxData
	userValidate       blackUserValidate
}

func NewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlackUserLogic {
	return &BlackUserLogic{
		ctx:                ctx,
		identityRepository: svcCtx.IdentityRepository,
		ctxdata:            svcCtx.Ctxdata,
		userValidate:       svcCtx.UserValidate,
		userResponse:       svcCtx.UserResponse,
		logger:             svcCtx.Logger,
	}
}

// 管理员对用户操作
func (l *BlackUserLogic) BlackUser(in *userrpc.BlackUserRequest) (*userrpc.Response, error) {
	resp := new(userrpc.Response)
	// 上下文中获取用户ID
	operatorID := l.ctxdata.GetUserIDFromContext(l.ctx)
	if operatorID == "" {
		l.logger.Error("上下文获取操作用户ID失败")
		return l.userResponse.FailResponse(resp, "身份校验失败！",
			errcode.InvalidParam, errcode.ValidateUserInfoFail)
	}

	// 参数校验
	i := validate.InputBlackUser{
		ID:         in.ID,
		OperatorID: in.Operator,
	}
	param, err := l.userValidate.ValidateBlackUser(i)
	if err != nil {
		l.logger.Error("校验InputBlack参数失败",
			zap.Any("validate", i),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "服务失败："+err.Error(),
			errcode.InvalidParam, errcode.ValidateUserRequestParamFail)
	}

	if param.OperatorID != operatorID {
		l.logger.Error("上下文读取到的操作者ID与实际传参操作者ID不同",
			zap.Strings("param-operatorID||ctxdata-operatorID", []string{
				param.OperatorID, operatorID,
			}),
		)
		return l.userResponse.FailResponse(resp, "信息校验失败！",
			errcode.Forbidden, errcode.UserOperationOutOfBounds)
	}

	// 校验操作者ID与用户ID的有效性
	operatorRole, err := l.identityRepository.GetUserRole(l.ctx, operatorID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据操作者ID查询用户身份权限等级失败",
			zap.String("operator-info", operatorID),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.InterCallToolFail)
	}
	if operatorRole < 2 {
		l.logger.Error("操作者权限不足",
			zap.String("operator-info", operatorID),
			zap.Int("operator-role", operatorRole),
		)
		return l.userResponse.FailResponse(resp, "权限不足！",
			errcode.Forbidden, errcode.UserOperationOutOfBounds)
	}

	// 不能对自己操作
	if param.ID == operatorID {
		l.logger.Error("操作者预对自己账号进行操作",
			zap.String("operator-info", operatorID),
		)
		return l.userResponse.FailResponse(resp, "不能对自己账号操作！",
			errcode.Forbidden, errcode.UserOperationOutOfBounds)
	}

	idExist, err := l.identityRepository.CheckIDExists(l.ctx, param.ID)
	if err != nil {
		l.logger.Error("数据库操作失败：查询预被封禁用户账号ID失败",
			zap.String("black-user-id", param.ID),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.Failed)
	}
	if !idExist {
		l.logger.Error("要被封禁的用户账号不存在",
			zap.String("black-user-id", param.ID),
		)
		return l.userResponse.FailResponse(resp, "要封禁的账号不存在！",
			errcode.InternalError, errcode.Failed)
	}

	// 同等级不能封禁
	blackUserRole, err := l.identityRepository.GetUserRole(l.ctx, param.ID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据预被封禁用户ID查询用户身份权限等级失败",
			zap.String("black-user-info", param.ID),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.InterCallToolFail)
	}
	if blackUserRole >= operatorRole {
		l.logger.Error("要封禁的用户权限与操作者权限相同",
			zap.String("operator-info", operatorID),
			zap.Int("operator-role", operatorRole),
			zap.String("black-user-info", param.ID),
			zap.Int("black-user-role", blackUserRole),
		)
		return l.userResponse.FailResponse(resp, "对方权限与你相同或更高！",
			errcode.Forbidden, errcode.UserOperationOutOfBounds)
	}

	// 修改数据库字段
	err = l.identityRepository.UpdateUsersAccount(l.ctx, model.UserAccount{
		ID:     in.ID,
		Status: 1,
	})
	if err != nil {
		l.logger.Error("数据库操作失败：软删除被拉黑用户账号失败",
			zap.String("black-user-info", param.ID),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "用户账号封禁失败！",
			errcode.InternalError, errcode.InterCallToolFail)
	}

	l.logger.Info("封禁用户成功",
		zap.String("operator-info", operatorID),
		zap.Int("operator-role", operatorRole),
		zap.String("black-user-info", param.ID),
		zap.Int("black-user-role", blackUserRole),
	)
	return l.userResponse.SuccessResponse(resp, "账号封禁成功！")
}
