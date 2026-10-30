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
	Cache struct {
		Addr     string
		Password string
		DB       int
	}
	Smtp struct {
		Host     string
		Port     int
		User     string
		Password string
		From     string
	}
	Interceptor struct {
		ServiceToken string
		SkipToken    string
	}
}
