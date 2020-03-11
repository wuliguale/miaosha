package controllers

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"github.com/streadway/amqp"
	"miaosha-demo/common"
	"miaosha-demo/services"
	"strconv"
)

type ProductController struct {
	Ctx            iris.Context
	ProductService services.IProductService
	OrderService   services.IOrderService
	Session        *sessions.Session
}

//商品列表
func (p *ProductController) GetAll() mvc.View{
	page := p.Ctx.URLParamInt64Default("page", 1)
	pageSize := p.Ctx.URLParamInt64Default("pageSize", 10)

	productList, err := p.ProductService.GetProductAll(page, pageSize, "id", true)
	if err != nil {
		return errorReturnView(p.Ctx, err.Error(), "/", 500)
	}

	productListMap := common.StructPtrArray2MapArray(productList)

	for k,v := range productListMap {
		v["UrlDetail"] = "/product/one?id=" + fmt.Sprintf("%v", v["Id"])
		productListMap[k] = v
	}

	return mvc.View{
		Name: "product/list",
		Data: iris.Map{
			"productList": productListMap,
		},
	}
}


//一个商品的详情
func (p *ProductController) GetOne() mvc.View{
	idString := p.Ctx.URLParam("id")
	id, err := strconv.ParseInt(idString, 10, 16)
	if err != nil {
		return errorReturnView(p.Ctx, err.Error(), "/product/all", 500)
	}

	product, err := p.ProductService.GetProductByID(uint64(id))
	if err != nil {
		return errorReturnView(p.Ctx, err.Error(), "/product/all", 500)
	}

	return mvc.View{
		Name: "product/detail",
		Data: iris.Map{
			"product": common.StructPtr2Map(product),
		},
	}
}

//秒杀接口
func (p *ProductController) GetOrder() {
	pid := p.Ctx.PostValueInt64Default("pid", 0)
	uid, err := services.GetUidFromCookie(p.Ctx)
	if pid == 0 || uid == 0 || err != nil {
		ReturnJsonFail(p.Ctx, "参数错误")
		return
	}

	//从连接池获取连接
	redisConn := RedisPool.Get()
	defer redisConn.Close()

	//秒杀是否已结束
	pidOverKey := "pid_over_" + strconv.Itoa(int(pid))
	isOver, err := redis.Bool(redisConn.Do("get", pidOverKey))
	if err != nil {
		ReturnJsonFail(p.Ctx, "redis判断秒杀结束错误")
		return
	}
	if isOver {
		ReturnJsonFail(p.Ctx, "秒杀已结束")
	}

	//是否重复购买
	isRepeatKey := "pid_" + strconv.Itoa(int(pid))
	isRepeat, err := redis.Bool(redisConn.Do("getbit", isRepeatKey, uid))
	if err != nil {
		ReturnJsonFail(p.Ctx, "redis检查是否重复购买错误")
		return
	}
	if isRepeat {
		ReturnJsonFail(p.Ctx, "不能重复购买")
		return
	}

	//检查库存
	numKey := "pid_num" + strconv.Itoa(int(pid))
	num, err := redis.Int(redisConn.Do("decr", numKey))
	if err != nil {
		ReturnJsonFail(p.Ctx, "redis检查库存错误")
		return
	}
	if num <= 0 {
		ReturnJsonFail(p.Ctx, "已无库存")
	}

	ch, err := RabbitMqConn.Channel()
	if err != nil {
		ReturnJsonFail(p.Ctx, "rabbitmq channel错误")
		return
	}
	defer ch.Close()

	body := strconv.Itoa(int(pid)) + "_" + strconv.Itoa(int(uid))

	err = ch.Publish(
		"miaosha_demo",
		"miaosha_demo_routing_key",
		false,
		false,
		amqp.Publishing{
			ContentType:"text/plain",
			Body:[]byte(body),
		})

	if err != nil {
		ReturnJsonFail(p.Ctx, "mq publish 错误")
		return
	}

	data := map[string]interface{}{}
	ReturnJsonSuccess(p.Ctx, data)
	return
}


