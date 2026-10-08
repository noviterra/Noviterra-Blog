// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"Noviterra/server/app/article/articleapi/internal/config"
	"Noviterra/server/app/article/articlerpc/articleservice"
	"Noviterra/server/app/internal/middleware"
	"Noviterra/server/utils/common"
	gormrepo "Noviterra/server/utils/gormRepo"
	jwtauth "Noviterra/server/utils/jwtAuth"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	ArticleModel articleservice.ArticleService
	Middleware   *middleware.Middleware
	CommonUtils  *common.CommonUtils
}

func NewServiceContext(c config.Config) *ServiceContext {
	// article rpc服务
	articleService := articleservice.NewArticleService(zrpc.MustNewClient(c.RpcClientConf))

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
	userRepo := gormrepo.NewRepository(gormrepo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	})

	// 集成中间件
	m := middleware.NewMiddleware(authentication, userRepo)

	return &ServiceContext{
		Config:       c,
		ArticleModel: articleService,
		Middleware:   m,
		CommonUtils:  commonUtils,
	}
}
