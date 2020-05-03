package common

import (
	"fmt"
	"github.com/streadway/amqp"
	"io"
)


type RabbitmqPool struct {
	pool *Pool
}


func NewRabbitmqPool(consul *ConsulClient) (rabbitmqPool *RabbitmqPool, err error) {
	serviceName := "miaosha-demo-rabbitmq"
	serviceChan := consul.ChanList[serviceName]

	makeFunc := func(serviceInfo *ConsulServiceInfo) (io.Closer, error) {
		//*amqp.Connection
		//return amqp.Dial("amqp://root:root@172.18.0.99:5672/")
		url := fmt.Sprintf("amqp://%s:%s@%s:%d/", "root", "root", serviceInfo.Host, serviceInfo.Port)
		fmt.Println(url)
		return amqp.Dial(url)
	}

	//TODO get from consul kv
	poolConfig, err := NewPoolConfig(1, 3, 3600, serviceChan, makeFunc, nil)
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

