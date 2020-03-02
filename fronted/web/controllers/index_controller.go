package controllers

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"miaosha-demo/services"
)

type IndexController struct {
	Ctx iris.Context
}


func (i *IndexController) Get() mvc.View {
	data := []map[string]string{
		{"url" : "/product/all", "desc" : "商品列表"},
	}

	isLogin := services.CheckLoginCookie(i.Ctx)

	return mvc.View {
		Name : "index/index",
		Data : iris.Map {
			"data" : data,
			"urlLogin" : "/user/login",
			"urlLogout" : "/user/logout",
			"isLogin" : isLogin,
		},
	}
}
