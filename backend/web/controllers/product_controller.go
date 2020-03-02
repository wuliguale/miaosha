package controllers

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"miaosha-demo/common"
	"miaosha-demo/datamodels"
	"miaosha-demo/services"
	"reflect"
	"strconv"
	"time"
)

type ProductController struct {
	Ctx            iris.Context
	ProductService services.IProductService
}

func (p *ProductController) BeforeActivation(b mvc.BeforeActivation) {
	// 单独指定地址和方法
	b.Handle("GET", "/update/{id:long}", "GetUpdate")
	b.Handle("GET", "/delete/{id:long}", "GetDelete")
}


//所有商品
func (p *ProductController) GetAll() mvc.View{
	page := p.Ctx.URLParamInt64Default("page", 1)
	pageSize := p.Ctx.URLParamInt64Default("pageSize", 10)

	productList, err := p.ProductService.GetProductAll(page, pageSize, "id", true)
	if err != nil {
		return errorReturnView(p.Ctx, err.Error(), "/product/all", 500)
	}

	productListMap := common.StructPtrArray2MapArray(productList)

	for k,v := range productListMap {
		v["UrlDetail"] = "/product/one?id=" + fmt.Sprintf("%v", v["Id"])
		v["UrlEdit"] = "/product/update/" + fmt.Sprintf("%v", v["Id"])
		v["UrlDelete"] = "/product/delete/" + fmt.Sprintf("%v", v["Id"])

		productListMap[k] = v
	}

	return mvc.View{
		Name: "product/list",
		Data: iris.Map{
			"productList": productListMap,
			"urlAdd" : "/product/add",
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


//商品添加页
func (p *ProductController) GetAdd() mvc.View {
	return mvc.View{
		Name : "product/add",
		Data : iris.Map{
			"Url" : "/product/add",
		},
	}
}


//商品添加提交
func (p *ProductController) PostAdd() mvc.View{
	name := p.Ctx.FormValue("name")
	num := p.Ctx.PostValueInt64Default("num", 0)
	image := p.Ctx.FormValue("image")
	url := p.Ctx.FormValue("url")

	product := &datamodels.Product{
		ProductName:name,
		Num:uint32(num),
		Image:image,
		Url:url,
		State:datamodels.STATE_ENABLE,
		CreateAt:time.Now().Unix(),
		UpdateAt:time.Now().Unix(),
	}

	err := p.ProductService.InsertProduct(product)
	if err != nil {
		return errorReturnView(p.Ctx, err.Error(), "/product/all", 500)
	}

	return messageThenRedirect("添加成功", "/product/all")
}

//商品编辑页
func (p *ProductController) GetUpdate(id int64) mvc.View {
	product, err := p.ProductService.GetProductByID(uint64(id))
	if err != nil {
		return errorReturnView(p.Ctx, err.Error(), "/product/all", 500)
	}

	productMap := structs.Map(product)

	productMap["CreateAtStr"] = time.Unix(reflect.ValueOf(productMap["CreateAt"]).Int(), 0).Format("2006-01-02 15:04:05")
	productMap["UpdateAtStr"] = time.Unix(reflect.ValueOf(productMap["UpdateAt"]).Int(), 0).Format("2006-01-02 15:04:05")
	productMap["UrlDetail"] = "/product/one?id=" + string(reflect.ValueOf(productMap["Id"]).Uint())

	return mvc.View {
		Name : "product/update",
		Data : iris.Map {
			"product" : productMap,
			"urlSubmit" : "/product/update",
			"urlBack" : "/product/all",
		},
	}
}

//商品编辑提交
func (p *ProductController) PostUpdate() mvc.View{
	id := uint64(p.Ctx.PostValueInt64Default("id", 0))
	name := p.Ctx.FormValue("name")
	num := p.Ctx.PostValueInt64Default("num", 0)
	image := p.Ctx.FormValue("image")
	url := p.Ctx.FormValue("url")

	if id <= 0 {
		return messageThenRedirect("id错误", "/product/all")
	}

	product, err := p.ProductService.GetProductByID(id)
	if err != nil {
		return errorReturnView(p.Ctx, err.Error(), "/product/all", 500)
	}

	product.ProductName = name
	product.Num = uint32(num)
	product.Image = image
	product.Url = url
	product.UpdateAt = time.Now().Unix()

	err = p.ProductService.UpdateProduct(product)
	if err != nil {
		return errorReturnView(p.Ctx, err.Error(), "/product/all", 500)
	}
	return messageThenRedirect("编辑成功", "/product/all")
}

func (p *ProductController) GetDelete(id int64) mvc.View{
	product, err := p.ProductService.GetProductByID(uint64(id))
	if err != nil {
		return errorReturnView(p.Ctx, err.Error(), "/product/all", 500)
	}

	err = p.ProductService.DeleteProduct(product)
	if err != nil {
		return errorReturnView(p.Ctx, err.Error(), "/product/all", 500)
	}

	return messageThenRedirect("删除成功", "/product/all")
}
