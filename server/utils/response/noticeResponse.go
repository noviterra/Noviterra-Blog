package response

import (
	"Noviterra/server/app/model"
	"Noviterra/server/app/notice/noticerpc/types/noticerpc"
	"Noviterra/server/utils/errcode"
)

type NoticeResponse struct {
}

func NewNoticeResponse() *NoticeResponse {
	return &NoticeResponse{}
}

func (nr *NoticeResponse) SuccessResponse(resp *noticerpc.Response, msg string) (*noticerpc.Response, error) {
	if resp == nil {
		resp = new(noticerpc.Response)
	}
	resp.Message = msg
	resp.Status = errcode.RequestSuccess
	resp.Code = errcode.Success

	return resp, nil
}

func (nr *NoticeResponse) FailResponse(resp *noticerpc.Response, msg string, status int32, code int64) (*noticerpc.Response, error) {
	if resp == nil {
		resp = new(noticerpc.Response)
	}
	resp.Message = msg
	resp.Status = status
	resp.Code = code

	return resp, nil
}

func (nr *NoticeResponse) FailResponseWithErr(resp *noticerpc.Response, err error) (*noticerpc.Response, error) {
	if resp == nil {
		resp = new(noticerpc.Response)
	}

	return resp, err
}

func (nr *NoticeResponse) SuccessGetNoticeResponse(resp *noticerpc.GetNoticeResponse,
	dtos []model.NoticeLookDTO) (*noticerpc.GetNoticeResponse, error) {
	if resp == nil {
		resp = new(noticerpc.GetNoticeResponse)
	}

	for _, notice := range dtos {
		var n noticerpc.Notice

		n.Publisher = notice.Publisher
		n.Title = notice.Title
		n.Message = notice.Message
		n.CreatedAt = notice.CreatedAt.Unix()

		resp.Notices = append(resp.Notices, &n)
	}
	resp.Resp = &noticerpc.Response{
		Message: "公告加载完成！",
		Code:    errcode.Success,
		Status:  errcode.RequestSuccess,
	}

	return resp, nil
}

func (nr *NoticeResponse) FailGetNoticeResponse(resp *noticerpc.GetNoticeResponse,
	msg string, status int32, code int64) (*noticerpc.GetNoticeResponse, error) {
	if resp == nil {
		resp = new(noticerpc.GetNoticeResponse)
	}
	resp.Notices = nil
	resp.Resp = &noticerpc.Response{
		Message: msg,
		Code:    code,
		Status:  status,
	}

	return resp, nil
}

func (nr *NoticeResponse) FailGetNoticeResponseWithErr(resp *noticerpc.GetNoticeResponse, err error) (*noticerpc.GetNoticeResponse, error) {
	if resp == nil {
		resp = new(noticerpc.GetNoticeResponse)
	}

	return resp, err
}
