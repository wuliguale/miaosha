package main

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"miaosha-demo/common"
	"miaosha-demo/fronted/web/controllers"
	"miaosha-demo/repositories"
	"miaosha-demo/rpc"
	"miaosha-demo/services"
)

func main() {
	app := iris.New()
	app.Logger().SetLevel("debug")

	template := iris.HTML("./web/views", ".html").Reload(true)
	app.RegisterView(template)
	app.HandleDir("/assets", "./../backend/web/assets")

	//错误处理，如404
	app.OnAnyErrorCode(func(ctx iris.Context) {
		ctx.ViewData("msg", ctx.GetStatusCode())
		ctx.ViewData("url", "/")
		ctx.View("error.html")
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := common.NewConfigConsul()
	fmt.Println("new config,", err)
	cache := common.NewFreeCacheClient(20)

	Consul, err := common.NewConsulClient(config, cache)
	fmt.Println("new consul", err)

	//一直watch consul上的service
	//todo watch中修改全局的pool
	serviceNameList := Consul.Config.GetServiceNameList()
	for _, serviceName := range serviceNameList {
		go Consul.WatchServiceByName(serviceName)
	}

	//取consul上redis service的配置
	redisClusterClient, err := common.NewRedisClusterClient(Consul)
	mysqlPool, err := common.NewMysqlPool(Consul)
	rabbitmqPool, err := common.NewRabbitmqPool(Consul)
	rpcUser, err := user.NewRpcUser(Consul)

	userRepository := repositories.NewUserRepository(mysqlPool)
	userService := services.NewUserService(userRepository)

	//首页
	index0Party := app.Party("/")
	index0 := mvc.New(index0Party)
	index0.Handle(new(controllers.IndexController))

	indexParty := app.Party("/index")
	index := mvc.New(indexParty)
	index.Handle(new(controllers.IndexController))

	productRepository := repositories.NewProductRepository(mysqlPool)
	productService := services.NewProductService(productRepository)
	productParty := app.Party("/product")
	product := mvc.New(productParty)
	product.Register(ctx, productService, redisClusterClient, rabbitmqPool)
	product.Handle(new(controllers.ProductController))

	userParty := app.Party("/user")
	user := mvc.New(userParty)
	user.Register(ctx, userService, rpcUser)
	user.Handle(new(controllers.UserController))

	app.Run(
		iris.Addr("0.0.0.0:8082"),
		//iris.WithoutVersionChecker,
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)

}
