package main

import (
	"flag"
	"fmt"

	"Noviterra/server/app/identity/identityrpc/internal/config"
	"Noviterra/server/app/identity/identityrpc/internal/server"
	"Noviterra/server/app/identity/identityrpc/internal/svc"
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "server/app/identity/identityrpc/etc/identityrpc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		identityrpc.RegisterIdentityServiceServer(grpcServer, server.NewIdentityServiceServer(ctx))

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
