package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go.uber.org/zap"
	"miaosha-demo/common"
	"miaosha-demo/fronted/web/controllers"
	"miaosha-demo/repositories"
	"miaosha-demo/rpc"
	"miaosha-demo/services"
)

func main() {
	flagPort := flag.Int64("port", 8082, "server port")
	flag.Parse()
	port := *flagPort

	common.NewZapLogger()

	defer func() {
		//recover panic, only for unexpect exception
		r := recover()

		if r != nil {
			var rerr error
			switch e := r.(type) {
			case string:
				rerr = errors.New(e)
			case error:
				rerr = e
			default:
				rerr = errors.New(fmt.Sprintf("%v", e))
			}
			common.ZapError("recover error", rerr)
		}

		zap.L().Sync()
	} ()

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
	if err != nil {
		common.ZapError("new config consul fail", err)
		return
	}
	cache := common.NewFreeCacheClient(20)

	Consul, err := common.NewConsulClient(config, cache)
	if err != nil {
		common.ZapError("new consul client fail", err)
		return
	}

	//取consul上redis service的配置
	redisClusterClient, err := common.NewRedisClusterClient(Consul)
	if err != nil {
		common.ZapError("new redis cluster fail", err)
		return
	}

	mysqlPoolUser, err := common.NewMysqlPoolUser(Consul)
	if err != nil {
		common.ZapError("new mysql pool user fail", err)
		return
	}

	mysqlPoolProduct, err := common.NewMysqlPoolProduct(Consul)
	if err != nil {
		common.ZapError("new mysql pool product fail", err)
		return
	}

	rabbitmqPool, err := common.NewRabbitmqPool(Consul)
	if err != nil {
		common.ZapError("new rabbitmq pool fail", err)
		return
	}

	rpcUser, err := user.NewRpcUser(Consul)
	if err != nil {
		common.ZapError("new rpc user fail", err)
		return
	}

	userRepository := repositories.NewUserRepository(mysqlPoolUser)
	userService := services.NewUserService(userRepository)

	//首页
	index0Party := app.Party("/")
	index0 := mvc.New(index0Party)
	index0.Handle(new(controllers.IndexController))

	indexParty := app.Party("/index")
	index := mvc.New(indexParty)
	index.Handle(new(controllers.IndexController))

	productRepository := repositories.NewProductRepository(mysqlPoolProduct)
	productService := services.NewProductService(productRepository)
	productParty := app.Party("/product")
	product := mvc.New(productParty)
	product.Register(ctx, productService, redisClusterClient, rabbitmqPool)
	product.Handle(new(controllers.ProductController))

	userParty := app.Party("/user")
	user := mvc.New(userParty)
	user.Register(ctx, userService, rpcUser)
	user.Handle(new(controllers.UserController))

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	app.Run(
		iris.Addr(addr),
		//iris.WithoutVersionChecker,
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)
}

