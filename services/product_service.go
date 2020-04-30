package services

import (
	"miaosha-demo/datamodels"
	"miaosha-demo/repositories"
)

type IProductService interface {
	GetProductByID(uint64) (*datamodels.Product, error)
	GetProductAll(int64, int64, string, bool) ([]*datamodels.Product, error)
	InsertProduct(*datamodels.Product) error
	UpdateProduct(*datamodels.Product) error
	DeleteProduct(*datamodels.Product) error
}

type ProductService struct {
	productRepository repositories.IProduct
}

func NewProductService(repository repositories.IProduct) IProductService {
	return &ProductService{repository}
}

func (p *ProductService) GetProductByID(pid uint64) (*datamodels.Product, error) {
	return p.productRepository.SelectByPk(pid)
}


func (p *ProductService) GetProductAll(page, pageSize int64, orderField string, orderDesc bool) (productList []*datamodels.Product, err error) {
	return p.productRepository.SelectAll(page, pageSize, orderField, orderDesc)
}

func (p *ProductService) InsertProduct(product *datamodels.Product) error {
	return p.productRepository.Insert(product)
}

func (p *ProductService) UpdateProduct(product *datamodels.Product) error {
	return p.productRepository.Update(product)
}

func (p *ProductService) DeleteProduct(product *datamodels.Product) error {
	return p.productRepository.Delete(product)
}


