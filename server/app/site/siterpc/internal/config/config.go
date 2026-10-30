package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Repo struct {
		UserName string
		Password string
		Host     string
		Port     int
		DBName   string
	}
	Interceptor struct {
		ServiceToken string
		SkipToken    string
	}
}
