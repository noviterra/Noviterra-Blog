package svc

import (
	"Noviterra/server/app/user/userrpc/internal/config"
	"Noviterra/server/utils/bcryptinfo"
	"Noviterra/server/utils/cache"
	"Noviterra/server/utils/common"
	"Noviterra/server/utils/email"
	"Noviterra/server/utils/jwtauth"
	"Noviterra/server/utils/repo"
	"Noviterra/server/utils/zaplogger"
)

type ServiceContext struct {
	Config             config.Config
	Logger             *zaplogger.Logger
	EmailSender        *email.EmailSender
	CommonUtils        *common.CommonUtils
	UserCache          *cache.UserCache
	UserRepository     *repo.UserRepository
	Crypt              *bcryptinfo.Crypt
	Authentication     *jwtauth.Authentication
	IdentityRepository *repo.IdentityRepository
	IdentityCache      *cache.IdentityCache
}

func NewServiceContext(c config.Config) *ServiceContext {
	// zap日志
	logger := zaplogger.NewLogger()

	// 邮件发送器
	emailConfig := email.EmailConfig{
		Host:     c.Smtp.Host,
		Port:     c.Smtp.Port,
		Password: c.Smtp.Password,
		User:     c.Smtp.User,
		From:     c.Smtp.From,
	}
	emailSender := email.NewEmailSender(emailConfig)

	// 普通工具包 如随机数生产等
	commonUtils := common.NewCommonUtils()

	// 加密工具
	crypt := bcryptinfo.NewCrypt()

	// Redis缓存
	redisConfig := cache.RedisConfig{
		Addr:     c.Cache.Addr,
		Password: c.Cache.Password,
		DB:       c.Cache.DB,
	}
	userCache := cache.NewUserCache(redisConfig, commonUtils)
	identityCache := cache.NewIdentityCache(redisConfig)

	// JWT
	authenticationConfig := jwtauth.AuthenticationConfig{
		AccessSecret:  c.JWT.AccessSecret,
		AccessExpire:  c.JWT.AccessExpire,
		RefreshSecret: c.JWT.RefreshSecret,
		RefreshExpire: c.JWT.RefreshExpire,
	}
	authentication := jwtauth.NewAuthentication(authenticationConfig)

	// 数据库
	gormConfig := repo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	}
	userRepo := repo.NewUserRepository(gormConfig, "user")

	identityRepo := repo.NewIdentityRespository(gormConfig, "identity")

	return &ServiceContext{
		Config:             c,
		Logger:             logger,
		EmailSender:        emailSender,
		CommonUtils:        commonUtils,
		UserCache:          userCache,
		UserRepository:     userRepo,
		Crypt:              crypt,
		Authentication:     authentication,
		IdentityRepository: identityRepo,
		IdentityCache:      identityCache,
	}
}
