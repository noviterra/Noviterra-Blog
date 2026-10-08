package main

import (
	"flag"
	"fmt"

	"Noviterra/server/app/article/articlerpc/internal/config"
	"Noviterra/server/app/article/articlerpc/internal/server"
	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/articlerpc.yaml", "the config file")

func main() {
	// 加载.env到系统变量
	err := godotenv.Load("/.env")
	if err != nil {
		panic(err)
	}
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv()) // 加载环境变量
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		articlerpc.RegisterArticleServiceServer(grpcServer, server.NewArticleServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
