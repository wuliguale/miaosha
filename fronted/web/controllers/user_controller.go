package controllers

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"miaosha-demo/fronted/middleware"
	userRpc "miaosha-demo/rpc"
	"miaosha-demo/services"
	"time"
)

type UserController struct {
	Ctx     iris.Context
	UserService services.IUserService
}


func (u *UserController) BeforeActivation(b mvc.BeforeActivation) {
	// 单独指定地址和方法，中间件
	b.Handle("GET", "/order", "GetOrder", middleware.NeedLogin)
}


//注册页
func (u *UserController) GetRegister() mvc.View {
	return mvc.View{
		Name: "user/register",
		Data : iris.Map {
			"urlPost" : "/user/register",
		},
	}
}

//注册提交
func (u *UserController) PostRegister() mvc.View {
	userName := u.Ctx.FormValue("username")
	nickName := u.Ctx.FormValue("nickname")
	password := u.Ctx.FormValue("password")

	_, err := userRpc.RpcUserServiceReg(userName, nickName, password)
	if err != nil {
		return errorReturnView(u.Ctx, err.Error(), "/", 500)
	}

	return messageThenRedirect("注册成功", "/user/login")
}

//登录页
func (u *UserController) GetLogin() mvc.View {
	return mvc.View{
		Name: "user/login",
		Data : iris.Map {
			"urlSubmit" : "/user/login",
		},
	}
}

//登录提交
func (u *UserController) PostLogin() mvc.View {
	userName := u.Ctx.FormValue("username")
	password := u.Ctx.FormValue("password")

	user, err := userRpc.RpcUserServiceLogin(userName, password)
	fmt.Println(user, err)
	if err != nil {
		return messageThenRedirect("密码错误", "/user/login")
	}

	duration := time.Duration(24 * time.Hour)
	services.SetLoginCookie(u.Ctx, user, duration)

	return messageThenRedirect("登录成功", "/")
}


//退出登录
func (u *UserController) GetLogout() {
	u.Ctx.RemoveCookie("sign")
	u.Ctx.Redirect("/")
}


//用户订单页
func (u *UserController) GetOrder() mvc.View {
	data := 1

	return mvc.View {
		Name : "user/order",
		Data : iris.Map {
			"data" : data,
		},
	}
}

