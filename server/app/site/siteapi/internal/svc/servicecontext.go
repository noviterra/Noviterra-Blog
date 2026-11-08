// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/app/internal/middleware"
	"Noviterra/server/app/notice/noticerpc/noticeservice"
	"Noviterra/server/app/site/siteapi/internal/config"
	"Noviterra/server/app/site/siterpc/siteservice"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/jwtauth"
	"Noviterra/server/utils/repo"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                  config.Config
	SiteModel               siteservice.SiteService
	NoticeModel             noticeservice.NoticeService
	CtaData                 *ctxdata.CtxData
	AuthMiddleware          *middleware.AuthMiddleware
	ServiceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func NewServiceContext(c config.Config) *ServiceContext {
	rpc := initRPC(c)
	tool := initTool()
	store := initStore(c)
	aspect := initAspect(c, tool, store)

	return &ServiceContext{
		Config:                  c,
		SiteModel:               rpc.siteModel,
		NoticeModel:             rpc.noticeModel,
		CtaData:                 tool.ctxdata,
		AuthMiddleware:          aspect.authMiddleware,
		ServiceTokenInterceptor: aspect.serviceTokenInterceptor,
	}
}

type rpcCollection struct {
	siteModel   siteservice.SiteService
	noticeModel noticeservice.NoticeService
}

func initRPC(c config.Config) rpcCollection {
	siteService := siteservice.NewSiteService(zrpc.MustNewClient(c.SiteRPC))
	noticeService := noticeservice.NewNoticeService(zrpc.MustNewClient(c.NoticeRPC))

	return rpcCollection{
		siteModel:   siteService,
		noticeModel: noticeService,
	}
}

type toolCollection struct {
	ctxdata *ctxdata.CtxData
}

func initTool() toolCollection {
	cd := ctxdata.NewCtxData()

	return toolCollection{
		ctxdata: cd,
	}
}

type storeCollection struct {
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

	identityRepository := repo.NewIdentityRespository(repositoryConfig, "identity")

	return storeCollection{
		identityRepository: identityRepository,
	}
}

type aspectCollection struct {
	serviceTokenInterceptor *interceptor.ServiceTokenInterceptor
	authMiddleware          *middleware.AuthMiddleware
}

func initAspect(c config.Config, tool toolCollection, store storeCollection) aspectCollection {
	serviceTokenInterceptorConfig := interceptor.ServiceTokenConfig{
		SecretToken: c.Interceptor.ServiceToken,
		SkipToken:   c.Interceptor.SkipToken,
	}
	jwtConfig := jwtauth.AuthenticationConfig{
		AccessSecret:  c.JWT.AccessSecret,
		AccessExpire:  c.JWT.AccessExpire,
		RefreshSecret: c.JWT.RefreshSecret,
		RefreshExpire: c.JWT.RefreshExpire,
	}
	jwtAuth := jwtauth.NewAuthentication(jwtConfig)

	serviceTokenInterceptor := interceptor.NewServiceTokenInterceptor(tool.ctxdata, serviceTokenInterceptorConfig)
	authMiddleware := middleware.NewAuthMiddleware(jwtAuth, store.identityRepository, tool.ctxdata)

	return aspectCollection{
		serviceTokenInterceptor: serviceTokenInterceptor,
		authMiddleware:          authMiddleware,
	}
}
