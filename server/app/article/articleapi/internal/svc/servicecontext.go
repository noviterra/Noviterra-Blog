// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"Noviterra/server/app/article/articleapi/internal/config"
	"Noviterra/server/app/article/articlerpc/articleservice"
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/app/internal/middleware"
	"Noviterra/server/utils/common"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/jwtauth"
	"Noviterra/server/utils/repo"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                  config.Config
	ArticleModel            articleservice.ArticleService
	AuthMiddleware          *middleware.AuthMiddleware
	CommonUtils             *common.CommonUtils
	ServiceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func NewServiceContext(c config.Config) *ServiceContext {
	// article rpc服务
	articleService := articleservice.NewArticleService(zrpc.MustNewClient(c.ArticleRpc))

	// jwt插件
	jwtConfig := jwtauth.AuthenticationConfig{
		AccessSecret:  c.JWT.AccessSecret,
		AccessExpire:  c.JWT.AccessExpire,
		RefreshSecret: c.JWT.RefreshSecret,
		RefreshExpire: c.JWT.RefreshExpire,
	}
	authentication := jwtauth.NewAuthentication(jwtConfig)

	// 普通工具插件
	commonUtils := common.NewCommonUtils()

	// 用户仓库实例 用于鉴权中获取用户的信息
	identityRepo := repo.NewIdentityRespository(repo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	}, "user")

	// 认证中间件
	m := middleware.NewAuthMiddleware(authentication, identityRepo)

	// ctxdata
	ctxData := ctxdata.NewCtxData()

	// 拦截器
	serviceTokenInterceptor := interceptor.NewServiceTokenInterceptor(ctxData, interceptor.ServiceTokenConfig{
		SecretToken: c.Interceptor.ServiceToken,
		SkipToken:   c.Interceptor.SkipToken,
	})

	return &ServiceContext{
		Config:                  c,
		ArticleModel:            articleService,
		AuthMiddleware:          m,
		CommonUtils:             commonUtils,
		ServiceTokenInterceptor: serviceTokenInterceptor,
	}
}
