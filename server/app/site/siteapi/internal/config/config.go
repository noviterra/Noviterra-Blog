// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	SiteRPC   zrpc.RpcClientConf
	NoticeRPC zrpc.RpcClientConf
	Repo      struct {
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
