package svc

import (
	"Noviterra/server/app/article/articlerpc/internal/config"
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/utils/bloomfilter"
	"Noviterra/server/utils/cache"
	"Noviterra/server/utils/common"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/repo"
	"Noviterra/server/utils/response"
	"Noviterra/server/utils/snowflake"
	"Noviterra/server/utils/zaplogger"
	"fmt"
)

type ServiceContext struct {
	Config                  config.Config
	CommonUtils             *common.CommonUtils
	SnowFlake               *snowflake.SnowFlake
	CtxData                 *ctxdata.CtxData
	Logger                  *zaplogger.Logger
	ArticleBloomFilter      *bloomfilter.ArticleBloomFilter
	ArticleResponse         *response.ArticleResponse
	ArticleRepository       *repo.ArticleRepository
	ArticleCache            *cache.ArticleCache
	UserRepository          *repo.UserRepository
	IdentityRepository      *repo.IdentityRepository
	ServiceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func NewServiceContext(c config.Config) *ServiceContext {
	tool := initTool()
	store := initStore(c, tool)
	aspect := initAspect(c, tool)

	return &ServiceContext{
		Config:                  c,
		SnowFlake:               tool.snowFlake,
		CtxData:                 tool.ctxData,
		CommonUtils:             tool.commonUtils,
		Logger:                  tool.logger,
		ArticleResponse:         tool.articleResponse,
		ArticleBloomFilter:      tool.articleBloomFilter,
		UserRepository:          store.userRepository,
		ArticleRepository:       store.articleRepository,
		ArticleCache:            store.articleCache,
		IdentityRepository:      store.identityRepository,
		ServiceTokenInterceptor: aspect.serviceTokenInterceptor,
	}
}

type toolCollection struct {
	commonUtils        *common.CommonUtils
	snowFlake          *snowflake.SnowFlake
	ctxData            *ctxdata.CtxData
	logger             *zaplogger.Logger
	articleBloomFilter *bloomfilter.ArticleBloomFilter
	articleResponse    *response.ArticleResponse
}

func initTool() toolCollection {
	// 雪花ID生成器
	sf, err := snowflake.NewSnowFlake(1)
	if err != nil {
		panic(fmt.Errorf("雪花ID生成器初始化失败！%w", err))
	}

	// 工具插件
	cu := common.NewCommonUtils()

	// 上下文存储器 用于rpc层的身份校验
	cd := ctxdata.NewCtxData()

	// 日志写入器
	logger := zaplogger.NewLogger()

	// 文章布隆过滤器
	articleBloomFilter, err := bloomfilter.NewArticleBloomFilter(1000, 0.001)
	if err != nil {
		panic(fmt.Errorf("布隆过滤器初始化失败！%w", err))
	}

	articleResponse := response.NewArticleResponse()

	return toolCollection{
		commonUtils:        cu,
		snowFlake:          sf,
		ctxData:            cd,
		logger:             logger,
		articleBloomFilter: articleBloomFilter,
		articleResponse:    articleResponse,
	}
}

type storeCollection struct {
	articleRepository  *repo.ArticleRepository
	articleCache       *cache.ArticleCache
	userRepository     *repo.UserRepository
	identityRepository *repo.IdentityRepository
}

func initStore(c config.Config, tool toolCollection) storeCollection {
	repositoryConfig := repo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	}

	cacheConfig := cache.RedisConfig{
		Addr:     c.Cache.Addr,
		Password: c.Cache.Password,
		DB:       c.Cache.DB,
	}

	userRepo := repo.NewUserRepository(repositoryConfig, "user")
	identityRepo := repo.NewIdentityRespository(repositoryConfig, "identity")
	articleRepo := repo.NewArticleRepository(repositoryConfig, "article")

	articleCache := cache.NewArticleCache(cacheConfig, tool.commonUtils)

	return storeCollection{
		articleRepository:  articleRepo,
		articleCache:       articleCache,
		userRepository:     userRepo,
		identityRepository: identityRepo,
	}
}

type aspectCollection struct {
	serviceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func initAspect(c config.Config, tool toolCollection) aspectCollection {
	serviceTokenInterceptorConfig := interceptor.ServiceTokenConfig{
		SecretToken: c.Interceptor.ServiceToken,
		SkipToken:   c.Interceptor.SkipToken,
	}

	serviceTokenInterceptor := interceptor.NewServiceTokenInterceptor(tool.ctxData, serviceTokenInterceptorConfig)

	return aspectCollection{
		serviceTokenInterceptor: serviceTokenInterceptor,
	}
}
