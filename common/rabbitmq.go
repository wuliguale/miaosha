package common

import (
	"fmt"
	"github.com/streadway/amqp"
	"io"
	"strconv"
)


type RabbitmqPool struct {
	pool *Pool
}


func NewRabbitmqPool(consul *ConsulClient) (rabbitmqPool *RabbitmqPool, err error) {
	serviceName := "miaosha-demo-rabbitmq"
	serviceInfoList, err := consul.GetServiceListByName(serviceName)
	if err != nil {
		return nil, err
	}

	var addressList []map[string]string
	for _,serviceInfo := range serviceInfoList.List {
		address := map[string]string{
			"host" : serviceInfo.Host,
			"port" : strconv.Itoa(serviceInfo.Port),
			"user" : "root",
			"password" : "root",
		}

		addressList = append(addressList, address)
	}

	fmt.Println(addressList)

	makeFunc := func(address map[string]string) (io.Closer, error) {
		//*amqp.Connection
		//return amqp.Dial("amqp://root:root@172.18.0.99:5672/")
		url := fmt.Sprintf("amqp://%s:%s@%s:%s/", address["user"], address["password"], address["host"], address["port"])
		fmt.Println(url)
		return amqp.Dial(url)
	}

	//TODO get from consul kv
	poolConfig, err := NewPoolConfig(1, 3, 3600, addressList, 0, makeFunc, nil)
	pool, err :=  NewPool(poolConfig)

	rabbitmqPool = &RabbitmqPool{pool}
	return rabbitmqPool, err
}


func (rabbitmqPool *RabbitmqPool) Get() (conn *amqp.Connection, err error) {
	closer, err := rabbitmqPool.pool.Get()
	if err != nil {
		fmt.Println("get error", err)
		return
	}

	conn, ok := closer.(*amqp.Connection)
	if !ok {
		fmt.Println("assert", ok)
	}

	return conn, nil
}


func (rabbitmqPool *RabbitmqPool) Put(conn *amqp.Connection) error {
	return rabbitmqPool.pool.Put(conn)
}


func (rabbitmqPool *RabbitmqPool) Close(conn *amqp.Connection) error {
	return rabbitmqPool.pool.CloseConn(conn)
}

