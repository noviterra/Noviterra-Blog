package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"go.uber.org/zap"
)

type (
	searchArticleValidate interface {
		ValidatePage(page, pageSize int32) (int, int, error)
		ValidateQuery(query string) (string, error)
	}
	searchArticleLogger interface {
		Error(msg string, fields ...zap.Field)
		Info(msg string, fields ...zap.Field)
	}
	searchArticleResository interface {
		SearchArticle(ctx context.Context, like string, page, pageSize int) ([]model.ArticleList, int64, error)
	}
	searchArticleResponse interface {
		SuccessSearchArticleResponse(resp *articlerpc.SearchArticleResponse,
			msg string, articles []*articlerpc.ArticleList, totalCount int64) (*articlerpc.SearchArticleResponse, error)
		FailSearchArticleResponse(resp *articlerpc.SearchArticleResponse, msg string, status int32, code int64) (*articlerpc.SearchArticleResponse, error)
		FailSearchArticleResponseWithErr(resp *articlerpc.SearchArticleResponse, err error) (*articlerpc.SearchArticleResponse, error)
	}
)

type SearchArticleLogic struct {
	ctx               context.Context
	articleValidate   searchArticleValidate
	logger            searchArticleLogger
	articleRepository searchArticleResository
	articleResponse   searchArticleResponse
}

func NewSearchArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchArticleLogic {
	return &SearchArticleLogic{
		ctx:               ctx,
		articleValidate:   svcCtx.ArticleValidate,
		logger:            svcCtx.Logger,
		articleRepository: svcCtx.ArticleRepository,
		articleResponse:   svcCtx.ArticleResponse,
	}
}

// 搜索文章
func (l *SearchArticleLogic) SearchArticle(in *articlerpc.SearchArticleRequest) (*articlerpc.SearchArticleResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(articlerpc.SearchArticleResponse)
	// 参数校验
	page, pageSize, err := l.articleValidate.ValidatePage(in.Page, in.PageSize)
	if err != nil {
		l.logger.Error("请求参数校验失败：page/page-size",
			zap.Int32("page", in.Page),
			zap.Int32("page-size", in.PageSize),
			zap.Error(err),
		)
		return l.articleResponse.FailSearchArticleResponseWithErr(resp, err)
	}
	if page == 0 || pageSize == 0 {
		l.logger.Error("请求参数校验失败：page/page-size",
			zap.Int32("page", in.Page),
			zap.Int32("page-size", in.PageSize),
		)
		return l.articleResponse.FailSearchArticleResponse(resp, "参数异常！", errcode.InvalidParam, errcode.ValidateArticleRequestParamFail)
	}

	like, err := l.articleValidate.ValidateQuery(in.Query)
	if err != nil {
		l.logger.Error("校验查询参数失败",
			zap.String("query", in.Query),
			zap.Error(err),
		)
		return l.articleResponse.FailSearchArticleResponseWithErr(resp, err)
	}
	if like == "" {
		l.logger.Error("校验查询参数失败：模糊匹配模板为空",
			zap.String("query", in.Query),
		)
		return l.articleResponse.FailSearchArticleResponse(resp, "参数异常！", errcode.InvalidParam, errcode.ValidateArticleRequestParamFail)
	}
	// 查库
	searchArticleList, total, err := l.articleRepository.SearchArticle(l.ctx, like, page, pageSize)
	if err != nil {
		l.logger.Error("数据库操作失败：搜索文章失败",
			zap.String("query", in.Query),
			zap.Int("page", page),
			zap.Int("page-size", pageSize),
			zap.Error(err),
		)
		return l.articleResponse.FailSearchArticleResponse(resp, "数据库操作异常！", errcode.InternalError, errcode.SearchArticleFail)
	}
	if searchArticleList == nil || total == 0 {
		l.logger.Info("搜索文章结果为空",
			zap.String("query", in.Query),
			zap.Int64("article-count", total),
		)
		return l.articleResponse.SuccessSearchArticleResponse(resp, "未查询到相关文章！", nil, total)
	}

	// 数据映射
	resultList := make([]*articlerpc.ArticleList, 0, total)
	for _, list := range searchArticleList {
		ls := new(articlerpc.ArticleList)

		ls.ArticleID = list.ArticleID
		ls.Title = list.Title
		ls.Information = list.Information
		ls.Cover = list.Cover
		ls.Tags = list.Tags
		ls.CreatedAt = list.CreatedAt.Unix()

		resultList = append(resultList, ls)
	}

	l.logger.Info("搜索文章成功",
		zap.String("query", in.Query),
		zap.Int("page", page),
		zap.Int("page-size", pageSize),
		zap.Int64("article-count", total),
	)
	return l.articleResponse.SuccessSearchArticleResponse(resp, "查询列表加载完成！", resultList, total)
}
