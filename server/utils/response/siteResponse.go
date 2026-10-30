package response

import (
	"Noviterra/server/app/site/siterpc/types/siterpc"
	"Noviterra/server/utils/errcode"
)

type SiteResponse struct {
}

func NewSiteResponse() *SiteResponse {
	return &SiteResponse{}
}

func (sr *SiteResponse) SuccessResponse(resp *siterpc.Response, msg string) (*siterpc.Response, error) {
	if resp == nil {
		resp = new(siterpc.Response)
	}
	resp.Status = errcode.RequestSuccess
	resp.Code = errcode.Success
	resp.Message = msg

	return resp, nil
}

func (sr *SiteResponse) FailResponse(resp *siterpc.Response, msg string, status int32, code int64) (*siterpc.Response, error) {
	if resp == nil {
		resp = new(siterpc.Response)
	}
	resp.Status = status
	resp.Code = code
	resp.Message = msg

	return resp, nil
}

func (sr *SiteResponse) FailResponseWithErr(resp *siterpc.Response, err error) (*siterpc.Response, error) {
	if resp == nil {
		resp = new(siterpc.Response)
	}

	return resp, err
}
