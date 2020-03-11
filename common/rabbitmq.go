package common

import "github.com/streadway/amqp"

func NewRabbitMqConn() (*amqp.Connection, error){
	//TODO mq连接池
	return amqp.Dial("amqp://root:root@192.168.125.128:5672/")
}
