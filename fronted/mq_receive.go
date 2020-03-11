package main

import (
	"github.com/streadway/amqp"
	"log"
	"miaosha-demo/common"
	"miaosha-demo/datamodels"
	"miaosha-demo/repositories"
	"miaosha-demo/services"
	"strconv"
	"strings"
	"time"
)

func main() {
	conn, err := common.NewRabbitMqConn()
	common.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	//声明交换器，并指定备份交换器
	argTable := amqp.Table{"alternate-exchange":"miaosha_demo_exchange_ae"}
	err = ch.ExchangeDeclare(
		"miaosha_demo_exchange",
		"topic",
		true,
		false,
		false,
		false,
		argTable,
	)
	common.FailOnError(err, "Failed to declare exchange")

	//备份交换器
	err = ch.ExchangeDeclare(
		"miaosha_demo_exchange_ae",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	common.FailOnError(err, "Failed to declare exchange ae")

	//死信交换器
	err = ch.ExchangeDeclare(
		"miaosha_demo_exchange_dead",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	common.FailOnError(err, "Failed to declare exchange dead")

	//消费者流量控制
	err = ch.Qos(
		1,
		0,
		false,
	)
	common.FailOnError(err, "qos fail")

	//声明队列，并绑定死信交换器
	myQueueArgs := amqp.Table{
		"x-dead-letter-exchange" : "miaosha_demo_exchange_dead",
		"x-dead-letter-routing-key" : "miaosha_demo",
	}
	q, err := ch.QueueDeclare(
		"miaosha_demo_queue",
		true,
		false,
		false,
		false,
		myQueueArgs,
	)
	common.FailOnError(err, "Failed to declare queue")

	err = ch.QueueBind(
		q.Name,
		"aaa.*.ccc",
		"miaosha_demo_exchange",
		false,
		nil,
	)
	common.FailOnError(err, "Failed to bind queue")

	//备份队列
	qAe, err := ch.QueueDeclare(
		"miaosha_demo_queue_ae",
		true,
		false,
		false,
		false,
		nil,
	)
	common.FailOnError(err, "Failed to declare queue ae")

	err = ch.QueueBind(
		qAe.Name,
		"",
		"miaosha_demo_exchange_ae",
		false,
		nil,
	)
	common.FailOnError(err, "Failed to bind  queue ae")


	//死信队列
	qDead, err := ch.QueueDeclare(
		"miaosha_demo_queue_dead",
		true,
		false,
		false,
		false,
		nil,
	)
	common.FailOnError(err, "Failed to declare queue dead")

	err = ch.QueueBind(
		qDead.Name,
		"",
		"miaosha_demo_exchange_dead",
		false,
		nil,
	)
	common.FailOnError(err, "Failed to bind queue dead")


	//正常消费
	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	common.FailOnError(err, "Failed to register  consumer")

	//连接db
	db, err := common.NewMysqlConn()
	if err != nil {
		common.FailOnError(err, "new db fail")
	}

	orderRepository := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepository)

	go func() {
		for d := range msgs {
			uidPidSlice := strings.Split(string(d.Body), "_")
			pid, err := strconv.Atoi(uidPidSlice[0])
			if err != nil {
				common.FailOnError(err, "get pid fail")
			}
			uid , err := strconv.Atoi(uidPidSlice[1])
			if err != nil {
				common.FailOnError(err, "get uid fail")
			}

			order := &datamodels.Order{}
			order.Uid = uint32(uid)
			order.Pid = uint64(pid)
			order.State = datamodels.OrderWait
			order.CreateAt = time.Now().Unix()

			err = orderService.InsertIgnoreOrder(order)
			if err != nil {
				common.FailOnError(err, "add order fail")
			}

			//消费后确认
			d.Ack(false)
		}
	}()


	/*
	//备份队列消费
	msgsAe, err := ch.Consume(
		qAe.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	common.FailOnError(err, "Failed to register a consumer2")
	go func() {
		for d2 := range msgsAe {
			log.Printf("ae: [x] %v", d2)
			d2.Ack(false)
		}
	}()

	//死信队列消费
	msgsDead, err := ch.Consume(
		qDead.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	common.FailOnError(err, "Failed to register a consumer3")
	go func() {
		for d3 := range msgsDead {
			log.Printf("dead: [x] %v", d3)
			d3.Ack(false)
		}
	}()
	*/

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	forever := make(chan bool)
	<-forever

}

