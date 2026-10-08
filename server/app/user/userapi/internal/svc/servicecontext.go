// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"Noviterra/server/app/internal/middleware"
	"Noviterra/server/app/user/userapi/internal/config"
	"Noviterra/server/app/user/userrpc/userservice"
	gormrepo "Noviterra/server/utils/gormRepo"
	jwtauth "Noviterra/server/utils/jwtAuth"
	zaplogger "Noviterra/server/utils/zapLogger"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	Logger        *zaplogger.Logger
	UserModel     userservice.UserService
	Authorization *jwtauth.Authentication
	Middleware    *middleware.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 依赖注入
	logger := zaplogger.NewLogger()
	userService := userservice.NewUserService(zrpc.MustNewClient(c.UserRpc)) // user-rpc服务连接

	// JWT
	authenticationConfig := jwtauth.AuthenticationConfig{
		AccessSecret:  c.JWT.AccessSecret,
		AccessExpire:  c.JWT.AccessExpire,
		RefreshSecret: c.JWT.RefreshSecret,
		RefreshExpire: c.JWT.RefreshExpire,
	}
	authentication := jwtauth.NewAuthentication(authenticationConfig)

	// 获取仓库操作实例 用于注入中间件
	userRepo := gormrepo.NewRepository(gormrepo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	})

	// 中间件
	m := middleware.NewMiddleware(authentication, userRepo)

	return &ServiceContext{
		Config:        c,
		Logger:        logger,
		UserModel:     userService,
		Authorization: authentication,
		Middleware:    m,
	}
}
