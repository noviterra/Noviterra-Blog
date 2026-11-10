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
	getUsersInfoUserRepository interface {
		GetAllUsersInfo(size, page int) ([]model.UserInfo, error)
	}
	getUsersInfoUserResponse interface {
		FailResponse(resp *userrpc.Response,
			msg string, status int32, code int64) (*userrpc.Response, error)
		FailResponseWithErr(resp *userrpc.Response,
			err error) (*userrpc.Response, error)
		SuccessResponse(resp *userrpc.Response,
			msg string) (*userrpc.Response, error)
		SuccessGetMoreUsersInfoResponse(resp *userrpc.GetMoreUsersInfoResponse,
			msg string, usersInfo []*userrpc.UserInfo) (*userrpc.GetMoreUsersInfoResponse, error)
		FailGetMoreUsersInfoResponse(resp *userrpc.GetMoreUsersInfoResponse,
			msg string, status int32, code int64) (*userrpc.GetMoreUsersInfoResponse, error)
		FailGetMoreUsersInfoResponseWithErr(resp *userrpc.GetMoreUsersInfoResponse,
			err error) (*userrpc.GetMoreUsersInfoResponse, error)
	}
	getUsersInfoLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type GetMoreUsersInfoLogic struct {
	ctx            context.Context
	userRepository getUsersInfoUserRepository
	userResponse   getUsersInfoUserResponse
	logger         getUsersInfoLogger
}

func NewGetMoreUsersInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMoreUsersInfoLogic {
	return &GetMoreUsersInfoLogic{
		ctx:            ctx,
		userRepository: svcCtx.UserRepository,
		userResponse:   svcCtx.UserResponse,
		logger:         svcCtx.Logger,
	}
}

func (l *GetMoreUsersInfoLogic) GetMoreUsersInfo(in *userrpc.GetMoreUsersInfoRequest) (*userrpc.GetMoreUsersInfoResponse, error) {
	resp := new(userrpc.GetMoreUsersInfoResponse)

	// 分页获取用户数据
	userList, err := l.userRepository.GetAllUsersInfo(int(in.Size), int(in.Page))
	if err != nil {
		l.logger.Error("数据库操作失败：分页获取用户信息失败",
			zap.Int64("size", in.Size),
			zap.Int64("page", in.Page),
			zap.Error(err),
		)
		return l.userResponse.FailGetMoreUsersInfoResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.InterCallToolFail)
	}
	if len(userList) == 0 {
		l.logger.Error("分页获取的用户信息为空",
			zap.Int64("size", in.Size),
			zap.Int64("page", in.Page),
		)
		return l.userResponse.FailGetMoreUsersInfoResponse(resp, "用户信息为空，换一页试试！",
			errcode.NotFound, errcode.GetUserInfoFail)
	}

	usersInfo := make([]*userrpc.UserInfo, 0, len(userList))
	for _, userInfo := range userList {
		item := &userrpc.UserInfo{
			ID:               userInfo.ID,
			UserName:         userInfo.Name,
			Email:            userInfo.Email,
			Avatar:           userInfo.Avatar,
			Gender:           int64(userInfo.Gender),
			Age:              int64(userInfo.Age),
			Role:             int64(userInfo.Role),
			SelfIntroduction: userInfo.SelfIntroduction,
			CreateAt:         userInfo.CreatedAt.Unix(),
			UpdatedAt:        userInfo.UpdatedAt.Unix(),
		}
		usersInfo = append(usersInfo, item)
	}

	l.logger.Info("分页获取用户信息成功",
		zap.Int64("size", in.Size),
		zap.Int64("page", in.Page),
		zap.Int("user-count", len(usersInfo)),
	)
	return l.userResponse.SuccessGetMoreUsersInfoResponse(resp, "获取用户信息成功！", usersInfo)
}
