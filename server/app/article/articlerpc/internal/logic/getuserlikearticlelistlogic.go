package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
)

type (
	getUserLIkeArticleListCommentValidate interface {
		ValidateUserLikeCollectArticleList(in validate.InputUserLikeCollectArticleList) (validate.OutputUserLikeCollectArticleList, error)
	}

	getUserLikeArticleListArticleResponse interface {
		SuccessGetArticleListResponse(resp *articlerpc.GetArticleListResponse,
			msg string, articles []model.ArticleList, totalCount int64) (*articlerpc.GetArticleListResponse, error)
		FailGetArticleListResponse(resp *articlerpc.GetArticleListResponse, msg string,
			status int32, code int64) (*articlerpc.GetArticleListResponse, error)
		FailGetArticleListResponseWithErr(resp *articlerpc.GetArticleListResponse, err error) (*articlerpc.GetArticleListResponse, error)
	}

	getUserLikeArticleListIdentityRepository interface {
		CheckIDExists(ctx context.Context, id string) (bool, error)
	}

	getUserLikeArticleListCommentCache interface {
		GetUserLikeCollectArticleList(ctx context.Context, userID, field string) ([]int64, error)
	}

	getUserLikeArticleListArticleRepository interface {
		GetArticleListByArticleID(ctx context.Context,
			articleIDs []int64, page, pageSize int) ([]model.ArticleList, int64, error)
	}
)

type GetUserLikeArticleListLogic struct {
	ctx                context.Context
	commentValidate    getUserLIkeArticleListCommentValidate
	articleResponse    getUserLikeArticleListArticleResponse
	identityRepository getUserLikeArticleListIdentityRepository
	commentCache       getUserLikeArticleListCommentCache
	articleRepository  getUserLikeArticleListArticleRepository
}

func NewGetUserLikeArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLikeArticleListLogic {
	return &GetUserLikeArticleListLogic{
		ctx:                ctx,
		commentValidate:    svcCtx.CommentValidate,
		articleResponse:    svcCtx.ArticleResponse,
		identityRepository: svcCtx.IdentityRepository,
		commentCache:       svcCtx.CommentCache,
		articleRepository:  svcCtx.ArticleRepository,
	}
}

// 得到用户点赞或收藏的全部文章列表
func (l *GetUserLikeArticleListLogic) GetUserLikeArticleList(in *articlerpc.GetUserLikeCollectArticleListRequest) (*articlerpc.GetArticleListResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(articlerpc.GetArticleListResponse)

	// 参数校验
	param, err := l.commentValidate.ValidateUserLikeCollectArticleList(validate.InputUserLikeCollectArticleList{
		UserID:   in.UserID,
		Page:     in.Page,
		PageSize: in.PageSize,
		Fileds:   in.Field,
	})
	if err != nil {
		return l.articleResponse.FailGetArticleListResponseWithErr(resp, err)
	}

	// 被查询用户ID校验 查询用户ID是否存在
	idExist, err := l.identityRepository.CheckIDExists(l.ctx, param.UserID)
	if err != nil {
		return l.articleResponse.FailGetArticleListResponse(resp, "系统繁忙，请稍后再试！", errcode.InternalError, errcode.Failed)
	}
	if !idExist {
		return l.articleResponse.FailGetArticleListResponse(resp, "查看的用户信息异常！", errcode.InternalError, errcode.Failed)
	}

	// 查询缓存 得到用户全部点赞或收藏的文章ID
	ids, err := l.commentCache.GetUserLikeCollectArticleList(l.ctx, param.UserID, param.Fileds)
	if err != nil {
		return l.articleResponse.FailGetArticleListResponse(resp, "系统繁忙，请稍后再试！", errcode.InternalError, errcode.Failed)
	}
	if len(ids) == 0 {
		return l.articleResponse.SuccessGetArticleListResponse(resp, "列表加载完成！", []model.ArticleList{}, 0)
	}

	// ids不为空 查库
	articleList, total, err := l.articleRepository.GetArticleListByArticleID(l.ctx, ids, param.Page, param.PageSize)
	if err != nil {
		return l.articleResponse.FailGetArticleListResponse(resp, "系统繁忙，请稍后再试！", errcode.InternalError, errcode.Failed)
	}

	return l.articleResponse.SuccessGetArticleListResponse(resp, "列表加载完成！", articleList, total)
}
