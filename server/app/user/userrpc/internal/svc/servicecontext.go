package svc

import (
	"Noviterra/server/app/user/userrpc/internal/config"
	bcryptinfo "Noviterra/server/utils/bcryptInfo"
	"Noviterra/server/utils/common"
	"Noviterra/server/utils/email"
	gormrepo "Noviterra/server/utils/gormRepo"
	jwtauth "Noviterra/server/utils/jwtAuth"
	rediscache "Noviterra/server/utils/redisCache"
	zaplogger "Noviterra/server/utils/zapLogger"
)

type ServiceContext struct {
	Config         config.Config
	Logger         *zaplogger.Logger
	EmailSender    *email.EmailSender
	CommonUtils    *common.CommonUtils
	RDB            *rediscache.RedisCache
	DB             *gormrepo.Repository
	Crypt          *bcryptinfo.Crypt
	Authentication *jwtauth.Authentication
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
	redisConfig := rediscache.RedisConfig{
		Addr:     c.Cache.Addr,
		Password: c.Cache.Password,
		DB:       c.Cache.DB,
	}
	rdb := rediscache.NewRedisCache(redisConfig)

	// JWT
	authenticationConfig := jwtauth.AuthenticationConfig{
		AccessSecret:  c.JWT.AccessSecret,
		AccessExpire:  c.JWT.AccessExpire,
		RefreshSecret: c.JWT.RefreshSecret,
		RefreshExpire: c.JWT.RefreshExpire,
	}
	authentication := jwtauth.NewAuthentication(authenticationConfig)

	// 数据库
	gormConfig := gormrepo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	}
	db := gormrepo.NewRepository(gormConfig)

	return &ServiceContext{
		Config:         c,
		Logger:         logger,
		EmailSender:    emailSender,
		CommonUtils:    commonUtils,
		RDB:            rdb,
		DB:             db,
		Crypt:          crypt,
		Authentication: authentication,
	}
}
