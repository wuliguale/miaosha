package main

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"miaosha-demo/backend/web/controllers"
	"miaosha-demo/common"
	"miaosha-demo/repositories"
	"miaosha-demo/services"
)

func main() {
	app := iris.New()
	app.Logger().SetLevel("debug")

	template := iris.HTML("./backend/web/views", ".html").Reload(true)
	app.RegisterView(template)

	app.HandleDir("/assets", "./backend/web/assets")

	//错误处理，如404
	app.OnAnyErrorCode(func(ctx iris.Context) {
		ctx.ViewData("msg", ctx.GetStatusCode())
		ctx.ViewData("url", "/")
		ctx.View("error.html")
	})

	//连接数据库
	db, err := common.NewMysqlConn()
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//首页
	index0Party := app.Party("/")
	index0 := mvc.New(index0Party)
	index0.Handle(new(controllers.IndexController))

	indexParty := app.Party("/index")
	index := mvc.New(indexParty)
	index.Handle(new(controllers.IndexController))

	productRepository := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepository)
	productParty := app.Party("/product")
	product := mvc.New(productParty)
	product.Register(ctx, productService)
	product.Handle(new(controllers.ProductController))

	orderRepository := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepository)
	orderParty := app.Party("/order")
	order := mvc.New(orderParty)
	order.Register(ctx, orderService)
	order.Handle(new(controllers.OrderController))

	app.Run(
		iris.Addr("localhost:8080"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)
}



