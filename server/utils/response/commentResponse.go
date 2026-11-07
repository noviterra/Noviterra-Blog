package response

import (
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
)

type CommentResponse struct {
}

func NewCommentResponse() *CommentResponse {
	return &CommentResponse{}
}

func (cr *CommentResponse) SuccessResponse(resp *commentrpc.Response, msg string) (*commentrpc.Response, error) {
	if resp == nil {
		resp = new(commentrpc.Response)
	}
	resp.Status = errcode.RequestSuccess
	resp.Code = errcode.Success
	resp.Message = msg

	return resp, nil
}

func (cr *CommentResponse) FailResponse(resp *commentrpc.Response, msg string, status int32, code int64) (*commentrpc.Response, error) {
	if resp == nil {
		resp = new(commentrpc.Response)
	}
	resp.Status = status
	resp.Code = code
	resp.Message = msg

	return resp, nil
}

func (cr *CommentResponse) FailResponseWithErr(resp *commentrpc.Response, err error) (*commentrpc.Response, error) {
	if resp == nil {
		resp = new(commentrpc.Response)
	}

	return resp, err
}

func (cr *CommentResponse) FailGetTopCommentResponseWithErr(resp *commentrpc.GetTopCommentResponse,
	err error) (*commentrpc.GetTopCommentResponse, error) {
	if resp == nil {
		resp = new(commentrpc.GetTopCommentResponse)
	}

	return resp, err
}

func (cr *CommentResponse) FailGetTopCommentResponse(resp *commentrpc.GetTopCommentResponse,
	msg string, status int32, code int64) (*commentrpc.GetTopCommentResponse, error) {
	if resp == nil {
		resp = new(commentrpc.GetTopCommentResponse)
	}

	resp.Comments = nil
	resp.Resp = &commentrpc.Response{
		Message: msg,
		Status:  status,
		Code:    code,
	}

	return resp, nil
}

func (cr *CommentResponse) SuccessGetTopCommentResponseDataNil(resp *commentrpc.GetTopCommentResponse) (*commentrpc.GetTopCommentResponse, error) {
	if resp == nil {
		resp = new(commentrpc.GetTopCommentResponse)
	}
	resp.Comments = []*commentrpc.CommentInfo{}
	resp.Resp = &commentrpc.Response{
		Message: "评论加载完成！",
		Status:  errcode.RequestSuccess,
		Code:    errcode.Success,
	}

	return resp, nil
}

func (cr *CommentResponse) SuccessGetTopCommentResponse(resp *commentrpc.GetTopCommentResponse,
	topComments []model.TopCommentNode,
	speakers map[string]model.CommentSpeaker,
	loginStatus bool) (*commentrpc.GetTopCommentResponse, error) {
	if resp == nil {
		resp = new(commentrpc.GetTopCommentResponse)
	}

	if len(topComments) == 0 || len(speakers) == 0 {
		resp.Comments = []*commentrpc.CommentInfo{}
	} else {
		// 检查登录状态
		limit := len(topComments)
		if !loginStatus {
			limit = limit / 2
			if limit < 1 {
				limit = 1
			}
		}

		for i := 0; i < limit; i++ {
			comment := topComments[i]
			speakerID := comment.SpeakerID
			speaker, ok := speakers[speakerID]
			if !ok {
				continue
			}

			resp.Comments = append(resp.Comments, &commentrpc.CommentInfo{
				CommentID:     comment.CommentID,
				SpeakerName:   speaker.Name,
				SpeakerAvatar: speaker.Avatar,
				Content:       comment.Content,
				Count:         comment.Count,
				SpeakTime:     comment.CreatedAt.Unix(),
			})
		}
	}

	resp.Resp = &commentrpc.Response{
		Message: "评论加载完成！",
		Status:  errcode.RequestSuccess,
		Code:    errcode.Success,
	}

	return resp, nil
}
