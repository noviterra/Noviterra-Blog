package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Cache struct {
		Addr     string
		Password string
		DB       int
	}
	Repo struct {
		UserName string
		Password string
		Host     string
		Port     int
		DBName   string
	}
	JWT struct {
		AccessSecret  string
		AccessExpire  int64
		RefreshSecret string
		RefreshExpire int64
	}
	Interceptor struct {
		ServiceToken string
		SkipToken    string
	}
}
