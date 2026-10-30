package response

import (
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/utils/errcode"
)

type IdentityResponse struct {
}

func NewIdentityResponse() *IdentityResponse {
	return &IdentityResponse{}
}

func (ir *IdentityResponse) SuccessResponse(resp *identityrpc.Response, msg string) (*identityrpc.Response, error) {
	if resp == nil {
		resp = new(identityrpc.Response)
	}

	resp.Message = msg
	resp.Code = errcode.Success
	resp.Status = errcode.RequestSuccess

	return resp, nil
}

func (ir *IdentityResponse) FailResponse(resp *identityrpc.Response, msg string, status int32, code int64) (*identityrpc.Response, error) {
	if resp == nil {
		resp = new(identityrpc.Response)
	}

	resp.Message = msg
	resp.Status = status
	resp.Code = code

	return resp, nil
}

func (ir *IdentityResponse) FailResponseWithErr(resp *identityrpc.Response, err error) (*identityrpc.Response, error) {
	if resp == nil {
		resp = new(identityrpc.Response)
	}

	return resp, err
}

func (ir *IdentityResponse) SuccessGenerateTokenResponse(resp *identityrpc.GenerateTokenResponse,
	accessToken, refreshToken string) (*identityrpc.GenerateTokenResponse, error) {
	if resp == nil {
		resp = new(identityrpc.GenerateTokenResponse)
	}

	resp.Token = &identityrpc.TokenData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	resp.Resp = &identityrpc.Response{
		Message: "Token刷新成功！",
		Code:    errcode.Success,
		Status:  errcode.RequestSuccess,
	}

	return resp, nil
}

func (ir *IdentityResponse) FailGenerateTokenResponse(resp *identityrpc.GenerateTokenResponse,
	msg string, status int32, code int64) (*identityrpc.GenerateTokenResponse, error) {
	if resp == nil {
		resp = new(identityrpc.GenerateTokenResponse)
	}

	resp.Token = nil
	resp.Resp = &identityrpc.Response{
		Message: msg,
		Code:    code,
		Status:  status,
	}

	return resp, nil
}

func (ir *IdentityResponse) FailGenerateTokenResponseWithErr(resp *identityrpc.GenerateTokenResponse, err error) (*identityrpc.GenerateTokenResponse, error) {
	if resp == nil {
		resp = new(identityrpc.GenerateTokenResponse)
	}

	return resp, err
}

func (ir *IdentityResponse) SuccessLoginByEmailResponse(resp *identityrpc.LoginByEmailResponse, id string) (*identityrpc.LoginByEmailResponse, error) {
	if resp == nil {
		resp = new(identityrpc.LoginByEmailResponse)
	}

	resp.ID = id
	resp.Resp = &identityrpc.Response{
		Message: "邮箱登录成功！",
		Code:    errcode.Success,
		Status:  errcode.RequestSuccess,
	}

	return resp, nil
}

func (ir *IdentityResponse) FailLoginByEmailResponse(resp *identityrpc.LoginByEmailResponse,
	msg string, status int32, code int64) (*identityrpc.LoginByEmailResponse, error) {
	if resp == nil {
		resp = new(identityrpc.LoginByEmailResponse)
	}

	resp.ID = ""
	resp.Resp = &identityrpc.Response{
		Message: msg,
		Code:    code,
		Status:  status,
	}

	return resp, nil
}

func (ir *IdentityResponse) FailLoginByEmailResponseWithErr(resp *identityrpc.LoginByEmailResponse,
	err error) (*identityrpc.LoginByEmailResponse, error) {
	if resp == nil {
		resp = new(identityrpc.LoginByEmailResponse)
	}

	return resp, err
}

func (ir *IdentityResponse) SuccessRefreshAccessTokenResponse(resp *identityrpc.RefreshAccessTokenResponse,
	accessToken, refreshToken string) (*identityrpc.RefreshAccessTokenResponse, error) {
	if resp == nil {
		resp = new(identityrpc.RefreshAccessTokenResponse)
	}

	resp.Token = &identityrpc.TokenData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	resp.Resp = &identityrpc.Response{
		Message: "Token刷新成功！",
		Code:    errcode.Success,
		Status:  errcode.RequestSuccess,
	}

	return resp, nil
}

func (ir *IdentityResponse) FailRefreshAccessTokenResponse(resp *identityrpc.RefreshAccessTokenResponse,
	msg string, status int32, code int64) (*identityrpc.RefreshAccessTokenResponse, error) {
	if resp == nil {
		resp = new(identityrpc.RefreshAccessTokenResponse)
	}

	resp.Token = nil
	resp.Resp = &identityrpc.Response{
		Message: msg,
		Code:    code,
		Status:  status,
	}

	return resp, nil
}

func (ir *IdentityResponse) FailRefreshAccessTokenResponseWithErr(resp *identityrpc.RefreshAccessTokenResponse,
	err error) (*identityrpc.RefreshAccessTokenResponse, error) {
	if resp == nil {
		resp = new(identityrpc.RefreshAccessTokenResponse)
	}

	return resp, err
}
