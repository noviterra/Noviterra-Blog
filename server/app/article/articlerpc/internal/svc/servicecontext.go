package svc

import (
	"Noviterra/server/app/article/articlerpc/internal/config"
	"Noviterra/server/app/internal/interceptor"
	"Noviterra/server/utils/common"
	"Noviterra/server/utils/ctxdata"
	gormrepo "Noviterra/server/utils/gormRepo"
	"Noviterra/server/utils/snowflake"
)

type ServiceContext struct {
	Config      config.Config
	SnowFlake   *snowflake.SnowFlake
	CtxData     *ctxdata.CtxData
	Interceptor *interceptor.Interceptor
	Repo        *gormrepo.Repository
	CommonUtils *common.CommonUtils
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 雪花ID生成器
	sf, err := snowflake.NewSnowFlake(1)
	if err != nil {
		panic(err)
	}

	// 工具插件
	cu := common.NewCommonUtils()

	// 上下文存储器 用于rpc层的身份校验
	cd := ctxdata.NewCtxData()

	// 拦截器
	itp := interceptor.NewInterceptor(cd)

	// 仓库操作实例
	repoConfig := gormrepo.RepoConfig{
		UserName: c.Repo.UserName,
		Password: c.Repo.Password,
		Host:     c.Repo.Host,
		Port:     c.Repo.Port,
		DBName:   c.Repo.DBName,
	}
	repo := gormrepo.NewRepository(repoConfig)

	return &ServiceContext{
		Config:      c,
		SnowFlake:   sf,
		CtxData:     cd,
		Interceptor: itp,
		Repo:        repo,
		CommonUtils: cu,
	}
}
