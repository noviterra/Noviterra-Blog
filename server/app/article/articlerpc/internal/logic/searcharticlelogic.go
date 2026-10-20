package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"go.uber.org/zap"
)

type searchArticleCommonUtils interface {
	ValidatePage(page, pageSize int32) (int, int, error)
	ValidateQuery(query string) (string, error)
}

type searchArticleLogger interface {
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
}

type searchArticleResository interface {
	SearchArticle(ctx context.Context, like string, page, pageSize int) ([]model.ArticleList, int64, error)
}

type searchArticleResponse interface {
	RPCSuccessSearchArticleResponse(msg string, articles []*articlerpc.ArticleList, totalCount int64) (*articlerpc.SearchArticleResponse, error)
	RPCFailSearchArticleResponse(msg string, code int64) (*articlerpc.SearchArticleResponse, error)
	RPCFailSearchArticleResponseWithErr(err error) (*articlerpc.SearchArticleResponse, error)
}
type SearchArticleLogic struct {
	ctx               context.Context
	commonUtils       searchArticleCommonUtils
	logger            searchArticleLogger
	articleRepository searchArticleResository
	articleResponse   searchArticleResponse
}

func NewSearchArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchArticleLogic {
	return &SearchArticleLogic{
		ctx:               ctx,
		commonUtils:       svcCtx.CommonUtils,
		logger:            svcCtx.Logger,
		articleRepository: svcCtx.ArticleRepository,
		articleResponse:   svcCtx.ArticleResponse,
	}
}

// 搜索文章
func (l *SearchArticleLogic) SearchArticle(in *articlerpc.SearchArticleRequest) (*articlerpc.SearchArticleResponse, error) {
	// todo: add your logic here and delete this line
	// 参数校验
	page, pageSize, err := l.commonUtils.ValidatePage(in.Page, in.PageSize)
	if err != nil {
		l.logger.Error("article/searcharticle：页数与页面数量大小参数有误，校验未通过！",
			zap.Int("page", page), zap.Int("pagesize", pageSize))
		return l.articleResponse.RPCFailSearchArticleResponseWithErr(err)
	}
	if page == 0 || pageSize == 0 {
		l.logger.Error("article/searcharticle：页数与页面数量大小参数有误，实际数量为0！")
		return l.articleResponse.RPCFailSearchArticleResponse("参数异常！", errcode.BadPageData)
	}

	like, err := l.commonUtils.ValidateQuery(in.Query)
	if err != nil {
		l.logger.Error("article/searcharticle：查询内容校验未通过！",
			zap.String("Query", in.Query), zap.String("ERROR", err.Error()))
		return l.articleResponse.RPCFailSearchArticleResponseWithErr(err)
	}
	if like == "" {
		l.logger.Error("article/searcharticle：查询内容模糊匹配模板为空！")
		return l.articleResponse.RPCFailSearchArticleResponse("参数异常！", errcode.BadPageData)
	}
	// 查库
	searchArticleList, total, err := l.articleRepository.SearchArticle(l.ctx, like, page, pageSize)
	if err != nil {
		l.logger.Error("article/searcharticle：数据库查询文章操作失败",
			zap.String("ERROR", err.Error()), zap.String("Query", in.Query))
		return l.articleResponse.RPCFailSearchArticleResponse("数据库操作异常！", errcode.ServerError)
	}
	if searchArticleList == nil || total == 0 {
		l.logger.Error("article/searcharticle：数据库查询文章得到的数量为空！",
			zap.String("Query", in.Query))
		return l.articleResponse.RPCSuccessSearchArticleResponse("未查询到相关文章！", nil, total)
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

	l.logger.Info("article/searcharticle：查询文章列表完成",
		zap.String("Query", in.Query), zap.Int64("TotalCount", total))
	return l.articleResponse.RPCSuccessSearchArticleResponse("查询列表加载完成！", resultList, total)
}
