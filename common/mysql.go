package common

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"io"
	"strconv"
)


type MysqlPool struct {
	pool *Pool
}

func NewMysqlPool(consul *ConsulClient) (mysqlPool *MysqlPool, err error) {
	serviceName := "miaosha-demo-proxysql"
	serviceInfoList, err := consul.GetServiceListByName(serviceName)
	if err != nil {
		return nil, err
	}

	var addressList []map[string]string
	for _,serviceInfo := range serviceInfoList.List {
		address := map[string]string{
			"host" : serviceInfo.Host,
			"port" : strconv.Itoa(serviceInfo.Port),
			"user" : "user1",
			"password" : "password1",
			"db" : "miaosha-product",
		}

		addressList = append(addressList, address)
	}

	fmt.Println(addressList)

	makeFunc := func(address map[string]string) (io.Closer, error) {
		//*gorm.DB
		dsn := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", address["user"], address["password"], address["host"], address["port"], address["db"])
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
	poolConfig, err := NewPoolConfig(6, 10, 3600, addressList, 0, makeFunc, validateFunc)
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
