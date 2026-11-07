package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/comment/commentrpc/internal/svc"
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"
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
	}
}

// 得到某个位置的全部顶层评论
func (l *GetTopCommentLogic) GetTopComment(in *commentrpc.GetTopCommentRequest) (*commentrpc.GetTopCommentResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(commentrpc.GetTopCommentResponse)
	// 参数校验
	address, err := l.commentValidate.ValidateCommentAddress(in.Address)
	if err != nil {
		return l.commentResponse.FailGetTopCommentResponseWithErr(resp, err)
	}

	var topCommentList []model.TopCommentNode
	// 读缓存
	topCommentList, cacheOK, err := l.commentCache.GetTopComment(l.ctx, address)
	if err != nil {
		return l.commentResponse.FailGetTopCommentResponse(resp, "系统繁忙，请稍后再试！", errcode.InternalError, errcode.Failed)
	}
	if !cacheOK { // 缓存未命中
		// 查库
		topCommentList, err = l.commentRepository.FirstLoadTopComment(l.ctx, address)
		if err != nil {
			return l.commentResponse.FailGetTopCommentResponse(resp, "系统繁忙，请稍后再试！", errcode.InternalError, errcode.Failed)
		}
	}
	if len(topCommentList) == 0 {
		return l.commentResponse.SuccessGetTopCommentResponseDataNil(resp)
	}
	// 缓存回写
	err = l.commentCache.SetTopComment(l.ctx, address, topCommentList)
	if err != nil {
		// 日志记录
	}

	// 去重得到用户ID列表
	speakerIDs := l.dataDeduplication.GetUserIDListsFromTopComment(topCommentList)

	// 查询用户表 得到speaker info
	topCommentSpeakerInfo, err := l.userRepository.GetUserInfoInComment(l.ctx, speakerIDs)
	if err != nil {
		return l.commentResponse.FailGetTopCommentResponse(resp, "系统繁忙，请稍后再试！", errcode.InternalError, errcode.Failed)
	}
	if len(topCommentSpeakerInfo) == 0 {
		return l.commentResponse.SuccessGetTopCommentResponseDataNil(resp)
	}

	// 登录状态校验
	loginStatus := l.ctxdata.GetUserLoginStatus(l.ctx)
	id := l.ctxdata.GetUserIDFromContext(l.ctx)

	return l.commentResponse.SuccessGetTopCommentResponse(
		resp,
		topCommentList,
		topCommentSpeakerInfo,
		loginStatus && id != "",
	)
}
