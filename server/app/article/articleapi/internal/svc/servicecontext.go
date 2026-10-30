// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"Noviterra/server/app/article/articleapi/internal/config"
	"Noviterra/server/app/article/articlerpc/articleservice"
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/utils/common"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/repo"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                  config.Config
	ArticleModel            articleservice.ArticleService
	CommonUtils             *common.CommonUtils
	CtxData                 *ctxdata.CtxData
	ServiceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func NewServiceContext(c config.Config) *ServiceContext {
	tool := initTool()
	rpc := initRPC(c)
	aspect := initAspect(c, tool)

	return &ServiceContext{
		Config:                  c,
		ArticleModel:            rpc.ArticleModel,
		ServiceTokenInterceptor: aspect.ServiceTokenInterceptor,
		CommonUtils:             tool.CommonUtils,
		CtxData:                 tool.ctxdata,
	}
}

// rpc服务类
type rpcCollection struct {
	ArticleModel articleservice.ArticleService
}

func initRPC(c config.Config) rpcCollection {
	// article rpc服务
	articleService := articleservice.NewArticleService(zrpc.MustNewClient(c.ArticleRpc))

	return rpcCollection{
		ArticleModel: articleService,
	}
}

// 工具类
type toolCollection struct {
	ctxdata     *ctxdata.CtxData
	CommonUtils *common.CommonUtils
}

func initTool() toolCollection {
	// 上下文存储操作
	cd := ctxdata.NewCtxData()

	// 基础工具包
	commonUtils := common.NewCommonUtils()

	return toolCollection{
		ctxdata:     cd,
		CommonUtils: commonUtils,
	}
}

// 存储类
type storeCollection struct {
	identityRepository *repo.IdentityRepository
}

func initStore(c config.Config) storeCollection {
	// 存储配置
	repositoryConfig := repo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	}
	// 构建存储实例
	identityRepo := repo.NewIdentityRespository(repositoryConfig, "identity")

	return storeCollection{
		identityRepository: identityRepo,
	}
}

// 请求拦截类
type aspectCollection struct {
	ServiceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func initAspect(c config.Config, tool toolCollection) aspectCollection {
	// JWT配置

	// 拦截器配置
	interceptorConfig := interceptor.ServiceTokenConfig{
		SecretToken: c.Interceptor.ServiceToken,
		SkipToken:   c.Interceptor.SkipToken,
	}
	it := interceptor.NewServiceTokenInterceptor(tool.ctxdata, interceptorConfig)

	return aspectCollection{
		ServiceTokenInterceptor: it,
	}
}
