package svc

import (
	"Noviterra/server/app/comment/commentrpc/internal/config"
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/utils/cache"
	"Noviterra/server/utils/common"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/repo"
	"Noviterra/server/utils/response"
	"Noviterra/server/utils/validate"
	"Noviterra/server/utils/woker"
	"context"
)

type ServiceContext struct {
	Config                  config.Config
	CommentCache            *cache.CommentCache
	CommentResponse         *response.CommentResponse
	CtxData                 *ctxdata.CtxData
	CommentValidate         *validate.CommentValidate
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
		CtxData:                 tool.ctxdata,
		CommentResponse:         tool.commentResponse,
		CommentValidate:         tool.commentValidate,
		ServiceTokenInterceptor: aspect.serviceTokenInterceptor,
		CommentWoker:            woke.commentWoker,
	}
}

type storeCollection struct {
	commentCache      *cache.CommentCache
	commentRepository *repo.CommentRepository
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

	commentCache := cache.NewCommentCache(cacheConfig)
	commentRepository := repo.NewCommentRepository(repositoryConfig, "comment", tool.commonUtils)

	return storeCollection{
		commentCache:      commentCache,
		commentRepository: commentRepository,
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
	commentResponse *response.CommentResponse
	ctxdata         *ctxdata.CtxData
	commentValidate *validate.CommentValidate
	commonUtils     *common.CommonUtils
}

func initTool() toolCollection {
	commentResponse := response.NewCommentResponse()
	cd := ctxdata.NewCtxData()
	commonUtils := common.NewCommonUtils()
	articleValidate := validate.NewArticleValidate(commonUtils)
	identityValidate := validate.NewIdentityValidate()
	commentValidate := validate.NewCommentValidate(articleValidate, identityValidate)

	return toolCollection{
		commentResponse: commentResponse,
		ctxdata:         cd,
		commentValidate: commentValidate,
		commonUtils:     commonUtils,
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
