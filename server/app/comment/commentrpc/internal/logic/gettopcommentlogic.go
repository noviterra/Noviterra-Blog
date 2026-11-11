package logic

import (
	"context"

	"Noviterra/server/app/comment/commentrpc/internal/svc"
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"

	"go.uber.org/zap"
)

type (
	getTopCommentCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
		GetUserLoginStatus(ctx context.Context) bool
	}

	getTopCommentValidate interface {
		ValidateCommentAddress(address int64) (int64, error)
	}

	getTopCommentResponse interface {
		FailGetTopCommentResponseWithErr(resp *commentrpc.GetTopCommentResponse,
			err error) (*commentrpc.GetTopCommentResponse, error)
		FailGetTopCommentResponse(resp *commentrpc.GetTopCommentResponse,
			msg string, status int32, code int64) (*commentrpc.GetTopCommentResponse, error)
		SuccessGetTopCommentResponse(resp *commentrpc.GetTopCommentResponse,
			topComments []model.TopCommentNode,
			speakers map[string]model.CommentSpeaker,
			status bool) (*commentrpc.GetTopCommentResponse, error)
		SuccessGetTopCommentResponseDataNil(resp *commentrpc.GetTopCommentResponse) (*commentrpc.GetTopCommentResponse, error)
	}

	getTopCommentCache interface {
		GetTopComment(ctx context.Context,
			address int64) ([]model.TopCommentNode, bool, error)
		SetTopComment(ctx context.Context,
			address int64, comments []model.TopCommentNode) error
	}

	getTopCommentRepository interface {
		FirstLoadTopComment(ctx context.Context,
			address int64) ([]model.TopCommentNode, error)
	}

	getTopCommentDataDeduplication interface {
		GetUserIDListsFromTopComment(commentList []model.TopCommentNode) []string
	}

	getTopCommentUserRepository interface {
		GetUserInfoInComment(ctx context.Context, userIDs []string) (map[string]model.CommentSpeaker, error)
	}

	getTopCommentLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type GetTopCommentLogic struct {
	ctx               context.Context
	commentValidate   getTopCommentValidate
	commentResponse   getTopCommentResponse
	ctxdata           getTopCommentCtxData
	commentCache      getTopCommentCache
	commentRepository getTopCommentRepository
	dataDeduplication getTopCommentDataDeduplication
	userRepository    getTopCommentUserRepository
	logger            getTopCommentLogger
}

func NewGetTopCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTopCommentLogic {
	return &GetTopCommentLogic{
		ctx:               ctx,
		commentValidate:   svcCtx.CommentValidate,
		commentResponse:   svcCtx.CommentResponse,
		ctxdata:           svcCtx.CtxData,
		commentCache:      svcCtx.CommentCache,
		commentRepository: svcCtx.CommentRepository,
		dataDeduplication: svcCtx.DataDeduplication,
		userRepository:    svcCtx.UserRepository,
		logger:            svcCtx.Logger,
	}
}

// 得到某个位置的全部顶层评论
func (l *GetTopCommentLogic) GetTopComment(in *commentrpc.GetTopCommentRequest) (*commentrpc.GetTopCommentResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(commentrpc.GetTopCommentResponse)
	// 参数校验
	address, err := l.commentValidate.ValidateCommentAddress(in.Address)
	if err != nil {
		l.logger.Error("校验评论地址参数失败",
			zap.Int64("address", in.Address),
			zap.Error(err),
		)
		return l.commentResponse.FailGetTopCommentResponseWithErr(resp, err)
	}

	var topCommentList []model.TopCommentNode
	// 读缓存
	topCommentList, cacheOK, err := l.commentCache.GetTopComment(l.ctx, address)
	if err != nil {
		l.logger.Error("缓存操作失败：查询顶层评论缓存失败",
			zap.Int64("address", address),
			zap.Error(err),
		)
		return l.commentResponse.FailGetTopCommentResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.GetTopCommentCacheFail)
	}
	if !cacheOK { // 缓存未命中
		l.logger.Info("顶层评论缓存未命中，查询数据库",
			zap.Int64("address", address),
		)
		// 查库
		topCommentList, err = l.commentRepository.FirstLoadTopComment(l.ctx, address)
		if err != nil {
			l.logger.Error("数据库操作失败：查询顶层评论失败",
				zap.Int64("address", address),
				zap.Error(err),
			)
			return l.commentResponse.FailGetTopCommentResponse(resp, "系统繁忙，请稍后再试！",
				errcode.InternalError, errcode.GetTopCommentFail)
		}
	}
	if len(topCommentList) == 0 {
		return l.commentResponse.SuccessGetTopCommentResponseDataNil(resp)
	}
	// 缓存回写
	err = l.commentCache.SetTopComment(l.ctx, address, topCommentList)
	if err != nil {
		l.logger.Error("缓存操作失败：写入顶层评论缓存失败",
			zap.Int64("address", address),
			zap.Int("comment-count", len(topCommentList)),
			zap.Error(err),
		)
	}

	// 去重得到用户ID列表
	speakerIDs := l.dataDeduplication.GetUserIDListsFromTopComment(topCommentList)

	// 查询用户表 得到speaker info
	topCommentSpeakerInfo, err := l.userRepository.GetUserInfoInComment(l.ctx, speakerIDs)
	if err != nil {
		l.logger.Error("数据库操作失败：查询顶层评论用户信息失败",
			zap.Int64("address", address),
			zap.Int("speaker-count", len(speakerIDs)),
			zap.Error(err),
		)
		return l.commentResponse.FailGetTopCommentResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.GetTopCommentUserInfoFail)
	}
	if len(topCommentSpeakerInfo) == 0 {
		return l.commentResponse.SuccessGetTopCommentResponseDataNil(resp)
	}

	// 登录状态校验
	loginStatus := l.ctxdata.GetUserLoginStatus(l.ctx)
	id := l.ctxdata.GetUserIDFromContext(l.ctx)

	l.logger.Info("查询顶层评论成功",
		zap.Int64("address", address),
		zap.Int("comment-count", len(topCommentList)),
		zap.Int("speaker-count", len(topCommentSpeakerInfo)),
	)
	return l.commentResponse.SuccessGetTopCommentResponse(
		resp,
		topCommentList,
		topCommentSpeakerInfo,
		loginStatus && id != "",
	)
}
