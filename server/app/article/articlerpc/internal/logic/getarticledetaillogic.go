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
		l.logger.Error("校验文章ID参数失败",
			zap.Int64("article-id", in.ArticleID),
			zap.Error(err),
		)
		return l.articleResponse.FailGetArticleDetailResponseWithErr(resp, err)
	}
	if articleID == 0 {
		l.logger.Error("校验文章ID失败：文章ID为空",
			zap.Int64("article-id", in.ArticleID),
		)
		return l.articleResponse.FailGetArticleDetailResponse(resp, "文章不存在",
			errcode.NotFound, errcode.ValidateArticleIDFail)
	}
	// 布隆过滤器 -> 缓存 -> 仓库
	isExistArticle := l.bloomFilter.IsExistArticle(in.ArticleID)
	if !isExistArticle { // 为false则表示一定不存在 true表示可能存在需要进一步校验
		l.logger.Error("布隆过滤器校验失败：文章不存在",
			zap.Int64("article-id", articleID),
		)
		return l.articleResponse.FailGetArticleDetailResponse(resp, "文章不存在",
			errcode.NotFound, errcode.ArticleNotFound)
	}

	// 缓存
	articleCache, err := l.articleCache.HGetArticle(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("缓存操作失败：根据文章ID查询文章缓存失败",
			zap.Int64("article-id", articleID),
			zap.Error(err),
		)
		return l.articleResponse.FailGetArticleDetailResponse(resp, "缓存操作异常",
			errcode.InternalError, errcode.GetArticleCacheFail)
	}
	if articleCache != nil { // 缓存命中
		l.logger.Info("文章缓存命中",
			zap.Int64("article-id", articleID),
		)
		return l.articleResponse.SuccessGetArticleDetailResponse(resp, "文章加载成功！", articleCache)
	}

	// 缓存未命中
	// 数据库
	articleInfo, err := l.articleRepository.GetArticleDetail(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据文章ID查询文章详情失败",
			zap.Int64("article-id", articleID),
			zap.Error(err),
		)
		return l.articleResponse.FailGetArticleDetailResponse(resp, "数据库操作失败！",
			errcode.InternalError, errcode.GetArticleInfoFail)
	}
	if articleInfo == nil {
		l.logger.Error("数据库查询到的文章信息为空，文章不存在",
			zap.Int64("article-id", articleID),
		)
		return l.articleResponse.FailGetArticleDetailResponse(resp, "访问的文章不存在或已被删除！",
			errcode.NotFound, errcode.ArticleNotFound)
	}
	l.logger.Info("数据库查询文章详情成功",
		zap.Int64("article-id", articleID),
	)

	return l.articleResponse.SuccessGetArticleDetailResponse(resp, "文章加载成功！", articleInfo)
}
