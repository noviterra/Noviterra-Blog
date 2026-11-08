package svc

import (
	"Noviterra/server/app/comment/commentrpc/internal/config"
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/utils/cache"
	"Noviterra/server/utils/common"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/repo"
	"Noviterra/server/utils/response"
	"Noviterra/server/utils/snowflake"
	"Noviterra/server/utils/validate"
	"Noviterra/server/utils/woker"
	"context"
	"fmt"
)

type ServiceContext struct {
	Config                  config.Config
	CommentCache            *cache.CommentCache
	CommentRepository       *repo.CommentRepository
	UserRepository          *repo.UserRepository
	CommentResponse         *response.CommentResponse
	CtxData                 *ctxdata.CtxData
	CommentValidate         *validate.CommentValidate
	ArticleValidate         *validate.ArticleValidate
	Snowfake                *snowflake.SnowFlake
	DataDeduplication       *common.DataDeduplication
	ServiceTokenInterceptor *interceptor.ServiceTokenInterceptor
	CommentWoker            *woker.LikeSyncWoker
}

func NewServiceContext(c config.Config) *ServiceContext {
	tool := initTool()
	store := initStore(c, tool)
	aspect := initAspect(c, tool)
	woke := initWoke(store)
	woke.commentWoker.Start(context.Background())

	return &ServiceContext{
		Config:                  c,
		CommentCache:            store.commentCache,
		CommentRepository:       store.commentRepository,
		UserRepository:          store.userRepository,
		CtxData:                 tool.ctxdata,
		CommentResponse:         tool.commentResponse,
		CommentValidate:         tool.commentValidate,
		ArticleValidate:         tool.articleValidate,
		Snowfake:                tool.snowfake,
		DataDeduplication:       tool.dataDeduplication,
		ServiceTokenInterceptor: aspect.serviceTokenInterceptor,
		CommentWoker:            woke.commentWoker,
	}
}

type storeCollection struct {
	commentCache      *cache.CommentCache
	commentRepository *repo.CommentRepository
	userRepository    *repo.UserRepository
}

func initStore(c config.Config, tool toolCollection) storeCollection {
	cacheConfig := cache.RedisConfig{
		Addr:     c.Cache.Addr,
		Password: c.Cache.Password,
		DB:       c.Cache.DB,
	}
	repositoryConfig := repo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	}

	commentCache := cache.NewCommentCache(cacheConfig, tool.typeConversion)
	commentRepository := repo.NewCommentRepository(repositoryConfig, "comment", tool.commonUtils)
	userRepository := repo.NewUserRepository(repositoryConfig, "user")

	return storeCollection{
		commentCache:      commentCache,
		commentRepository: commentRepository,
		userRepository:    userRepository,
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

	serviceTokenInterceptor := interceptor.NewServiceTokenInterceptor(tool.ctxdata, serviceTokenInterceptorConfig)

	return aspectCollection{
		serviceTokenInterceptor: serviceTokenInterceptor,
	}
}

type toolCollection struct {
	commentResponse   *response.CommentResponse
	ctxdata           *ctxdata.CtxData
	commentValidate   *validate.CommentValidate
	commonUtils       *common.CommonUtils
	articleValidate   *validate.ArticleValidate
	snowfake          *snowflake.SnowFlake
	typeConversion    *common.TypesConversion
	dataDeduplication *common.DataDeduplication
}

func initTool() toolCollection {
	commentResponse := response.NewCommentResponse()
	cd := ctxdata.NewCtxData()
	commonUtils := common.NewCommonUtils()
	identityValidate := validate.NewIdentityValidate()
	articleValidate := validate.NewArticleValidate(identityValidate)
	commentValidate := validate.NewCommentValidate(articleValidate, articleValidate)
	typeConversion := common.NewTypesConversion()
	dataDeduplication := common.NewDataDeduplication()
	sf, err := snowflake.NewSnowFlake(1)
	if err != nil {
		panic(fmt.Errorf("雪花ID生成器初始化失败！%w", err))
	}

	return toolCollection{
		commentResponse:   commentResponse,
		ctxdata:           cd,
		commentValidate:   commentValidate,
		commonUtils:       commonUtils,
		snowfake:          sf,
		typeConversion:    typeConversion,
		dataDeduplication: dataDeduplication,
		articleValidate:   articleValidate,
	}
}

type wokeCollection struct {
	commentWoker *woker.LikeSyncWoker
}

func initWoke(store storeCollection) wokeCollection {
	commentWoker := woker.NewLikeSyncWoker(store.commentCache, store.commentRepository)

	return wokeCollection{
		commentWoker: commentWoker,
	}
}
