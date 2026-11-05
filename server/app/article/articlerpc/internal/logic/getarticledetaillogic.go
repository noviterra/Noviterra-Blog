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
	IsExistArticle(data int64) bool
}

type getDetailArticleValidate interface {
	ValidateSnowID(id int64) (int64, error)
}

type getDetailArticleLogger interface {
	Error(msg string, fileds ...zap.Field)
	Info(msg string, fileds ...zap.Field)
}

type getDetailArticleResponse interface {
	SuccessGetArticleDetailResponse(resp *articlerpc.GetArticleDetailResponse,
		msg string, article *model.Article) (*articlerpc.GetArticleDetailResponse, error)
	FailGetArticleDetailResponse(resp *articlerpc.GetArticleDetailResponse,
		msg string, status int32, code int64) (*articlerpc.GetArticleDetailResponse, error)
	FailGetArticleDetailResponseWithErr(resp *articlerpc.GetArticleDetailResponse,
		err error) (*articlerpc.GetArticleDetailResponse, error)
}

type getDetailArticleRepository interface {
	GetArticleDetail(ctx context.Context, articleID int64) (*model.Article, error)
}

type getDetailArticleCache interface {
	HGetArticle(ctx context.Context, articleID int64) (*model.Article, error)
}
type GetArticleDetailLogic struct {
	ctx context.Context

	bloomFilter       getDetailArticleBloomFilter
	articleValidate   getDetailArticleValidate
	logger            getDetailArticleLogger
	articleResponse   getDetailArticleResponse
	articleRepository getDetailArticleRepository
	articleCache      getDetailArticleCache
}

func NewGetArticleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleDetailLogic {
	return &GetArticleDetailLogic{
		ctx:               ctx,
		bloomFilter:       svcCtx.ArticleBloomFilter,
		articleValidate:   svcCtx.ArticleValidate,
		logger:            svcCtx.Logger,
		articleResponse:   svcCtx.ArticleResponse,
		articleRepository: svcCtx.ArticleRepository,
		articleCache:      svcCtx.ArticleCache,
	}
}

// 查询文章详细信息
func (l *GetArticleDetailLogic) GetArticleDetail(in *articlerpc.GetArticleDetailRequest) (*articlerpc.GetArticleDetailResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(articlerpc.GetArticleDetailResponse)
	// 获取用户详细内容 不需要校验用户权限 游客、普通用户都可以访问该资源
	// 参数校验
	articleID, err := l.articleValidate.ValidateSnowID(in.ArticleID)
	if err != nil {
		l.logger.Error("article/getArticleDetail：请求文章ID校验失败", zap.Int64("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.FailGetArticleDetailResponseWithErr(resp, err)
	}
	if articleID == 0 {
		l.logger.Error("article/getArticleDetail：请求文章ID校验未通过", zap.Int64("ArticleID", in.ArticleID))
		return l.articleResponse.FailGetArticleDetailResponse(resp, "文章不存在", errcode.NotFound, errcode.Failed)
	}
	// 布隆过滤器 -> 缓存 -> 仓库
	isExistArticle := l.bloomFilter.IsExistArticle(in.ArticleID)
	if !isExistArticle { // 为false则表示一定不存在 true表示可能存在需要进一步校验
		l.logger.Error("article/getArticleDetail：访问的ID在布隆过滤器中不存在，已过滤", zap.Int64("ArticleID", in.ArticleID))
		return l.articleResponse.FailGetArticleDetailResponse(resp, "文章不存在", errcode.NotFound, errcode.Failed)
	}

	// 缓存
	articleCache, err := l.articleCache.HGetArticle(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("article/getArticleDetail：缓存根据文章ID查询哈希表中的数据失败", zap.Int64("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.FailGetArticleDetailResponse(resp, "缓存操作异常", errcode.ServerError, errcode.Failed)
	}
	if articleCache != nil { // 缓存命中
		l.logger.Info("article/getArticleDetail：缓存命中成功，返回缓存中的数据",
			zap.Int64("ArticleID", in.ArticleID))
		return l.articleResponse.SuccessGetArticleDetailResponse(resp, "文章加载成功！", articleCache)
	}

	// 缓存未命中
	// 数据库
	articleInfo, err := l.articleRepository.GetArticleDetail(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("article/getArticleDetail：数据库根据文章ID查询文章失败", zap.Int64("Article", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.FailGetArticleDetailResponse(resp, "数据库操作失败！", errcode.ServerError, errcode.Failed)
	}
	if articleInfo == nil {
		l.logger.Error("article/getArticleDetail：用户请求的文章在数据库中不存在，但经过布隆过滤器、缓存", zap.Int64("Article", in.ArticleID))
		return l.articleResponse.FailGetArticleDetailResponse(resp, "访问的文章不存在或已被删除！", errcode.NotFound, errcode.Failed)
	}
	l.logger.Info("article/getArticleDetail：最终在数据库中查询到了文章", zap.Int64("ArticleID", in.ArticleID))

	return l.articleResponse.SuccessGetArticleDetailResponse(resp, "文章加载成功！", articleInfo)
}
