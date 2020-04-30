package controllers

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
)

type IndexController struct {
	Ctx            iris.Context
}


func (i *IndexController) Get() mvc.View {
	data := []map[string]string{
		{"url" : "/product/all", "desc" : "商品列表"},
	}

	return mvc.View {
		Name : "index/index",
		Data : iris.Map {
			"data" : data,
		},
	}
}
