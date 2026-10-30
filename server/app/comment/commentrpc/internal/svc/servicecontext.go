package svc

import (
	"Noviterra/server/app/comment/commentrpc/internal/config"
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/utils/cache"
	"Noviterra/server/utils/common"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/response"
	"Noviterra/server/utils/validate"
)

type ServiceContext struct {
	Config                  config.Config
	CommentCache            *cache.CommentCache
	CommentResponse         *response.CommentResponse
	CtxData                 *ctxdata.CtxData
	CommentValidate         *validate.CommentValidate
	ServiceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func NewServiceContext(c config.Config) *ServiceContext {
	tool := initTool()
	store := initStore(c)
	aspect := initAspect(c, tool)

	return &ServiceContext{
		Config:                  c,
		CommentCache:            store.commentCache,
		CtxData:                 tool.ctxdata,
		CommentResponse:         tool.commentResponse,
		CommentValidate:         tool.commentValidate,
		ServiceTokenInterceptor: aspect.serviceTokenInterceptor,
	}
}

type storeCollection struct {
	commentCache *cache.CommentCache
}

func initStore(c config.Config) storeCollection {
	cacheConfig := cache.RedisConfig{
		Addr:     c.Cache.Addr,
		Password: c.Cache.Password,
		DB:       c.Cache.DB,
	}

	commentCache := cache.NewCommentCache(cacheConfig)

	return storeCollection{
		commentCache: commentCache,
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
	}
}
