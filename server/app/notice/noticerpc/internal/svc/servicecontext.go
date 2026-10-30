package svc

import (
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/app/notice/noticerpc/internal/config"
	"Noviterra/server/utils/cache"
	"Noviterra/server/utils/common"
	"Noviterra/server/utils/ctxdata"
	"Noviterra/server/utils/email"
	"Noviterra/server/utils/repo"
	"Noviterra/server/utils/response"
	"Noviterra/server/utils/validate"
)

type ServiceContext struct {
	Config                  config.Config
	NoticeResponse          *response.NoticeResponse
	EmailSender             *email.EmailSender
	CommonUtils             *common.CommonUtils
	CtxData                 *ctxdata.CtxData
	NoticeValidate          *validate.NoticeValidate
	IdentityRepository      *repo.IdentityRepository
	NoticeCache             *cache.NoticeCache
	NoticeRepository        *repo.NoticeRepository
	ServiceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func NewServiceContext(c config.Config) *ServiceContext {
	tool := initTool(c)
	store := initStore(c)
	aspect := initAspect(c, tool)

	return &ServiceContext{
		Config:                  c,
		NoticeResponse:          tool.noticeResponse,
		EmailSender:             tool.emailSender,
		CommonUtils:             tool.commonUtils,
		CtxData:                 tool.ctxdata,
		NoticeValidate:          tool.noticeValidate,
		IdentityRepository:      store.identityRepository,
		NoticeCache:             store.noticeCache,
		NoticeRepository:        store.noticeRepository,
		ServiceTokenInterceptor: aspect.serviceTokenInterceptor,
	}
}

type toolCollection struct {
	noticeResponse *response.NoticeResponse
	emailSender    *email.EmailSender
	commonUtils    *common.CommonUtils
	ctxdata        *ctxdata.CtxData
	noticeValidate *validate.NoticeValidate
}

func initTool(c config.Config) toolCollection {
	emailSenderConfig := email.EmailConfig{
		Host:     c.Smtp.Host,
		Port:     c.Smtp.Port,
		User:     c.Smtp.User,
		Password: c.Smtp.Password,
		From:     c.Smtp.From,
	}

	noticeResponse := response.NewNoticeResponse()
	emailSender := email.NewEmailSender(emailSenderConfig)
	commonUtils := common.NewCommonUtils()
	cd := ctxdata.NewCtxData()
	noticeValidate := validate.NewNoticeValidate()

	return toolCollection{
		noticeResponse: noticeResponse,
		emailSender:    emailSender,
		commonUtils:    commonUtils,
		ctxdata:        cd,
		noticeValidate: noticeValidate,
	}
}

type storeCollection struct {
	identityRepository *repo.IdentityRepository
	noticeCache        *cache.NoticeCache
	noticeRepository   *repo.NoticeRepository
}

func initStore(c config.Config) storeCollection {
	repositoryConfig := repo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	}
	cacheConfig := cache.RedisConfig{
		Addr:     c.Cache.Addr,
		Password: c.Cache.Password,
		DB:       c.Cache.DB,
	}

	identityRepository := repo.NewIdentityRespository(repositoryConfig, "identity")
	noticeCache := cache.NewNoticeCache(cacheConfig)
	noticeRepository := repo.NewNoticeRepository(repositoryConfig, "notice")

	return storeCollection{
		identityRepository: identityRepository,
		noticeCache:        noticeCache,
		noticeRepository:   noticeRepository,
	}
}

type aspectCollection struct {
	serviceTokenInterceptor *interceptor.ServiceTokenInterceptor
}

func initAspect(c config.Config, tool toolCollection) aspectCollection {
	serviceTokenInterceptorConfig := interceptor.ServiceTokenConfig{
		SecretToken: c.Interceptor.ServiceToken,
		SkipToken:   c.Interceptor.SkipToken,
	}

	serviceTokenInterceptor := interceptor.NewServiceTokenInterceptor(tool.ctxdata, serviceTokenInterceptorConfig)

	return aspectCollection{
		serviceTokenInterceptor: serviceTokenInterceptor,
	}
}
