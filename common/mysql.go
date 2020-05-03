package common

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"io"
)


func NewMysqlPoolUser(consul *ConsulClient) (mysqlPool *MysqlPool, err error) {
	return NewMysqlPool(consul, "miaosha-user")
}


func NewMysqlPoolProduct(consul *ConsulClient) (mysqlPool *MysqlPool, err error) {
	return NewMysqlPool(consul, "miaosha-product")
}


type MysqlPool struct {
	pool *Pool
}

func NewMysqlPool(consul *ConsulClient, dbName string) (mysqlPool *MysqlPool, err error) {
	serviceName := "miaosha-demo-proxysql"
	serviceChan := consul.ChanList[serviceName]

	makeFunc := func(serviceInfo *ConsulServiceInfo) (io.Closer, error) {
		//*gorm.DB
		dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", "user1", "password1", serviceInfo.Host, serviceInfo.Port, dbName)
		//return gorm.Open("mysql", "root:root@(192.168.125.128:3306)/miaosha?charset=utf8&parseTime=True&loc=Local")
		return gorm.Open("mysql", dsn)
	}

	validateFunc := func(closer io.Closer) bool {
		db, ok := closer.(*gorm.DB)
		fmt.Println(ok)

		err := db.Exec("SELECT 1;").Error
		if err != nil {
			return false
		} else {
			return true
		}
	}

	//TODO get from consul kv
	poolConfig, err := NewPoolConfig(6, 10, 3600, serviceChan, makeFunc, validateFunc)
	pool, err :=  NewPool(poolConfig)

	mysqlPool = &MysqlPool{pool}
	return mysqlPool, err
}


func (mysqlPool MysqlPool) Get() (conn *gorm.DB, err error) {
	closer, err := mysqlPool.pool.Get()
	if err != nil {
		fmt.Println("get error", err)
		return
	}

	conn, ok := closer.(*gorm.DB)
	if !ok {
		fmt.Println("assert", ok)
	}

	return conn, nil
}


func (mysqlPool MysqlPool) Put(conn *gorm.DB) error {
	return mysqlPool.pool.Put(conn)
}


func (mysqlPool MysqlPool) Close(conn *gorm.DB) error {
	return mysqlPool.pool.CloseConn(conn)
}
