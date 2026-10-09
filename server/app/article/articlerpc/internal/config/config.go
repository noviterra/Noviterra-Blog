package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	Interceptor struct {
		ServiceToken string
		SkipToken    string
	}

	Repo struct {
		UserName string
		Password string
		Host     string
		Port     int
		DBName   string
	}
}
