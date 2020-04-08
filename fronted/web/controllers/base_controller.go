package controllers

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/streadway/amqp"
	"miaosha-demo/common"
)


const RETURN_SUCCESS  = 0
const RETURN_FAIL  = 1

var Consul *common.ConsulClient
//初始化redis连接
var RedisClusterClient *redis.ClusterClient
//初始化rabbitmq连接
var RabbitMqConn *amqp.Connection

func init() {
	RabbitMqConn, _ = common.NewRabbitMqConn()

	config, err := common.NewConfigConsul()
	fmt.Println("new config,", err)

	cache := common.NewFreeCacheClient(20)

	Consul, err = common.NewConsulClient(config, cache)
	fmt.Println("new consul", err)

	//一直watch consul上的service
	serviceNameList := Consul.Config.GetServiceNameList()
	for _, serviceName := range serviceNameList {
		go Consul.WatchServiceByName(serviceName)
	}

	//取consul上redis service的配置
	RedisClusterClient, err = common.NewRedisClusterClient(Consul)
}



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

//接口成功，返回数据
func ReturnJsonSuccess(ctx iris.Context, data map[string]interface{}) {
	ctx.JSON(iris.Map{
		"code":  RETURN_SUCCESS,
		"data" : data,
	})
}


//接口失败，返回错误信息
func ReturnJsonFail(ctx iris.Context, msg string) {
	ctx.JSON(iris.Map {
		"code" : RETURN_FAIL,
		"msg" : msg,
	})
}

