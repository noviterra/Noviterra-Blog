package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	Smtp struct {
		Host     string
		Port     int
		User     string
		Password string
		From     string
	}

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
}
