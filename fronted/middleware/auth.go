package middleware

import (
	"github.com/kataras/iris/v12"
	"miaosha-demo/services"
)

func NeedLogin(ctx iris.Context) {
	if !services.CheckLoginCookie(ctx) {
		ctx.Redirect("/user/login")
		return
	}

	ctx.Next()
}