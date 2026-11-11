package response

import (
	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/types/userrpc"
	"Noviterra/server/utils/errcode"
)

type UserResponse struct {
}

func NewUserResponse() *UserResponse {
	return &UserResponse{}
}

func (ur *UserResponse) FailResponse(resp *userrpc.Response,
	msg string, status int32, code int64) (*userrpc.Response, error) {
	if resp == nil {
		resp = new(userrpc.Response)
	}

	resp.Code = code
	resp.Message = msg
	resp.Status = status

	return resp, nil
}

func (ur *UserResponse) FailResponseWithErr(resp *userrpc.Response,
	err error) (*userrpc.Response, error) {
	if resp == nil {
		resp = new(userrpc.Response)
	}

	// rpc层返回error时 rep直接判为nil 没有必要对其赋值
	return resp, err
}

func (ur *UserResponse) SuccessResponse(resp *userrpc.Response,
	msg string) (*userrpc.Response, error) {
	if resp == nil {
		resp = new(userrpc.Response)
	}

	resp.Message = msg
	resp.Code = errcode.Success
	resp.Status = errcode.RequestSuccess

	return resp, nil
}

// SuccessGetMoreUsersInfoResponse 封装GetMoreUsersInfoResponse
func (ur *UserResponse) SuccessGetMoreUsersInfoResponse(resp *userrpc.GetMoreUsersInfoResponse,
	msg string, usersInfo []model.UserInfo) (*userrpc.GetMoreUsersInfoResponse, error) {
	if resp == nil {
		resp = new(userrpc.GetMoreUsersInfoResponse)
	}

	for _, userInfo := range usersInfo {
		resp.UsersInfo = append(resp.UsersInfo, &userrpc.UserInfo{
			ID:               userInfo.ID,
			Role:             int64(userInfo.Role),
			UserName:         userInfo.Name,
			Email:            userInfo.Email,
			Avatar:           userInfo.Avatar,
			Age:              int64(userInfo.Age),
			Gender:           int64(userInfo.Gender),
			SelfIntroduction: userInfo.SelfIntroduction,
			CreateAt:         userInfo.CreatedAt.Unix(),
			UpdatedAt:        userInfo.UpdatedAt.Unix(),
		})
	}
	resp.Resp = &userrpc.Response{
		Message: msg,
		Code:    errcode.Success,
		Status:  errcode.RequestSuccess,
	}

	return resp, nil
}

// FailGetMoreUsersInfoResponse 不携带err
func (ur *UserResponse) FailGetMoreUsersInfoResponse(resp *userrpc.GetMoreUsersInfoResponse,
	msg string, status int32, code int64) (*userrpc.GetMoreUsersInfoResponse, error) {
	if resp == nil {
		resp = new(userrpc.GetMoreUsersInfoResponse)
	}

	resp.UsersInfo = nil
	resp.Resp = &userrpc.Response{
		Message: msg,
		Code:    code,
		Status:  status,
	}

	return resp, nil
}

// FailGetMoreUsersInfoResponseWithErr 携带err
func (ur *UserResponse) FailGetMoreUsersInfoResponseWithErr(resp *userrpc.GetMoreUsersInfoResponse,
	err error) (*userrpc.GetMoreUsersInfoResponse, error) {
	if resp == nil {
		resp = new(userrpc.GetMoreUsersInfoResponse)
	}

	return resp, err
}

// SuccessGetSelfInfoResponse 封装UserInfoResponse
func (ur *UserResponse) SuccessGetSelfInfoResponse(resp *userrpc.UserInfoResponse,
	msg string, data *model.UserInfo) (*userrpc.UserInfoResponse, error) {
	if resp == nil {
		resp = new(userrpc.UserInfoResponse)
	}

	resp.Data = &userrpc.UserInfo{
		ID:               data.ID,
		Role:             int64(data.Role),
		UserName:         data.Name,
		Email:            data.Email,
		Avatar:           data.Avatar,
		Age:              int64(data.Age),
		Gender:           int64(data.Gender),
		SelfIntroduction: data.SelfIntroduction,
		CreateAt:         data.CreatedAt.Unix(),
		UpdatedAt:        data.UpdatedAt.Unix(),
	}
	resp.Resp = &userrpc.Response{
		Message: msg,
		Code:    errcode.Success,
		Status:  errcode.RequestSuccess,
	}

	return resp, nil
}

// FailGetSelfInfoResponse 不携带err
func (ur *UserResponse) FailGetSelfInfoResponse(resp *userrpc.UserInfoResponse,
	msg string, status int32, code int64) (*userrpc.UserInfoResponse, error) {
	if resp == nil {
		resp = new(userrpc.UserInfoResponse)
	}

	resp.Data = nil
	resp.Resp = &userrpc.Response{
		Message: msg,
		Code:    code,
		Status:  status,
	}

	return resp, nil
}

// FailGetSelfInfoResponseWithErr 携带err
func (ur *UserResponse) FailGetSelfInfoResponseWithErr(resp *userrpc.UserInfoResponse,
	err error) (*userrpc.UserInfoResponse, error) {
	if resp == nil {
		resp = new(userrpc.UserInfoResponse)
	}

	return resp, err
}
