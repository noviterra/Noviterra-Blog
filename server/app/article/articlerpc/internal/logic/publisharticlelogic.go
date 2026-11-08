package logic

import (
	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"
	"context"

	"go.uber.org/zap"
)

type (
	publishArticleIdentityRepository interface {
		GetUserRole(ctx context.Context, id string) (int, error)
	}

	publishArticleRepository interface {
		PublishArticle(ctx context.Context, article *model.ArticleInfo, stat *model.ArticleStat) error
	}

	publishArticleCache interface {
		AddArticleToRank(ctx context.Context, articleID int64) error
	}

	publishArticleCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	publishArticleCommonUtils interface {
		ValidateArticle(in validate.InputValidateArticle) (validate.OutputValidateArticle, error)
	}

	publishArticleSnowFlake interface {
		GenerateSnowFlakeID() int64
	}

	publishArticleResponse interface {
		SuccessPublishArticleResponse(resp *articlerpc.PublishArticleResponse, msg string, articleID int64) (*articlerpc.PublishArticleResponse, error)
		FailPublishArticleResponse(resp *articlerpc.PublishArticleResponse, msg string, status int32, code int64) (*articlerpc.PublishArticleResponse, error)
		FailPublishArticleResponseWithErr(resp *articlerpc.PublishArticleResponse, err error) (*articlerpc.PublishArticleResponse, error)
	}

	publishArticleBloomFilter interface {
		AddArticle(data int64)
	}

	publishArticleLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
		Warn(msg string, fields ...zap.Field)
	}
)

type PublishArticleLogic struct {
	ctx                context.Context
	identityRepository publishArticleIdentityRepository
	articleRepository  publishArticleRepository
	ctxData            publishArticleCtxData
	commonUtils        publishArticleCommonUtils
	snowFlake          publishArticleSnowFlake
	articleCache       publishArticleCache
	articleResponse    publishArticleResponse
	logger             publishArticleLogger
	bloomFilter        publishArticleBloomFilter
}

func NewPublishArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishArticleLogic {
	return &PublishArticleLogic{
		ctx:                ctx,
		identityRepository: svcCtx.IdentityRepository,
		articleRepository:  svcCtx.ArticleRepository,
		ctxData:            svcCtx.CtxData,
		commonUtils:        svcCtx.ArticleValidate,
		snowFlake:          svcCtx.SnowFlake,
		articleCache:       svcCtx.ArticleCache,
		articleResponse:    svcCtx.ArticleResponse,
		logger:             svcCtx.Logger,
		bloomFilter:        svcCtx.ArticleBloomFilter,
	}
}

// 发布文章
func (l *PublishArticleLogic) PublishArticle(in *articlerpc.PublishArticleRequest) (*articlerpc.PublishArticleResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(articlerpc.PublishArticleResponse)
	// 从上下文中获取用户信息
	publisherID := l.ctxData.GetUserIDFromContext(l.ctx)
	if publisherID == "" {
		l.logger.Error("article/publisharticlelogic：用户ID为空，疑似绕过api层进行请求（暂未锁定IP）")
		return l.articleResponse.FailPublishArticleResponse(resp, "用户ID为空，请正常访问资源！", errcode.Unauthorized, errcode.Failed)
	}

	// 校验用户身份 查库
	role, err := l.identityRepository.GetUserRole(l.ctx, publisherID)
	if err != nil {
		l.logger.Error("article/publisharticlelogic：数据库查询失败：查询用户权限失败", zap.String("ERROR", err.Error()), zap.String("UserID", publisherID))
		return l.articleResponse.FailPublishArticleResponse(resp, "数据库查询失败！", errcode.ServerError, errcode.Failed)
	}
	if role < 1 {
		l.logger.Error("article/publisharticlelogic：用户权限越界，需要检查发布文件接口是否暴露or用户信息库是否泄露", zap.String("UserID", publisherID))
		return l.articleResponse.FailPublishArticleResponse(resp, "权限不足，请求失败！", errcode.Forbidden, errcode.Failed)
	}

	// 校验与清洁文章内容 此处取消校验文章封面 封面校验在updatearticlecover.go位置
	cleanArticle, err := l.commonUtils.ValidateArticle(validate.InputValidateArticle{
		Title:       in.Title,
		Information: in.Information,
		Content:     in.Content,
		Tags:        in.Tags,
	})
	if err != nil { // 该方法返回的err全是自定义 是手动校验后硬编码的 因此可以直接返回error
		l.logger.Error("article/publisharticlelogic：文章内容校验未通过", zap.String("ERROR", err.Error()), zap.String("PublisherID", publisherID))
		return l.articleResponse.FailPublishArticleResponseWithErr(resp, err)
	}

	// 生成文章对应的唯一ID
	articleID := l.snowFlake.GenerateSnowFlakeID()

	// 插入到数据库中
	err = l.articleRepository.PublishArticle(l.ctx, &model.ArticleInfo{
		ArticleID:   articleID,
		PublisherID: publisherID,
		Title:       cleanArticle.Title,
		Content:     cleanArticle.Content,
		Information: cleanArticle.Information,
		Tags:        cleanArticle.Tags,
	}, &model.ArticleStat{
		ArticleID:    articleID,
		ViewCount:    0,
		LikeCount:    0,
		CollectCount: 0,
	})
	if err != nil {
		l.logger.Error("article/publisharticlelogic：数据库插入文章失败", zap.String("ERROR", err.Error()))
		return l.articleResponse.FailPublishArticleResponse(resp, "数据库操作失败！", errcode.ServerError, errcode.Failed)
	}

	// todo: 数据库插入成功后缓存插入失败怎么操作
	// 加入到缓存中的zset中 根据访问量或点赞量或收藏量进行排行top 对应三个top榜
	err = l.articleCache.AddArticleToRank(l.ctx, articleID)
	if err != nil { // 全部缓存失败都不抛出 继续进行
		l.logger.Warn("article/publisharticlelogic：文章插入到排行榜失败（缓存）", zap.String("ERROR", err.Error()))
		// return l.articleResponse.RPCFailResponse("缓存失败！", errcode.ServerError)
	}

	// 加入到布隆过滤器中
	l.bloomFilter.AddArticle(articleID)

	l.logger.Info("article/publisharticlelogic：文章发布成功", zap.Int64("ArticleID", articleID), zap.String("PublisherID", publisherID))
	return l.articleResponse.SuccessPublishArticleResponse(resp, "文章发布成功！", articleID)
}
