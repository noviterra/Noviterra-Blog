package main

import (
	"flag"
	"fmt"

	"Noviterra/server/app/site/siterpc/internal/config"
	"Noviterra/server/app/site/siterpc/internal/server"
	"Noviterra/server/app/site/siterpc/internal/svc"
	"Noviterra/server/app/site/siterpc/types/siterpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "server/app/site/siterpc/etc/siterpc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		siterpc.RegisterSiteServiceServer(grpcServer, server.NewSiteServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(ctx.ServiceTokenInterceptor.
		ServiceTokenInterceptor(c.Interceptor.ServiceToken, c.Interceptor.SkipToken))
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
