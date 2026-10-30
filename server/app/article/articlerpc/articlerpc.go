package main

import (
	"flag"
	"fmt"

	"Noviterra/server/app/article/articlerpc/internal/config"
	"Noviterra/server/app/article/articlerpc/internal/server"
	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "server/app/article/articlerpc/etc/articlerpc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		articlerpc.RegisterArticleServiceServer(grpcServer, server.NewArticleServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	},
	)

	// 在此注册拦截器
	s.AddUnaryInterceptors(ctx.ServiceTokenInterceptor.ServiceTokenInterceptor(c.Interceptor.ServiceToken, c.Interceptor.SkipToken))

	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
