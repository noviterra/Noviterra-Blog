package main

import (
	"Noviterra/server/app/comment/commentrpc/internal/config"
	"Noviterra/server/app/comment/commentrpc/internal/server"
	"Noviterra/server/app/comment/commentrpc/internal/svc"
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "server/app/comment/commentrpc/etc/commentrpc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		commentrpc.RegisterCommentServiceServer(grpcServer, server.NewCommentServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	// 拦截器
	s.AddUnaryInterceptors(ctx.ServiceTokenInterceptor.ServiceTokenInterceptor(c.Interceptor.ServiceToken, c.Interceptor.SkipToken))
	defer s.Stop()

	// 监听定时任务
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()

	// 优雅关闭 关闭时发送信号进行最后一次flush
	<-quit
	ctx.CommentWoker.Stop()
}
