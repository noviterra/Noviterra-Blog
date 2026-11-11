package svc

import (
	"Noviterra/server/app/user/userrpc/internal/config"
	"Noviterra/server/utils/bcryptinfo"
	"Noviterra/server/utils/cache"
	"Noviterra/server/utils/common"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/email"
	"Noviterra/server/utils/jwtauth"
	"Noviterra/server/utils/repo"
	"Noviterra/server/utils/response"
	"Noviterra/server/utils/validate"
	"Noviterra/server/utils/zaplogger"
)

type ServiceContext struct {
	Config             config.Config
	Logger             *zaplogger.Logger
	EmailSender        *email.EmailSender
	CommonUtils        *common.CommonUtils
	Crypt              *bcryptinfo.Crypt
	UserResponse       *response.UserResponse
	IdentityValidate   *validate.IdentityValidate
	UserCache          *cache.UserCache
	UserRepository     *repo.UserRepository
	IdentityRepository *repo.IdentityRepository
	UserValidate       *validate.UserValidate
	ArticleValidate    *validate.ArticleValidate
	Ctxdata            *ctxdata.CtxData
	NoticeCache        *cache.NoticeCache
	Authentication     *jwtauth.Authentication
}

func NewServiceContext(c config.Config) *ServiceContext {
	tool := initTool(c)
	aspect := initAspect(c)
	store := initStore(c, tool)

	return &ServiceContext{
		Config: c,

		Logger:             tool.logger,
		EmailSender:        tool.emailSender,
		CommonUtils:        tool.commonUtils,
		Crypt:              tool.crypt,
		UserResponse:       tool.UserResponse,
		IdentityValidate:   tool.identityValidate,
		UserValidate:       tool.userValidate,
		Ctxdata:            tool.ctxdata,
		ArticleValidate:    tool.articleValidate,
		UserCache:          store.userCache,
		UserRepository:     store.userRepository,
		IdentityRepository: store.identityRepository,
		NoticeCache:        store.noticeCache,
		Authentication:     aspect.Authentication,
	}
}

// 工具类
type toolCollection struct {
	logger           *zaplogger.Logger
	emailSender      *email.EmailSender
	commonUtils      *common.CommonUtils
	crypt            *bcryptinfo.Crypt
	UserResponse     *response.UserResponse
	identityValidate *validate.IdentityValidate
	userValidate     *validate.UserValidate
	ctxdata          *ctxdata.CtxData
	articleValidate  *validate.ArticleValidate
}

func initTool(c config.Config) toolCollection {
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

	// 用户统一响应
	userResponse := response.NewUserResponse()

	// 校验器
	identityValidate := validate.NewIdentityValidate()
	userValidate := validate.NewUserValidate(identityValidate)
	articleValidate := validate.NewArticleValidate(identityValidate)

	cd := ctxdata.NewCtxData()

	return toolCollection{
		logger:           logger,
		emailSender:      emailSender,
		commonUtils:      commonUtils,
		crypt:            crypt,
		UserResponse:     userResponse,
		identityValidate: identityValidate,
		userValidate:     userValidate,
		ctxdata:          cd,
		articleValidate:  articleValidate,
	}
}

// 存储类
type storeCollection struct {
	userCache          *cache.UserCache
	userRepository     *repo.UserRepository
	identityRepository *repo.IdentityRepository
	noticeCache        *cache.NoticeCache
}

func initStore(c config.Config, tool toolCollection) storeCollection {
	// 数据库配置
	repoConfig := repo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	}
	// 缓存配置
	cacheConfig := cache.RedisConfig{
		Addr:     c.Cache.Addr,
		Password: c.Cache.Password,
		DB:       c.Cache.DB,
	}

	// 用户数据库 认证数据库
	userRepo := repo.NewUserRepository(repoConfig, "user")
	identityRepo := repo.NewIdentityRespository(repoConfig, "identity")

	// 用户缓存 认证缓存
	userCache := cache.NewUserCache(cacheConfig, tool.commonUtils)
	noticeCache := cache.NewNoticeCache(cacheConfig)

	return storeCollection{
		userCache:          userCache,
		userRepository:     userRepo,
		identityRepository: identityRepo,
		noticeCache:        noticeCache,
	}
}

// 切面类 请求拦截
type aspectCollection struct {
	Authentication *jwtauth.Authentication
}

func initAspect(c config.Config) aspectCollection {
	// JWT配置
	authenticationConfig := jwtauth.AuthenticationConfig{
		AccessSecret:  c.JWT.AccessSecret,
		AccessExpire:  c.JWT.AccessExpire,
		RefreshSecret: c.JWT.RefreshSecret,
		RefreshExpire: c.JWT.RefreshExpire,
	}
	authMiddleware := jwtauth.NewAuthentication(authenticationConfig)

	return aspectCollection{
		Authentication: authMiddleware,
	}
}
