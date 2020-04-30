package controllers

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
)

//显示错误
func errorReturnView(ctx iris.Context, msg, url string, statusCode int) mvc.View {
	ctx.StatusCode(statusCode)

	return mvc.View{
		Name: "error.html",
		Data: iris.Map{
			"msg" : msg,
			"url" : url,
		},
	}
}

//提示信息并跳转
func messageThenRedirect(msg, url string) mvc.View {
	return mvc.View {
		Name : "message.html",
		Data : iris.Map {
			"msg" : msg,
			"url" : url,
		},
	}
}
