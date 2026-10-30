// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"Noviterra/server/app/identity/identityapi/internal/config"
	"Noviterra/server/app/identity/identityrpc/identityservice"
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/utils/ctxdata"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                  config.Config
	IdentityModel           identityservice.IdentityService
	ServiceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func NewServiceContext(c config.Config) *ServiceContext {
	rpc := initRPC(c)
	tool := initTool()
	aspect := initAspect(c, tool)

	return &ServiceContext{
		Config:                  c,
		IdentityModel:           rpc.identityModel,
		ServiceTokenInterceptor: aspect.serviceTokenInterceptor,
	}
}

type rpcCollection struct {
	identityModel identityservice.IdentityService
}

func initRPC(c config.Config) rpcCollection {
	identityService := identityservice.NewIdentityService(zrpc.MustNewClient(c.IdentityRPC))

	return rpcCollection{
		identityModel: identityService,
	}
}

type toolCollection struct {
	ctxdata *ctxdata.CtxData
}

func initTool() toolCollection {
	// 上下文存储操作
	cd := ctxdata.NewCtxData()

	return toolCollection{
		ctxdata: cd,
	}
}

type aspectCollection struct {
	serviceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func initAspect(c config.Config, tool toolCollection) aspectCollection {
	// 非法服务拦截器配置
	serviceTokenConfig := interceptor.ServiceTokenConfig{
		SecretToken: c.Interceptor.ServiceToken,
		SkipToken:   c.Interceptor.SkipToken,
	}

	// 拦截器实例
	serviceTokenInterceptor := interceptor.NewServiceTokenInterceptor(tool.ctxdata, serviceTokenConfig)

	return aspectCollection{
		serviceTokenInterceptor: serviceTokenInterceptor,
	}
}
