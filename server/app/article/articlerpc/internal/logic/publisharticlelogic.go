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
		l.logger.Error("上下文获取操作用户ID失败")
		return l.articleResponse.FailPublishArticleResponse(resp, "用户ID为空，请正常访问资源！",
			errcode.Unauthorized, errcode.GetUserInfoFail)
	}

	// 校验用户身份 查库
	role, err := l.identityRepository.GetUserRole(l.ctx, publisherID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据用户ID查询用户身份权限等级失败",
			zap.String("publisher-info", publisherID),
			zap.Error(err),
		)
		return l.articleResponse.FailPublishArticleResponse(resp, "数据库查询失败！",
			errcode.InternalError, errcode.GetUserInfoFail)
	}
	if role < 1 {
		l.logger.Error("用户权限不足，无法发布文章",
			zap.String("publisher-info", publisherID),
			zap.Int("publisher-role", role),
		)
		return l.articleResponse.FailPublishArticleResponse(resp, "权限不足，请求失败！",
			errcode.Forbidden, errcode.ArticleOperationOutOfBounds)
	}

	// 校验与清洁文章内容 此处取消校验文章封面 封面校验在updatearticlecover.go位置
	validateInput := validate.InputValidateArticle{
		Title:       in.Title,
		Information: in.Information,
		Content:     in.Content,
		Tags:        in.Tags,
	}
	cleanArticle, err := l.commonUtils.ValidateArticle(validateInput)
	if err != nil { // 该方法返回的err全是自定义 是手动校验后硬编码的 因此可以直接返回error
		l.logger.Error("校验InputValidateArticle参数失败",
			zap.String("publisher-info", publisherID),
			zap.Any("validate", validateInput),
			zap.Error(err),
		)
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
		l.logger.Error("数据库操作失败：插入文章数据失败",
			zap.Int64("article-id", articleID),
			zap.String("publisher-info", publisherID),
			zap.Error(err),
		)
		return l.articleResponse.FailPublishArticleResponse(resp, "数据库操作失败！",
			errcode.InternalError, errcode.PublishArticleFail)
	}

	// 加入到缓存中的zset中 根据访问量或点赞量或收藏量进行排行top 对应三个top榜
	err = l.articleCache.AddArticleToRank(l.ctx, articleID)
	if err != nil { // 全部缓存失败都不抛出 继续进行
		l.logger.Warn("缓存操作失败：文章加入排行榜失败",
			zap.Int64("article-id", articleID),
			zap.String("publisher-info", publisherID),
			zap.Error(err),
		)
		// return l.articleResponse.RPCFailResponse("缓存失败！", errcode.ServerError)
	}

	// 加入到布隆过滤器中
	l.bloomFilter.AddArticle(articleID)

	l.logger.Info("文章发布成功",
		zap.Int64("article-id", articleID),
		zap.String("publisher-info", publisherID),
		zap.Int("publisher-role", role),
	)
	return l.articleResponse.SuccessPublishArticleResponse(resp, "文章发布成功！", articleID)
}
