package logic

import (
	"Noviterra/server/app/comment/commentrpc/internal/svc"
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"
	"context"
)

type (
	createCommentCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	createCommentValidate interface {
		ValidateComment(in validate.InputComment) (validate.OutputComment, error)
	}

	createCommentResponse interface {
		SuccessResponse(resp *commentrpc.Response, msg string) (*commentrpc.Response, error)
		FailResponse(resp *commentrpc.Response, msg string, status int32, code int64) (*commentrpc.Response, error)
		FailResponseWithErr(resp *commentrpc.Response, err error) (*commentrpc.Response, error)
	}

	createCommentRepository interface {
		InsertComment(ctx context.Context, comment *model.ArticleComment) error
	}

	createCommentSnowFlake interface {
		GenerateSnowFlakeID() int64
	}
)

type CreateCommentLogic struct {
	ctx               context.Context
	commentResponse   createCommentResponse
	commentValidate   createCommentValidate
	ctxdata           createCommentCtxData
	commentRepository createCommentRepository
	snowflake         createCommentSnowFlake
}

func NewCreateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCommentLogic {
	return &CreateCommentLogic{
		ctx:               ctx,
		ctxdata:           svcCtx.CtxData,
		commentResponse:   svcCtx.CommentResponse,
		commentValidate:   svcCtx.CommentValidate,
		commentRepository: svcCtx.CommentRepository,
		snowflake:         svcCtx.Snowfake,
	}
}

// 发表评论
func (l *CreateCommentLogic) CreateComment(in *commentrpc.CommentRequest) (*commentrpc.Response, error) {
	// todo: add your logic here and delete this line
	resp := new(commentrpc.Response)
	// 身份校验
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		return l.commentResponse.FailResponse(resp, "信息校验失败！", errcode.Unauthorized, errcode.Failed)
	}

	// 参数校验
	commentReq, err := l.commentValidate.ValidateComment(validate.InputComment{
		Address:   in.Address,
		ParentID:  in.ParentID,
		SpeakerID: in.SpeakerID,
		Content:   in.Content,
	})
	if err != nil {
		return l.commentResponse.FailResponseWithErr(resp, err)
	}

	// 生成雪花ID作为评论的唯一ID
	commentID := l.snowflake.GenerateSnowFlakeID()

	// 入库
	err = l.commentRepository.InsertComment(l.ctx, &model.ArticleComment{
		CommentID: commentID,
		ParentID:  commentReq.ParentID,
		SpeakerID: commentReq.SpeakerID,
		Content:   commentReq.Content,
		Address:   commentReq.Address,
	})
	if err != nil {
		return l.commentResponse.FailResponse(resp, "系统繁忙，评论失败！", errcode.InternalError, errcode.Failed)
	}

	return l.commentResponse.SuccessResponse(resp, "评论成功！")
}
