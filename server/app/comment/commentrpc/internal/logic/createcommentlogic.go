package logic

import (
	"context"

	"Noviterra/server/app/comment/commentrpc/internal/svc"
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"

	"go.uber.org/zap"
)

type (
	createCommentCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	createCommentValidate interface {
		ValidateComment(in validate.InputComment) (validate.OutputComment, error)
	}

	createCommentResponse interface {
		SuccessCommentResponse(resp *commentrpc.CommentResponse,
			commentID int64) (*commentrpc.CommentResponse, error)
		FailCommentResponse(resp *commentrpc.CommentResponse,
			msg string, status int32, code int64) (*commentrpc.CommentResponse, error)
		FailCommentResponseWithErr(resp *commentrpc.CommentResponse,
			err error) (*commentrpc.CommentResponse, error)
	}

	createCommentRepository interface {
		InsertComment(ctx context.Context, comment *model.ArticleComment) (int64, error)
	}

	createCommentSnowFlake interface {
		GenerateSnowFlakeID() int64
	}

	createCommentCache interface {
		DelTreeComment(ctx context.Context,
			rootID int64) error
		DelTopComment(ctx context.Context,
			address int64) error
	}

	createCommentLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type CreateCommentLogic struct {
	ctx               context.Context
	commentResponse   createCommentResponse
	commentValidate   createCommentValidate
	ctxdata           createCommentCtxData
	commentRepository createCommentRepository
	snowflake         createCommentSnowFlake
	commentCache      createCommentCache
	logger            createCommentLogger
}

func NewCreateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCommentLogic {
	return &CreateCommentLogic{
		ctx:               ctx,
		ctxdata:           svcCtx.CtxData,
		commentResponse:   svcCtx.CommentResponse,
		commentValidate:   svcCtx.CommentValidate,
		commentRepository: svcCtx.CommentRepository,
		snowflake:         svcCtx.Snowfake,
		commentCache:      svcCtx.CommentCache,
		logger:            svcCtx.Logger,
	}
}

// 发表评论
func (l *CreateCommentLogic) CreateComment(in *commentrpc.CommentRequest) (*commentrpc.CommentResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(commentrpc.CommentResponse)
	// 身份校验
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		l.logger.Error("上下文获取操作用户ID失败")
		return l.commentResponse.FailCommentResponse(resp, "信息校验失败！",
			errcode.Unauthorized, errcode.CommentUserInfoInvalid)
	}

	// 参数校验
	commentReq, err := l.commentValidate.ValidateComment(validate.InputComment{
		Address:   in.Address,
		ParentID:  in.ParentID,
		SpeakerID: in.SpeakerID,
		Content:   in.Content,
	})
	if err != nil {
		l.logger.Error("校验发表评论参数失败",
			zap.String("user-info", id),
			zap.Int64("address", in.Address),
			zap.Int64("parent-id", in.ParentID),
			zap.String("speaker-id", in.SpeakerID),
			zap.Error(err),
		)
		return l.commentResponse.FailCommentResponseWithErr(resp, err)
	}

	// 生成雪花ID作为评论的唯一ID
	commentID := l.snowflake.GenerateSnowFlakeID()

	// 入库
	rootID, err := l.commentRepository.InsertComment(l.ctx, &model.ArticleComment{
		CommentID: commentID,
		ParentID:  commentReq.ParentID,
		SpeakerID: commentReq.SpeakerID,
		Content:   commentReq.Content,
		Address:   commentReq.Address,
	})
	if err != nil {
		l.logger.Error("数据库操作失败：发表评论失败",
			zap.String("user-info", id),
			zap.Int64("comment-id", commentID),
			zap.Int64("address", commentReq.Address),
			zap.String("speaker-id", commentReq.SpeakerID),
			zap.Error(err),
		)
		return l.commentResponse.FailCommentResponse(resp, "系统繁忙，评论失败！",
			errcode.InternalError, errcode.CreateCommentFail)
	}

	// 评论发布后 删除缓存 日志记录
	if commentReq.ParentID == 0 { // 发布的为顶层评论 则删除顶层评论缓存
		err = l.commentCache.DelTopComment(l.ctx, commentReq.Address)
	} else { // 子级评论 将所在的评论树缓存删除
		err = l.commentCache.DelTreeComment(l.ctx, rootID)
	}
	if err != nil {
		l.logger.Error("缓存操作失败：删除评论缓存失败",
			zap.String("user-info", id),
			zap.Int64("comment-id", commentID),
			zap.Int64("root-id", rootID),
			zap.Int64("address", commentReq.Address),
			zap.Error(err),
		)
	}

	l.logger.Info("用户发表评论成功",
		zap.String("user-info", id),
		zap.Int64("comment-id", commentID),
		zap.Int64("address", commentReq.Address),
	)
	return l.commentResponse.SuccessCommentResponse(resp, commentID)
}
