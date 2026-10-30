package response

import (
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"
	"Noviterra/server/utils/errcode"
)

type CommentResponse struct {
}

func NewCommentResponse() *CommentResponse {
	return &CommentResponse{}
}

func (cr *CommentResponse) newResp(resp *commentrpc.Response) *commentrpc.Response {
	if resp == nil {
		resp = new(commentrpc.Response)
	}
	return resp
}

func (cr *CommentResponse) SuccessResponse(resp *commentrpc.Response, msg string) (*commentrpc.Response, error) {
	resp = cr.newResp(resp)
	resp.Status = errcode.RequestSuccess
	resp.Code = errcode.Success
	resp.Message = msg

	return resp, nil
}

func (cr *CommentResponse) FailResponse(resp *commentrpc.Response, msg string, status int32, code int64) (*commentrpc.Response, error) {
	resp = cr.newResp(resp)
	resp.Status = status
	resp.Code = code
	resp.Message = msg

	return resp, nil
}

func (cr *CommentResponse) FailResponseWithErr(resp *commentrpc.Response, err error) (*commentrpc.Response, error) {
	resp = cr.newResp(resp)

	return resp, err
}
