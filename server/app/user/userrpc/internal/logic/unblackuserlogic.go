package logic

import (
	"Noviterra/server/utils/validate"
	"context"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"
	"Noviterra/server/utils/errcode"

	"go.uber.org/zap"
)

type (
	unBlackUserIdentityRepository interface {
		CheckIDExists(ctx context.Context, id string) (bool, error)
		UpdateUsersAccount(ctx context.Context,
			userAccount model.UserAccount) error
		GetUserRole(ctx context.Context, id string) (int, error)
	}
	unBlackUserResponse interface {
		FailResponse(resp *userrpc.Response,
			msg string, status int32, code int64) (*userrpc.Response, error)
		FailResponseWithErr(resp *userrpc.Response,
			err error) (*userrpc.Response, error)
		SuccessResponse(resp *userrpc.Response,
			msg string) (*userrpc.Response, error)
	}
	unBlackUserLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
	unBlackUserCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}
	unBlackUserValidate interface {
		ValidateBlackUser(in validate.InputBlackUser) (validate.OutputBlackUser, error)
	}
)

type UnBlackUserLogic struct {
	ctx                context.Context
	identityRepository unBlackUserIdentityRepository
	userResponse       unBlackUserResponse
	logger             unBlackUserLogger
	ctxdata            unBlackUserCtxData
	userValidate       unBlackUserValidate
}

func NewUnBlackUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnBlackUserLogic {
	return &UnBlackUserLogic{
		ctx:                ctx,
		identityRepository: svcCtx.IdentityRepository,
		userResponse:       svcCtx.UserResponse,
		logger:             svcCtx.Logger,
		ctxdata:            svcCtx.Ctxdata,
		userValidate:       svcCtx.UserValidate,
	}
}

func (l *UnBlackUserLogic) UnBlackUser(in *userrpc.UnBlackUserRequest) (*userrpc.Response, error) {
	resp := new(userrpc.Response)

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

	operatorRole, err := l.identityRepository.GetUserRole(l.ctx, param.OperatorID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据操作者ID查询用户身份权限等级失败",
			zap.String("operator-info", in.Operator),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.InterCallToolFail)
	}
	if operatorRole < 2 {
		l.logger.Error("操作者权限不足",
			zap.String("operator-info", in.Operator),
			zap.Int("operator-role", operatorRole),
		)
		return l.userResponse.FailResponse(resp, "操作者权限不足！",
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

	idExist, err := l.identityRepository.CheckIDExists(l.ctx, in.ID)
	if err != nil {
		l.logger.Error("数据库操作失败：查询预被解封用户账号ID失败",
			zap.String("unblack-user-id", in.ID),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.Failed)
	}
	if !idExist {
		l.logger.Error("要被解封的用户账号不存在",
			zap.String("unblack-user-id", in.ID),
		)
		return l.userResponse.FailResponse(resp, "要解封的用户ID不存在！",
			errcode.InternalError, errcode.Failed)
	}

	// 修改数据库字段
	err = l.identityRepository.UpdateUsersAccount(l.ctx, model.UserAccount{
		ID:     in.ID,
		Status: 0,
	})
	if err != nil {
		l.logger.Error("数据库操作失败：解封用户账号失败",
			zap.String("unblack-user-info", in.ID),
			zap.Error(err),
		)
		return l.userResponse.FailResponse(resp, "用户账号解封失败！",
			errcode.InternalError, errcode.InterCallToolFail)
	}

	l.logger.Info("解封用户成功",
		zap.String("operator-info", in.Operator),
		zap.Int("operator-role", operatorRole),
		zap.String("unblack-user-info", in.ID),
	)
	return l.userResponse.SuccessResponse(resp, "用户账号解封完成！")
}
