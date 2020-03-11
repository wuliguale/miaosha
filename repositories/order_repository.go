package repositories

import (
	"github.com/jinzhu/gorm"
	"miaosha-demo/datamodels"
)

type IOrder interface {
	Insert(*datamodels.Order) error
	InsertIgnore(*datamodels.Order) error
	Delete(*datamodels.Order) error
	Update(*datamodels.Order) error
	SelectByPk(uint64) (*datamodels.Order, error)
}

type OrderRepository struct {
	mysqlConn *gorm.DB
}


func NewOrderRepository(db *gorm.DB) IOrder {
	return &OrderRepository{mysqlConn: db}
}

func (o *OrderRepository) SelectByPk(oid uint64) (order *datamodels.Order, err error) {
	order = &datamodels.Order{}
	res := o.mysqlConn.First(order, oid)
	return order, res.Error
}

func (o *OrderRepository) Insert(order *datamodels.Order) error {
	return o.mysqlConn.Create(order).Error
}


func (o *OrderRepository) InsertIgnore(order *datamodels.Order) error {
	tableName := order.TableName()
	return o.mysqlConn.Exec("INSERT IGNORE INTO " + tableName + " (pid,uid,state,create_at) VALUES (?,?,?,?)", order.Pid, order.Uid, order.State, order.CreateAt).Error
}


func (o *OrderRepository) Update(order *datamodels.Order) error {
	return o.mysqlConn.Save(order).Error
}


func (o *OrderRepository) Delete(order *datamodels.Order) error {
	return o.mysqlConn.Delete(order).Error
}



