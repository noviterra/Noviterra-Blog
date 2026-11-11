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
	getUsersInfoIdentityRepository interface {
		GetUserRole(ctx context.Context, id string) (int, error)
	}
	getUsersInfoUserRepository interface {
		GetAllUsersInfo(size, page int) ([]model.UserInfo, error)
	}
	getUsersInfoUserResponse interface {
		SuccessGetMoreUsersInfoResponse(resp *userrpc.GetMoreUsersInfoResponse,
			msg string, usersInfo []model.UserInfo) (*userrpc.GetMoreUsersInfoResponse, error)
		FailGetMoreUsersInfoResponse(resp *userrpc.GetMoreUsersInfoResponse,
			msg string, status int32, code int64) (*userrpc.GetMoreUsersInfoResponse, error)
		FailGetMoreUsersInfoResponseWithErr(resp *userrpc.GetMoreUsersInfoResponse,
			err error) (*userrpc.GetMoreUsersInfoResponse, error)
	}
	getUsersInfoLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
	getUsersInfoArticleValidate interface {
		ValidatePage(page, pageSize int32) (int, int, error)
	}
	getUsersInfoCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}
)

type GetMoreUsersInfoLogic struct {
	ctx                context.Context
	userRepository     getUsersInfoUserRepository
	userResponse       getUsersInfoUserResponse
	logger             getUsersInfoLogger
	articleValidate    getUsersInfoArticleValidate
	ctxdata            getUsersInfoCtxData
	identityRepository getUsersInfoIdentityRepository
}

func NewGetMoreUsersInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMoreUsersInfoLogic {
	return &GetMoreUsersInfoLogic{
		ctx:                ctx,
		userRepository:     svcCtx.UserRepository,
		userResponse:       svcCtx.UserResponse,
		logger:             svcCtx.Logger,
		articleValidate:    svcCtx.ArticleValidate,
		ctxdata:            svcCtx.Ctxdata,
		identityRepository: svcCtx.IdentityRepository,
	}
}

func (l *GetMoreUsersInfoLogic) GetMoreUsersInfo(in *userrpc.GetMoreUsersInfoRequest) (*userrpc.GetMoreUsersInfoResponse, error) {
	resp := new(userrpc.GetMoreUsersInfoResponse)

	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		l.logger.Error("上下文获取操作用户ID失败")
		return l.userResponse.FailGetMoreUsersInfoResponse(resp, "身份校验失败！",
			errcode.InvalidParam, errcode.GetContextUserIDFail)
	}

	// 校验用户身份权限 >=2 被允许操作
	userRole, err := l.identityRepository.GetUserRole(l.ctx, id)
	if err != nil {
		l.logger.Error("数据库操作失败：获取操作用户的身份权限失败",
			zap.String("user-info", id),
			zap.Error(err),
		)
		return l.userResponse.FailGetMoreUsersInfoResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.GetUserRoleFail)
	}
	if userRole < 2 {
		l.logger.Error("用户权限不足，无法进行批量获取用户信息操作",
			zap.String("user-info", id),
			zap.Int("user-role", userRole),
		)
		return l.userResponse.FailGetMoreUsersInfoResponse(resp, "权限不足！",
			errcode.Forbidden, errcode.UserRoleInsufficient)
	}

	// 参数校验
	page, pageSize, err := l.articleValidate.ValidatePage(in.Page, in.Size)
	if err != nil {
		l.logger.Error("请求参数校验未通过！",
			zap.Strings("user-info", []string{id}),
			zap.Int32("size", in.Size),
			zap.Int32("page", in.Page),
			zap.Error(err),
		)
		return l.userResponse.FailGetMoreUsersInfoResponse(resp, "服务失败："+err.Error(),
			errcode.InvalidParam, errcode.ValidateUserRequestParamFail)
	}

	// 分页获取用户数据
	userList, err := l.userRepository.GetAllUsersInfo(page, pageSize)
	if err != nil {
		l.logger.Error("数据库操作失败：分页获取用户信息失败",
			zap.Int("size", pageSize),
			zap.Int("page", page),
			zap.Error(err),
		)
		return l.userResponse.FailGetMoreUsersInfoResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.GetUserListFail)
	}
	if len(userList) == 0 {
		l.logger.Error("分页获取的用户信息为空",
			zap.Int("size", pageSize),
			zap.Int("page", page),
		)
		return l.userResponse.FailGetMoreUsersInfoResponse(resp, "用户信息为空，换一页试试！",
			errcode.NotFound, errcode.UserInfoEmpty)
	}

	l.logger.Info("分页获取用户信息成功",
		zap.Int("size", pageSize),
		zap.Int("page", page),
		zap.Int("user-count", len(userList)),
	)
	return l.userResponse.SuccessGetMoreUsersInfoResponse(resp, "获取用户信息成功！", userList)
}
