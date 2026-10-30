package response

import (
	"Noviterra/server/app/user/userrpc/types/userrpc"
	"Noviterra/server/utils/errcode"
)

type UserResponse struct {
}

func NewUserResponse() *UserResponse {
	return &UserResponse{}
}

func (ur *UserResponse) FailResponse(resp *userrpc.Response, msg string, code int64) (*userrpc.Response, error) {
	if resp == nil {
		resp = new(userrpc.Response)
	}

	resp.Code = code
	resp.Message = msg

	return resp, nil
}

func (ur *UserResponse) FailResponseWithErr(resp *userrpc.Response, err error) (*userrpc.Response, error) {
	if resp == nil {
		resp = new(userrpc.Response)
	}

	// rpc层返回error时 rep直接判为nil 没有必要对其赋值
	return resp, err
}

func (ur *UserResponse) SuccessResponse(resp *userrpc.Response, msg string) (*userrpc.Response, error) {
	if resp == nil {
		resp = new(userrpc.Response)
	}

	resp.Message = msg
	resp.Code = errcode.Success

	return resp, nil
}
