// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"flag"
	"fmt"

	"Noviterra/server/app/user/userapi/internal/config"
	"Noviterra/server/app/user/userapi/internal/handler"
	"Noviterra/server/app/user/userapi/internal/svc"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/usersservice.yaml", "the config file")

func main() {
	err := godotenv.Load("/.env")
	if err != nil {
		panic(err)
	}

	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(c.RestConf, rest.WithCors("*"))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
