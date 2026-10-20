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
	Config             config.Config
	SnowFlake          *snowflake.SnowFlake
	CtxData            *ctxdata.CtxData
	Interceptor        *interceptor.Interceptor
	UserRepository     *repo.UserRepository
	IdentityRepository *repo.IdentityRepository
	CommonUtils        *common.CommonUtils
	ArticleRepository  *repo.ArticleRepository
	ArticleCache       *cache.ArticleCache
	ArticleResponse    *response.ArticleResponse
	Logger             *zaplogger.Logger
	ArticleBloomFilter *bloomfilter.ArticleBloomFilter
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 雪花ID生成器
	sf, err := snowflake.NewSnowFlake(1)
	if err != nil {
		panic(err)
	}

	// 工具插件
	cu := common.NewCommonUtils()

	// 上下文存储器 用于rpc层的身份校验
	cd := ctxdata.NewCtxData()

	// 拦截器
	itp := interceptor.NewInterceptor(cd)

	// 响应器
	articleResponse := response.NewArticleResponse()

	// article缓存
	cacheConfig := cache.RedisConfig{
		Addr:     c.Cache.Addr,
		Password: c.Cache.Password,
		DB:       c.Cache.DB,
	}
	articleCache := cache.NewArticleCache(cacheConfig, cu)

	// 仓库操作实例
	repoConfig := repo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	}
	userRepo := repo.NewUserRepository(repoConfig, "user")
	identityRepo := repo.NewIdentityRespository(repoConfig, "identity")

	// 日志写入器
	logger := zaplogger.NewLogger()

	// 文章仓库操作实例
	articleRepo := repo.NewArticleRepository(repoConfig, "article")

	// 文章布隆过滤器
	articleBloomFilter, err := bloomfilter.NewArticleBloomFilter(1000, 0.001)
	if err != nil {
		panic(fmt.Errorf("布隆过滤器初始化失败！，ERROR：%s", err))
	}

	return &ServiceContext{
		Config:             c,
		SnowFlake:          sf,
		CtxData:            cd,
		Interceptor:        itp,
		UserRepository:     userRepo,
		CommonUtils:        cu,
		ArticleRepository:  articleRepo,
		ArticleCache:       articleCache,
		ArticleResponse:    articleResponse,
		Logger:             logger,
		ArticleBloomFilter: articleBloomFilter,
		IdentityRepository: identityRepo,
	}
}
