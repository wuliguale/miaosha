package repositories

import (
	"errors"
	"miaosha-demo/common"
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
	mysqlPool *common.MysqlPool
}

func NewProductRepository(mysqlPool *common.MysqlPool) IProduct {
	return &ProductRepository{mysqlPool:mysqlPool}
}

func (p *ProductRepository) SelectByPk(pid uint64) (product *datamodels.Product, err error) {
	db, err := p.mysqlPool.Get()
	defer p.mysqlPool.Put(db)
	if err != nil {
		return nil, err
	}

	product = &datamodels.Product{}
	res := db.First(product, pid)
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

	db, err := p.mysqlPool.Get()
	defer p.mysqlPool.Put(db)
	if err != nil {
		return nil, err
	}

	productList = []*datamodels.Product{}
	res := db.Order(orderBy).Offset(offset).Limit(pageSize).Find(&productList)
	return productList, res.Error
}


func (p *ProductRepository) Insert(product *datamodels.Product) error {
	db, err := p.mysqlPool.Get()
	defer p.mysqlPool.Put(db)
	if err != nil {
		return  err
	}
	return db.Create(product).Error
}


func (p *ProductRepository) Update(product *datamodels.Product) error {
	db, err := p.mysqlPool.Get()
	defer p.mysqlPool.Put(db)
	if err != nil {
		return err
	}

	return db.Save(product).Error
}


func (p *ProductRepository) Delete(product *datamodels.Product) error {
	db, err := p.mysqlPool.Get()
	defer p.mysqlPool.Put(db)
	if err != nil {
		return  err
	}

	return db.Delete(product).Error
}


