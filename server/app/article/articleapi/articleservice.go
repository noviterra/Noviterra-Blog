// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"flag"
	"fmt"

	"Noviterra/server/app/article/articleapi/internal/config"
	"Noviterra/server/app/article/articleapi/internal/handler"
	"Noviterra/server/app/article/articleapi/internal/svc"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/articleservice.yaml", "the config file")

func main() {
	err := godotenv.Load("/.env")
	if err != nil {
		panic(err)
	}

	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
