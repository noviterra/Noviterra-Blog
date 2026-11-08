package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/comment/commentrpc/internal/svc"
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"
)

type (
	getTreeCommentResponse interface {
		FailGetTreeResponse(resp *commentrpc.GetTreeCommentResponse,
			msg string, status int32, code int64) (*commentrpc.GetTreeCommentResponse, error)
		FailGetTreeResponseWithErr(resp *commentrpc.GetTreeCommentResponse,
			err error) (*commentrpc.GetTreeCommentResponse, error)
		SuccessGetTreeResponse(resp *commentrpc.GetTreeCommentResponse,
			comments []model.CommentNode,
			speakers map[string]model.CommentSpeaker,
			loginStatus bool) (*commentrpc.GetTreeCommentResponse, error)
		SuccessTreeResponseNoData(resp *commentrpc.GetTreeCommentResponse) (*commentrpc.GetTreeCommentResponse, error)
	}

	getTreeCommentCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
		GetUserLoginStatus(ctx context.Context) bool
	}

	getTreeCommonCache interface {
		GetTreeComment(ctx context.Context,
			rootID int64) ([]model.CommentNode, bool, error)
		SetTreeComment(ctx context.Context,
			rootID int64, comments []model.CommentNode) error
	}

	getTreeCommentArticleValidate interface {
		ValidateSnowID(id int64) (int64, error)
	}

	getTreeCommentRepository interface {
		OpenComment(ctx context.Context,
			commentID int64) ([]model.CommentNode, error)
	}

	getTreeCommentDataDeduplication interface {
		GetUserIDListsFromTreeComment(commentList []model.CommentNode) []string
	}

	getTreeCommentUserRepository interface {
		GetUserInfoInComment(ctx context.Context, userIDs []string) (map[string]model.CommentSpeaker, error)
	}
)

type GetTreeCommentLogic struct {
	ctx               context.Context
	commentResponse   getTreeCommentResponse
	ctxdata           getTreeCommentCtxData
	commentCache      getTreeCommonCache
	articleValidate   getTreeCommentArticleValidate
	commentRepository getTreeCommentRepository
	dataDeduplication getTreeCommentDataDeduplication
	userRepository    getTreeCommentUserRepository
}

func NewGetTreeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTreeCommentLogic {
	return &GetTreeCommentLogic{
		ctx:               ctx,
		commentResponse:   svcCtx.CommentResponse,
		ctxdata:           svcCtx.CtxData,
		commentCache:      svcCtx.CommentCache,
		articleValidate:   svcCtx.ArticleValidate,
		commentRepository: svcCtx.CommentRepository,
		dataDeduplication: svcCtx.DataDeduplication,
		userRepository:    svcCtx.UserRepository,
	}
}

// 得到某个评论的全部子评论
func (l *GetTreeCommentLogic) GetTreeComment(in *commentrpc.GetTreeCommentRequest) (*commentrpc.GetTreeCommentResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(commentrpc.GetTreeCommentResponse)

	// 参数校验
	commentID, err := l.articleValidate.ValidateSnowID(in.CommentID)
	if err != nil {
		return l.commentResponse.FailGetTreeResponseWithErr(resp, err)
	}

	var treeCommentList []model.CommentNode
	// 读取缓存
	treeCommentList, cacheOK, err := l.commentCache.GetTreeComment(l.ctx, commentID)
	if err != nil {
		return l.commentResponse.FailGetTreeResponse(resp, "系统繁忙，评论加载失败！", errcode.InternalError, errcode.Failed)
	}
	if !cacheOK { // 缓存未命中
		// 查库
		treeCommentList, err = l.commentRepository.OpenComment(l.ctx, commentID)
		if err != nil {
			return l.commentResponse.FailGetTreeResponse(resp, "系统繁忙，评论加载失败！", errcode.InternalError, errcode.Failed)
		}
	}
	if len(treeCommentList) == 0 {
		return l.commentResponse.SuccessTreeResponseNoData(resp)
	}
	// 缓存回写
	err = l.commentCache.SetTreeComment(l.ctx, commentID, treeCommentList)
	if err != nil {
		// 日志记录
	}

	// 提取出用户列表
	speakerIDs := l.dataDeduplication.GetUserIDListsFromTreeComment(treeCommentList)

	// 查库 得到用户信息表
	treeCommentSpeakerInfo, err := l.userRepository.GetUserInfoInComment(l.ctx, speakerIDs)
	if err != nil {
		return l.commentResponse.FailGetTreeResponse(resp, "系统繁忙，请稍后再试！", errcode.InternalError, errcode.Failed)
	}
	if len(treeCommentSpeakerInfo) == 0 {
		return l.commentResponse.SuccessTreeResponseNoData(resp)
	}

	// 获取用户登录状态
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	loginStatus := l.ctxdata.GetUserLoginStatus(l.ctx)

	return l.commentResponse.SuccessGetTreeResponse(
		resp,
		treeCommentList,
		treeCommentSpeakerInfo,
		loginStatus && id != "",
	)
}
