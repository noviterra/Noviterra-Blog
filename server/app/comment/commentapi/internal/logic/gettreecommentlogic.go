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
	getTreeCommentModel interface {
		GetTreeComment(ctx context.Context, in *commentrpc.GetTreeCommentRequest,
			opts ...grpc.CallOption) (*commentrpc.GetTreeCommentResponse, error)
	}

	getTreeCommentServiceTokenInterceptor interface {
		GetServiceTokenInConf() string
		GetSkipTokenInConf() string
	}

	getTreeCommentCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
		GetUserLoginStatus(ctx context.Context) bool
	}
)

type GetTreeCommentLogic struct {
	ctx                     context.Context
	commentModel            getTreeCommentModel
	serviceTokenInterceptor getTreeCommentServiceTokenInterceptor
	ctxdata                 getTreeCommentCtxData
}

// 展开某个评论
func NewGetTreeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTreeCommentLogic {
	return &GetTreeCommentLogic{
		ctx:                     ctx,
		commentModel:            svcCtx.CommentModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
		ctxdata:                 svcCtx.CtxData,
	}
}

func (l *GetTreeCommentLogic) GetTreeComment(req *types.GetTreeCommentRequest) (resp *types.GetTreeCommentResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.GetTreeCommentResponse)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return l.failGetTreeCommentResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return l.failGetTreeCommentResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	var ctx context.Context
	// 获取用户登录状态
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	loginStatus := l.ctxdata.GetUserLoginStatus(l.ctx)
	if id == "" || !loginStatus {
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

	rep, err := l.commentModel.GetTreeComment(ctx, &commentrpc.GetTreeCommentRequest{
		CommentID: req.CommentID,
	})
	if err != nil {
		return l.failGetTreeCommentResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return l.failGetTreeCommentResponse(resp, rep.Resp.Message, int(rep.Resp.Status), int(rep.Resp.Code))
	}

	return l.successGetTreeCommentResponse(resp, rep)
}

func (l *GetTreeCommentLogic) failGetTreeCommentResponse(resp *types.GetTreeCommentResponse,
	msg string, status, code int) (*types.GetTreeCommentResponse, error) {
	if resp == nil {
		resp = new(types.GetTreeCommentResponse)
	}
	resp.Comments = nil
	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: msg,
	}

	return resp, nil
}

func (l *GetTreeCommentLogic) failGetTreeCommentResponseWithErr(resp *types.GetTreeCommentResponse,
	err error, status, code int) (*types.GetTreeCommentResponse, error) {
	if resp == nil {
		resp = new(types.GetTreeCommentResponse)
	}
	resp.Comments = nil
	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "服务失败：" + err.Error(),
	}

	return resp, nil
}

func (l *GetTreeCommentLogic) successGetTreeCommentResponse(resp *types.GetTreeCommentResponse,
	rep *commentrpc.GetTreeCommentResponse) (*types.GetTreeCommentResponse, error) {
	if resp == nil {
		resp = new(types.GetTreeCommentResponse)
	}
	if len(rep.Comments) == 0 {
		resp.Comments = []types.TreeCommentInfo{}
	} else {
		for _, comment := range rep.Comments {
			resp.Comments = append(resp.Comments, types.TreeCommentInfo{
				CommentID:     comment.CommentID,
				SpeakerName:   comment.SpeakerName,
				SpeakerAvatar: comment.SpeakerAvatar,
				Content:       comment.Content,
				SpeakTime:     comment.SpeakTime,
			})
		}
	}
	resp.Resp = types.Response{
		Status:  int(rep.Resp.Status),
		Code:    int(rep.Resp.Code),
		Message: rep.Resp.Message,
	}

	return resp, nil
}
