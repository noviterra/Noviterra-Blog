package logic

import (
	"context"

	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMoreUsersInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMoreUsersInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMoreUsersInfoLogic {
	return &GetMoreUsersInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMoreUsersInfoLogic) GetMoreUsersInfo(in *userrpc.GetMoreUsersInfoRequest) (*userrpc.GetMoreUsersInfoResponse, error) {
	// todo: add your logic here and delete this line
	// 分页获取用户数据
	userList, err := l.svcCtx.DB.GetAllUsersInfo(int(in.Size), int(in.Page))
	if err != nil {
		return &userrpc.GetMoreUsersInfoResponse{
			Code:    1,
			Message: "数据库操作异常，请联系管理员！",
		}, err
	}
	if len(userList) == 0 {
		return &userrpc.GetMoreUsersInfoResponse{
			Code:    1,
			Message: "用户信息为空，换一页试试！",
		}, nil
	}

	rep := new(userrpc.GetMoreUsersInfoResponse)
	rep.UsersInfo = make([]*userrpc.UserInfoResponse, 0, len(userList))

	for _, userInfo := range userList {
		item := &userrpc.UserInfoResponse{
			Data: &userrpc.UserInfoResponse_Data{
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
			},
		}
		rep.UsersInfo = append(rep.UsersInfo, item)
	}
	rep.Code = 0
	rep.Message = "获取用户信息成功！"

	return rep, nil
}
