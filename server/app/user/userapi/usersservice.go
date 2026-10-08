// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"Noviterra/server/app/user/userapi/internal/config"
	"Noviterra/server/app/user/userapi/internal/handler"
	"Noviterra/server/app/user/userapi/internal/svc"
	"flag"
	"fmt"
	"os"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "server/app/user/userapi/etc/usersservice.yaml", "the config file")

func main() {
	//wd, _ := os.Getwd()
	//abs, _ := filepath.Abs("etc/usersservice.yaml")
	//
	//fmt.Println("当前工作目录:", wd)
	//fmt.Println("实际查找路径:", abs)

	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.Infof("MYSQL_USER=%q MYSQL_HOST=%q", os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_HOST"))

	server := rest.MustNewServer(c.RestConf, rest.WithCors("*"))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
