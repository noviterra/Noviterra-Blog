package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"go.uber.org/zap"
)

type getDetailArticleBloomFilter interface {
	IsExistArticle(data string) bool
}

type getDetailArticleCommonUtils interface {
	CheckArticleID(articleID string) (bool, error)
}

type getDetailArticleLogger interface {
	Error(msg string, fileds ...zap.Field)
	Info(msg string, fileds ...zap.Field)
}

type getDetailArticleResponse interface {
	RPCSuccessGetArticleDetailResponse(msg string, article *articlerpc.Article) (*articlerpc.GetArticleDetailResponse, error)
	RPCFailGetArticleDetailResponse(msg string, code int64) (*articlerpc.GetArticleDetailResponse, error)
	RPCFailGetArticleDetailResponseWithErr(err error) (*articlerpc.GetArticleDetailResponse, error)
}

type getDetailArticleRepository interface {
	GetArticleDetail(ctx context.Context, articleID string) (*model.ArticleInfo, error)
}

type getDetailArticleCache interface {
	HGetArticle(ctx context.Context, articleID string) (*model.Article, error)
}
type GetArticleDetailLogic struct {
	ctx context.Context

	bloomFilter       getDetailArticleBloomFilter
	commonUtils       getDetailArticleCommonUtils
	logger            getDetailArticleLogger
	articleResponse   getDetailArticleResponse
	articleRepository getDetailArticleRepository
	articleCache      getDetailArticleCache
}

func NewGetArticleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleDetailLogic {
	return &GetArticleDetailLogic{
		ctx:               ctx,
		bloomFilter:       svcCtx.ArticleBloomFilter,
		commonUtils:       svcCtx.CommonUtils,
		logger:            svcCtx.Logger,
		articleResponse:   svcCtx.ArticleResponse,
		articleRepository: svcCtx.ArticleRepository,
		articleCache:      svcCtx.ArticleCache,
	}
}

// 查询文章详细信息
func (l *GetArticleDetailLogic) GetArticleDetail(in *articlerpc.GetArticleDetailRequest) (*articlerpc.GetArticleDetailResponse, error) {
	// todo: add your logic here and delete this line
	article := new(articlerpc.Article)
	// 获取用户详细内容 不需要校验用户权限 游客、普通用户都可以访问该资源
	// 参数校验
	ok, err := l.commonUtils.CheckArticleID(in.ArticleID)
	if err != nil {
		l.logger.Error("article/getArticleDetail：请求文章ID校验失败", zap.String("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.RPCFailGetArticleDetailResponseWithErr(err)
	}
	if !ok {
		l.logger.Error("article/getArticleDetail：请求文章ID校验未通过", zap.String("ArticleID", in.ArticleID))
		return l.articleResponse.RPCFailGetArticleDetailResponse("文章不存在", errcode.NotFound)
	}
	// 布隆过滤器 -> 缓存 -> 仓库
	isExistArticle := l.bloomFilter.IsExistArticle(in.ArticleID)
	if !isExistArticle { // 为false则表示一定不存在 true表示可能存在需要进一步校验
		l.logger.Error("article/getArticleDetail：访问的ID在布隆过滤器中不存在，已过滤", zap.String("ArticleID", in.ArticleID))
		return l.articleResponse.RPCFailGetArticleDetailResponse("文章不存在", errcode.NotFound)
	}

	// 缓存
	articleCache, err := l.articleCache.HGetArticle(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("article/getArticleDetail：缓存根据文章ID查询哈希表中的数据失败", zap.String("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.RPCFailGetArticleDetailResponse("缓存操作异常", errcode.ServerError)
	}
	if articleCache != nil { // 缓存命中
		article.ArticleID = articleCache.ArticleID
		article.PublisherID = articleCache.PublisherID
		article.Content = articleCache.Content
		article.Information = articleCache.Information
		article.Cover = articleCache.Cover
		article.Tags = articleCache.Tags
		article.CreatedAt = articleCache.CreatedAt
		article.UpdatedAt = articleCache.UpdatedAt

		l.logger.Info("article/getArticleDetail：缓存命中成功，返回缓存中的数据", zap.String("ArticleID", in.ArticleID), zap.Int64("ArticleUpdatedAt", article.UpdatedAt))
		return l.articleResponse.RPCSuccessGetArticleDetailResponse("文章加载成功！", article)
	}

	// 缓存未命中
	// 数据库
	articleInfo, err := l.articleRepository.GetArticleDetail(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("article/getArticleDetail：数据库根据文章ID查询文章失败", zap.String("Article", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.RPCFailGetArticleDetailResponse("数据库操作失败！", errcode.ServerError)
	}
	if articleInfo == nil {
		l.logger.Error("article/getArticleDetail：用户请求的文章在数据库中不存在，但经过布隆过滤器、缓存", zap.String("Article", in.ArticleID))
		return l.articleResponse.RPCFailGetArticleDetailResponse("访问的文章不存在或已被删除！", errcode.NotFound)
	}
	l.logger.Info("article/getArticleDetail：最终在数据库中查询到了文章", zap.String("ArticleID", in.ArticleID))

	// 数据映射
	article.ArticleID = articleInfo.ArticleID
	article.PublisherID = articleInfo.PublisherID
	article.Content = articleInfo.Content
	article.Information = articleInfo.Information
	article.Cover = articleInfo.Cover
	article.Tags = articleInfo.Tags
	// 时间戳 秒级
	article.CreatedAt = articleInfo.CreatedAt.Unix()
	article.UpdatedAt = articleInfo.UpdatedAt.Unix()

	return l.articleResponse.RPCSuccessGetArticleDetailResponse("文章加载成功！", article)
}
