package repositories

import (
	"errors"
	"github.com/jinzhu/gorm"
	"miaosha-demo/datamodels"
)

type IProduct interface {
	Insert(*datamodels.Product) error
	Delete(*datamodels.Product) error
	Update(*datamodels.Product) error
	SelectByPk(uint64) (*datamodels.Product, error)
	SelectAll(int64, int64, string, bool) ([]*datamodels.Product, error)
}

type ProductRepository struct {
	mysqlConn *gorm.DB
}

func NewProductRepository(db *gorm.DB) IProduct {
	return &ProductRepository{mysqlConn: db}
}

func (p *ProductRepository) SelectByPk(pid uint64) (product *datamodels.Product, err error) {
	product = &datamodels.Product{}
	res := p.mysqlConn.First(product, pid)
	return product, res.Error
}


func (p *ProductRepository) SelectAll(page, pageSize int64, orderField string, orderDesc bool) (productList []*datamodels.Product, err error){
	if page < 1{
		return nil, errors.New("page error")
	}

	if pageSize < 1 {
		return nil, errors.New("pageSize error")
	}

	offset := (page - 1) * pageSize
	orderBy := orderField
	if orderDesc {
		orderBy += " DESC"
	} else {
		orderBy += " ASC"
	}

	productList = []*datamodels.Product{}
	res := p.mysqlConn.Order(orderBy).Offset(offset).Limit(pageSize).Find(&productList)
	return productList, res.Error
}


func (p *ProductRepository) Insert(product *datamodels.Product) error {
	return p.mysqlConn.Create(product).Error
}


func (p *ProductRepository) Update(product *datamodels.Product) error {
	return p.mysqlConn.Save(product).Error
}


func (p *ProductRepository) Delete(product *datamodels.Product) error {
	return p.mysqlConn.Delete(product).Error
}


