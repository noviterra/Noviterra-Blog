package svc

import (
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/app/site/siterpc/internal/config"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/repo"
	"Noviterra/server/utils/response"
	"Noviterra/server/utils/validate"
	"Noviterra/server/utils/zaplogger"
)

type ServiceContext struct {
	Config                  config.Config
	CtxData                 *ctxdata.CtxData
	SiteValidate            *validate.SiteValidate
	SiteResponse            *response.SiteResponse
	Logger                  *zaplogger.Logger
	SiteRepository          *repo.SiteRepository
	IdentityRepository      *repo.IdentityRepository
	ServiceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func NewServiceContext(c config.Config) *ServiceContext {
	tool := initTool()
	store := initStore(c)
	aspect := initAspect(c, tool)

	return &ServiceContext{
		Config:                  c,
		CtxData:                 tool.ctxdata,
		SiteValidate:            tool.siteValidate,
		SiteResponse:            tool.siteResponse,
		Logger:                  tool.logger,
		SiteRepository:          store.siteRepository,
		IdentityRepository:      store.identityRepository,
		ServiceTokenInterceptor: aspect.serviceTokenInterceptor,
	}
}

type toolCollection struct {
	ctxdata      *ctxdata.CtxData
	siteValidate *validate.SiteValidate
	siteResponse *response.SiteResponse
	logger       *zaplogger.Logger
}

func initTool() toolCollection {
	cd := ctxdata.NewCtxData()
	siteValidate := validate.NewSiteValidate()
	siteResponse := response.NewSiteResponse()
	logger := zaplogger.NewLogger()

	return toolCollection{
		ctxdata:      cd,
		siteValidate: siteValidate,
		siteResponse: siteResponse,
		logger:       logger,
	}
}

type storeCollection struct {
	siteRepository     *repo.SiteRepository
	identityRepository *repo.IdentityRepository
}

func initStore(c config.Config) storeCollection {
	repositoryConfig := repo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	}

	siteRepository := repo.NewSiteRepository(repositoryConfig, "site")
	identityRepository := repo.NewIdentityRespository(repositoryConfig, "identity")

	return storeCollection{
		siteRepository:     siteRepository,
		identityRepository: identityRepository,
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
