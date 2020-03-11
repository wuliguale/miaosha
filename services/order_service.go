package services

import (
	"miaosha-demo/datamodels"
	"miaosha-demo/repositories"
)

type IOrderService interface {
	GetOrderByID(uint64) (*datamodels.Order, error)
	InsertOrder(*datamodels.Order) error
	InsertIgnoreOrder(*datamodels.Order) error
	UpdateOrder(*datamodels.Order) error
	DeleteOrder(*datamodels.Order) error
}

type OrderService struct {
	OrderRepository repositories.IOrder
}

func NewOrderService(repository repositories.IOrder) IOrderService {
	return &OrderService{OrderRepository: repository}
}

func (o *OrderService) GetOrderByID(oid uint64) (order *datamodels.Order, err error) {
	return o.OrderRepository.SelectByPk(oid)
}

func (o *OrderService) InsertOrder(order *datamodels.Order) error {
	return o.OrderRepository.Insert(order)
}

func (o *OrderService) InsertIgnoreOrder(order *datamodels.Order) error {
	return o.OrderRepository.InsertIgnore(order)
}

func (o *OrderService) UpdateOrder(order *datamodels.Order) error {
	return o.OrderRepository.Update(order)
}

func (o *OrderService) DeleteOrder(order *datamodels.Order) error {
	return o.OrderRepository.Delete(order)
}


