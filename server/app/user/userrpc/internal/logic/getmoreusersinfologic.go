package logic

import (
	"Noviterra/server/app/model"
	"context"

	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type getUsersInfoUserRepository interface {
	GetAllUsersInfo(size, page int) ([]model.UserInfo, error)
}

type GetMoreUsersInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger

	userRepository getUsersInfoUserRepository
}

func NewGetMoreUsersInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMoreUsersInfoLogic {
	return &GetMoreUsersInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),

		userRepository: svcCtx.UserRepository,
	}
}

func (l *GetMoreUsersInfoLogic) GetMoreUsersInfo(in *userrpc.GetMoreUsersInfoRequest) (*userrpc.GetMoreUsersInfoResponse, error) {
	// todo: add your logic here and delete this line
	// 分页获取用户数据
	userList, err := l.userRepository.GetAllUsersInfo(int(in.Size), int(in.Page))
	if err != nil {
		return &userrpc.GetMoreUsersInfoResponse{
			UsersInfo: nil,
			Resp: &userrpc.Response{
				Message: "",
				Code:    1,
			},
		}, err
	}
	if len(userList) == 0 {
		return &userrpc.GetMoreUsersInfoResponse{
			UsersInfo: nil,
			Resp: &userrpc.Response{
				Code:    1,
				Message: "用户信息为空，换一页试试！",
			},
		}, nil
	}

	rep := new(userrpc.GetMoreUsersInfoResponse)
	rep.UsersInfo = make([]*userrpc.UserInfo, 0, len(userList))

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
		rep.UsersInfo = append(rep.UsersInfo, item)
	}
	rep.Resp.Code = 0
	rep.Resp.Message = "获取用户信息成功！"

	return rep, nil
}
