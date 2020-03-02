package controllers

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"miaosha-demo/common"
	"miaosha-demo/services"
	"net/url"
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

	data := map[string]interface{}{"a":"aaa", "b":"bbb"}
	ReturnJsonSuccess(p.Ctx, data)
	url.Values{}.Encode()
}

