// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/comment/commentapi/internal/svc"
	"Noviterra/server/app/comment/commentapi/internal/types"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	getTopCommentModel interface {
		GetTopComment(ctx context.Context, in *commentrpc.GetTopCommentRequest,
			opts ...grpc.CallOption) (*commentrpc.GetTopCommentResponse, error)
	}

	getTopCommentServiceTokenInterceptor interface {
		GetServiceTokenInConf() string
		GetSkipTokenInConf() string
	}

	getTopCommentCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
		GetUserLoginStatus(ctx context.Context) bool
	}
)

type GetTopCommentLogic struct {
	ctx                     context.Context
	commentModel            getTopCommentModel
	ctxdata                 getTopCommentCtxData
	serviceTokenInterceptor getTopCommentServiceTokenInterceptor
}

// 得到一篇文章的顶层评论
func NewGetTopCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTopCommentLogic {
	return &GetTopCommentLogic{
		ctx:                     ctx,
		commentModel:            svcCtx.CommentModel,
		ctxdata:                 svcCtx.CtxData,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *GetTopCommentLogic) GetTopComment(req *types.GetTopCommentRequest) (resp *types.GetTopCommentResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.GetTopCommentResponse)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return l.failGetTopCommentResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return l.failGetTopCommentResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	var ctx context.Context
	// 校验是否登录
	loginStatus := l.ctxdata.GetUserLoginStatus(l.ctx)
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" || loginStatus == false {
		ctx = metadata.AppendToOutgoingContext(l.ctx,
			"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
			"skip-token", l.serviceTokenInterceptor.GetSkipTokenInConf(),
		)
	} else {
		ctx = metadata.AppendToOutgoingContext(l.ctx,
			"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
			"user-id", id,
		)
	}

	rep, err := l.commentModel.GetTopComment(ctx, &commentrpc.GetTopCommentRequest{
		Address: req.Address,
	})
	if err != nil {
		return l.failGetTopCommentResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return l.failGetTopCommentResponse(resp, rep.Resp.Message, int(rep.Resp.Status), int(rep.Resp.Code))
	}

	return l.successGetTopCommentResponse(resp, rep)
}

func (l *GetTopCommentLogic) failGetTopCommentResponse(resp *types.GetTopCommentResponse,
	msg string, status, code int) (*types.GetTopCommentResponse, error) {
	if resp == nil {
		resp = new(types.GetTopCommentResponse)
	}
	resp.Resp.Code = code
	resp.Resp.Status = status
	resp.Resp.Message = msg
	resp.Comments = nil

	return resp, nil
}

func (l *GetTopCommentLogic) failGetTopCommentResponseWithErr(resp *types.GetTopCommentResponse,
	err error, status, code int) (*types.GetTopCommentResponse, error) {
	if resp == nil {
		resp = new(types.GetTopCommentResponse)
	}
	resp.Resp.Code = code
	resp.Resp.Status = status
	resp.Resp.Message = "服务失败：" + err.Error()
	resp.Comments = nil

	return resp, nil
}

func (l *GetTopCommentLogic) successGetTopCommentResponse(resp *types.GetTopCommentResponse,
	rep *commentrpc.GetTopCommentResponse) (*types.GetTopCommentResponse, error) {
	if resp == nil {
		resp = new(types.GetTopCommentResponse)
	}

	if len(rep.Comments) == 0 {
		resp.Comments = []types.CommentInfo{}
	} else {
		for _, commentInfo := range rep.Comments {
			var c types.CommentInfo
			c.CommentID = commentInfo.CommentID
			c.SpeakerName = commentInfo.SpeakerName
			c.SpeakerAvatar = commentInfo.SpeakerAvatar
			c.Content = commentInfo.Content
			c.SpeakTime = commentInfo.SpeakTime
			c.Count = commentInfo.Count

			resp.Comments = append(resp.Comments, c)
		}
	}

	resp.Resp.Code = int(rep.Resp.Code)
	resp.Resp.Status = int(rep.Resp.Status)
	resp.Resp.Message = rep.Resp.Message

	return resp, nil
}
