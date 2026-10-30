package logic

import (
	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/common"
	"Noviterra/server/utils/errcode"
	"context"

	"go.uber.org/zap"
)

type publishArticleIdentityRepository interface {
	GetUserRole(id string) (int, error)
}

type publishArticleRepository interface {
	PublishArticle(ctx context.Context, article *model.ArticleInfo, stat *model.ArticleStat) error
}

type publishArticleCache interface {
	AddArticleToRank(ctx context.Context, articleID string) error
}

type publishArticleCtxData interface {
	GetUserIDFromContext(ctx context.Context) string
}

type publishArticleCommonUtils interface {
	ValidateArticle(in *common.ValidateArticleInput) (*common.ValidateArticleOutput, error)
	ToString(num int64) string
}

type publishArticleSnowFlake interface {
	GenerateSnowFlakeID() int64
}

type publishArticleResponse interface {
	RPCSuccessPublishArticleResponse(msg string, articleID string) (*articlerpc.PublishArticleResponse, error)
	RPCFailPublishArticleResponse(msg string, code int64) (*articlerpc.PublishArticleResponse, error)
	RPCFailPublishArticleResponseWithErr(err error) (*articlerpc.PublishArticleResponse, error)
}

type publishArticleBloomFilter interface {
	AddArticle(data string)
}

type publishArticleLogger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
}

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
		commonUtils:        svcCtx.CommonUtils,
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
	// 从上下文中获取用户信息
	publisherID := l.ctxData.GetUserIDFromContext(l.ctx)
	if publisherID == "" {
		l.logger.Error("article/publisharticlelogic：用户ID为空，疑似绕过api层进行请求（暂未锁定IP）")
		return l.articleResponse.RPCFailPublishArticleResponse("用户ID为空，请正常访问资源！", errcode.Unauthorized)
	}

	// 校验用户身份 查库
	role, err := l.identityRepository.GetUserRole(publisherID)
	if err != nil {
		l.logger.Error("article/publisharticlelogic：数据库查询失败：查询用户权限失败", zap.String("ERROR", err.Error()), zap.String("UserID", publisherID))
		return l.articleResponse.RPCFailPublishArticleResponse("数据库查询失败！", errcode.ServerError)
	}
	if role < 1 {
		l.logger.Error("article/publisharticlelogic：用户权限越界，需要检查发布文件接口是否暴露or用户信息库是否泄露", zap.String("UserID", publisherID))
		return l.articleResponse.RPCFailPublishArticleResponse("权限不足，请求失败！", errcode.Forbidden)
	}

	// 校验与清洁文章内容 此处取消校验文章封面 封面校验在updatearticlecover.go位置
	cleanArticle, err := l.commonUtils.ValidateArticle(&common.ValidateArticleInput{
		Title:       in.Title,
		Information: in.Information,
		Content:     in.Content,
		Tags:        in.Tags,
	})
	if err != nil { // 该方法返回的err全是自定义 是手动校验后硬编码的 因此可以直接返回error
		l.logger.Error("article/publisharticlelogic：文章内容校验未通过", zap.String("ERROR", err.Error()), zap.String("PublisherID", publisherID))
		return l.articleResponse.RPCFailPublishArticleResponseWithErr(err)
	}

	// 生成文章对应的唯一ID 转为字符串类型
	articleID := l.commonUtils.ToString(l.snowFlake.GenerateSnowFlakeID())

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
		return l.articleResponse.RPCFailPublishArticleResponse("数据库操作失败！", errcode.ServerError)
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

	l.logger.Info("article/publisharticlelogic：文章发布成功", zap.String("ArticleID", articleID), zap.String("PublisherID", publisherID))
	return l.articleResponse.RPCSuccessPublishArticleResponse("文章发布成功！", articleID)
}
