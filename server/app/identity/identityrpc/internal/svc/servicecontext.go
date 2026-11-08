package svc

import (
	"Noviterra/server/app/identity/identityrpc/internal/config"
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/utils/bcryptinfo"
	"Noviterra/server/utils/cache"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/jwtauth"
	"Noviterra/server/utils/repo"
	"Noviterra/server/utils/response"
	"Noviterra/server/utils/validate"
)

type ServiceContext struct {
	Config                  config.Config
	IdentityRepository      *repo.IdentityRepository
	NoticeCache             *cache.NoticeCache
	Authentication          *jwtauth.Authentication
	Crypt                   *bcryptinfo.Crypt
	Ctxdata                 *ctxdata.CtxData
	IdentityValidate        *validate.IdentityValidate
	IdentityResponse        *response.IdentityResponse
	ServiceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func NewServiceContext(c config.Config) *ServiceContext {
	tool := initTool(c)
	store := initStore(c)
	aspect := initAspect(c, tool)

	return &ServiceContext{
		Config:                  c,
		Crypt:                   tool.crypt,
		Authentication:          tool.authentication,
		IdentityResponse:        tool.identityResponse,
		Ctxdata:                 tool.ctxdata,
		IdentityValidate:        tool.identityValidate,
		NoticeCache:             store.noticeCache,
		IdentityRepository:      store.identityRepository,
		ServiceTokenInterceptor: aspect.serviceTokenInterceptor,
	}
}

type toolCollection struct {
	crypt            *bcryptinfo.Crypt
	authentication   *jwtauth.Authentication // utils中的jwt操作集合 解析、生成Token操作
	identityResponse *response.IdentityResponse
	ctxdata          *ctxdata.CtxData
	identityValidate *validate.IdentityValidate
}

func initTool(c config.Config) toolCollection {
	// jwt配置
	authenticationConfig := jwtauth.AuthenticationConfig{
		AccessSecret:  c.JWT.AccessSecret,
		AccessExpire:  c.JWT.AccessExpire,
		RefreshSecret: c.JWT.RefreshSecret,
		RefreshExpire: c.JWT.RefreshExpire,
	}

	// jwt
	authentication := jwtauth.NewAuthentication(authenticationConfig)
	// 加密工具 bcrypt
	crypt := bcryptinfo.NewCrypt()
	// 认证模块响应封装
	identityResponse := response.NewIdentityResponse()
	// 上下文存储
	cd := ctxdata.NewCtxData()
	// 校验器
	identityValidate := validate.NewIdentityValidate()

	return toolCollection{
		crypt:            crypt,
		authentication:   authentication,
		identityResponse: identityResponse,
		ctxdata:          cd,
		identityValidate: identityValidate,
	}
}

type storeCollection struct {
	identityRepository *repo.IdentityRepository
	noticeCache        *cache.NoticeCache
}

func initStore(c config.Config) storeCollection {
	// 缓存配置
	cacheConfig := cache.RedisConfig{
		Addr:     c.Cache.Addr,
		Password: c.Cache.Password,
		DB:       c.Cache.DB,
	}

	// 数据库配置
	repositoryConfig := repo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	}

	// 认证模块缓存
	noticeCache := cache.NewNoticeCache(cacheConfig)
	// 认证模块数据库
	identityRespository := repo.NewIdentityRespository(repositoryConfig, "identity")

	return storeCollection{
		identityRepository: identityRespository,
		noticeCache:        noticeCache,
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
