package repositories

import (
	"miaosha-demo/common"
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
	mysqlPool *common.MysqlPool
}


func NewOrderRepository(mysqlPool *common.MysqlPool) IOrder {
	return &OrderRepository{mysqlPool: mysqlPool}
}

func (o *OrderRepository) SelectByPk(oid uint64) (order *datamodels.Order, err error) {
	db, err := o.mysqlPool.Get()
	defer o.mysqlPool.Put(db)
	if err != nil {
		return nil, err
	}

	order = &datamodels.Order{}
	res := db.First(order, oid)
	return order, res.Error
}

func (o *OrderRepository) Insert(order *datamodels.Order) error {
	db, err := o.mysqlPool.Get()
	defer o.mysqlPool.Put(db)
	if err != nil {
		return  err
	}
	return db.Create(order).Error
}


func (o *OrderRepository) InsertIgnore(order *datamodels.Order) error {
	db, err := o.mysqlPool.Get()
	defer o.mysqlPool.Put(db)
	if err != nil {
		return  err
	}

	tableName := order.TableName()
	return db.Exec("INSERT IGNORE INTO " + tableName + " (pid,uid,state,create_at) VALUES (?,?,?,?)", order.Pid, order.Uid, order.State, order.CreateAt).Error
}


func (o *OrderRepository) Update(order *datamodels.Order) error {
	db, err := o.mysqlPool.Get()
	defer o.mysqlPool.Put(db)
	if err != nil {
		return err
	}

	return db.Save(order).Error
}


func (o *OrderRepository) Delete(order *datamodels.Order) error {
	db, err := o.mysqlPool.Get()
	defer o.mysqlPool.Put(db)
	if err != nil {
		return err
	}

	return db.Delete(order).Error
}

