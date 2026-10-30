package main

import (
	"flag"
	"fmt"

	"Noviterra/server/app/notice/noticerpc/internal/config"
	"Noviterra/server/app/notice/noticerpc/internal/server"
	"Noviterra/server/app/notice/noticerpc/internal/svc"
	"Noviterra/server/app/notice/noticerpc/types/noticerpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "server/app/notice/noticerpc/etc/noticerpc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		noticerpc.RegisterNoticeServiceServer(grpcServer, server.NewNoticeServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(ctx.ServiceTokenInterceptor.ServiceTokenInterceptor(c.Interceptor.ServiceToken, c.Interceptor.SkipToken))

	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
