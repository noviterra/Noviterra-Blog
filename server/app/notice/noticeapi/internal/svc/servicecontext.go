// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/app/internal/middleware"
	"Noviterra/server/app/notice/noticeapi/internal/config"
	"Noviterra/server/app/notice/noticerpc/noticeservice"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/jwtauth"
	"Noviterra/server/utils/repo"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                  config.Config
	NoticeModel             noticeservice.NoticeService
	ServiceTokenInterceptor *interceptor.ServiceTokenInterceptor
	AuthMiddleware          *middleware.AuthMiddleware
	CtxData                 *ctxdata.CtxData
}

func NewServiceContext(c config.Config) *ServiceContext {
	rpc := initRPC(c)
	tool := initTool(c)
	store := initStore(c)
	aspect := initAspect(c, tool, store)

	return &ServiceContext{
		Config:                  c,
		NoticeModel:             rpc.noticeModel,
		ServiceTokenInterceptor: aspect.serviceTokenInterceptor,
		AuthMiddleware:          aspect.authMiddleware,
		CtxData:                 tool.ctxdata,
	}
}

type rpcCollection struct {
	noticeModel noticeservice.NoticeService
}

func initRPC(c config.Config) rpcCollection {
	noticeService := noticeservice.NewNoticeService(zrpc.MustNewClient(c.NoticeRPC))

	return rpcCollection{
		noticeModel: noticeService,
	}
}

type toolCollection struct {
	ctxdata *ctxdata.CtxData
}

func initTool(c config.Config) toolCollection {
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
	authMiddleware          *middleware.AuthMiddleware
	serviceTokenInterceptor *interceptor.ServiceTokenInterceptor
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
		authMiddleware:          authMiddleware,
		serviceTokenInterceptor: serviceTokenInterceptor,
	}
}
