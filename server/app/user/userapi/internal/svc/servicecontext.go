// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/app/internal/middleware"
	"Noviterra/server/app/user/userapi/internal/config"
	"Noviterra/server/app/user/userrpc/userservice"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/jwtauth"
	"Noviterra/server/utils/repo"
	"Noviterra/server/utils/zaplogger"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                  config.Config
	Logger                  *zaplogger.Logger
	UserModel               userservice.UserService
	CtxData                 *ctxdata.CtxData
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
		UserModel:               rpc.userModel,
		Logger:                  tool.logger,
		CtxData:                 tool.ctxdata,
		AuthMiddleware:          aspect.authMiddleware,
		ServiceTokenInterceptor: aspect.serviceTokenInterceptor,
	}
}

type rpcCollection struct {
	userModel userservice.UserService
}

func initRPC(c config.Config) rpcCollection {
	userService := userservice.NewUserService(zrpc.MustNewClient(c.UserRpc))

	return rpcCollection{
		userModel: userService,
	}
}

type toolCollection struct {
	ctxdata *ctxdata.CtxData
	logger  *zaplogger.Logger
}

func initTool() toolCollection {
	cd := ctxdata.NewCtxData()
	logger := zaplogger.NewLogger()

	return toolCollection{
		ctxdata: cd,
		logger:  logger,
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
	jwtConfig := jwtauth.AuthenticationConfig{
		AccessSecret:  c.JWT.AccessSecret,
		AccessExpire:  c.JWT.AccessExpire,
		RefreshSecret: c.JWT.RefreshSecret,
		RefreshExpire: c.JWT.RefreshExpire,
	}
	jwtAuth := jwtauth.NewAuthentication(jwtConfig)
	serviceTokenInterceptorConfig := interceptor.ServiceTokenConfig{
		SecretToken: c.Interceptor.ServiceToken,
		SkipToken:   c.Interceptor.SkipToken,
	}

	authMiddleware := middleware.NewAuthMiddleware(jwtAuth, store.identityRepository, tool.ctxdata)
	serviceTokenInterceptor := interceptor.NewServiceTokenInterceptor(tool.ctxdata, serviceTokenInterceptorConfig)

	return aspectCollection{
		authMiddleware:          authMiddleware,
		serviceTokenInterceptor: serviceTokenInterceptor,
	}
}
