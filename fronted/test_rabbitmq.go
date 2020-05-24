package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"log"
	"miaosha-demo/common"
	"time"
)

func main() {
	common.NewZapLogger()
	defer func() {
		//recover panic, only for unexpect exception
		r := recover()

		if r != nil {
			var rerr error
			switch e := r.(type) {
			case string:
				rerr = errors.New(e)
			case error:
				rerr = e
			default:
				rerr = errors.New(fmt.Sprintf("%v", e))
			}
			common.ZapError("recover error", rerr)
		}

		zap.L().Sync()
	} ()

	flagSend := flag.Int("send", 1, "mq send")
	flagOffset := flag.Int("offset", 0, "message offset")
	flagLimit := flag.Int("limit", 0, "message limit")
	flag.Parse()

	if *flagSend == 1 {
		send(*flagOffset, *flagLimit)
	} else {
		receive()
	}
}


func send(offset, limit int) {
	config, err := common.NewConfigConsul()
	if err != nil {
		common.ZapError("new config fail", err)
		return
	}

	cache := common.NewFreeCacheClient(10)
	consul, err := common.NewConsulClient(config, cache)
	if err != nil {
		common.ZapError("new consul fail", err)
		return
	}

	mqPool, err := common.NewRabbitmqPool(consul)
	if err != nil {
		common.ZapError("new rabbitmq pool fail", err)
		return
	}

	conn, err := mqPool.Get()
	defer mqPool.Put(conn)
	if err != nil {
		common.ZapError("mq get fail", err)
		return
	}

	ch, err := conn.Channel()
	defer ch.Close()
	if err != nil {
		common.ZapError("mq new channel fail", err)
		return
	}

	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	if err := ch.Confirm(false); err != nil {
		common.ZapError("mq confirm mode fail", err)
	}

	channelReturn := make(chan amqp.Return)
	ch.NotifyReturn(channelReturn)

	go func() {
		for ret := range channelReturn {
			zap.L().Error(fmt.Sprintf("mq return %v", ret))
		}
	}()

	go func() {
		//confirmed := <-confirms
		for confirmed := range confirms {
			if !confirmed.Ack {
				zap.L().Info(fmt.Sprintf("mq confirm %v", confirmed))
			}
		}
	}()

	end := offset + limit
	timeStart := time.Now()
	for i := offset; i <= end; i++ {
		pid := i * 2
		uid := i
		body := fmt.Sprintf("%d_%d", pid, uid)

		//使用Channel.NotifyReturn 处理发送失败被返回的消息
		//Channel.NotifyPublish（添加监听） + Channel.Confirm（进入confirm模式） 确保所有消息发送成功
		err = ch.Publish(
			"miaosha_demo_exchange",          // exchange
			"aaa.bbb.ccc", // routing key 绑定键可以模糊，发送消息的路由键不能模糊
			false, //没有绑定的队列时，true返回消息，false丢弃
			false, //建议false，否则会发不到队列。没有消费者时，true返回，false丢弃
			amqp.Publishing{
				//DeliveryMode: amqp.Persistent,	//消息持久化， queue durable+消息持久化，才能不丢消息
				ContentType: "text/plain",
				Body:        []byte(body),
				//Expiration:"5000",
			})

		common.FailOnError(err, "Failed to publish a message")
	}

	timeEnd := time.Now()
	timeTotal := timeEnd.Sub(timeStart).Milliseconds()
	timeAvg := timeTotal / int64(limit)

	fmt.Println(fmt.Sprintf("mq send: %d, time total: %d ms, time avg: %d ms", limit, timeTotal, timeAvg))
}


func receive() {
	config, err := common.NewConfigConsul()
	if err != nil {
		common.ZapError("new config fail", err)
		return
	}

	freeCache := common.NewFreeCacheClient(10)
	consulClient, err := common.NewConsulClient(config, freeCache)
	if err != nil {
		common.ZapError("new consul fail", err)
		return
	}

	rabbitmqPool, err := common.NewRabbitmqPool(consulClient)
	if err != nil {
		common.ZapError("new rabbitmq pool fail", err)
		return
	}

	conn, err := rabbitmqPool.Get()
	//conn, err := amqp.Dial("amqp://root:root@172.18.0.99/:5672/")
	defer conn.Close()
	if err != nil {
		common.ZapError("rabbitmq get fail", err)
		return
	}

	ch, err := conn.Channel()
	defer ch.Close()
	if err != nil {
		common.ZapError("rabbitmq new channel fail", err)
		return
	}

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
	if err != nil {
		common.ZapError("rabbitmq exchange declare fail", err)
		return
	}

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
	if err != nil {
		common.ZapError("rabbitmq exchange ae declare fail", err)
		return
	}

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
	if err != nil {
		common.ZapError("rabbitmq exchange dead declare fail", err)
		return
	}

	//消费者流量控制
	err = ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		common.ZapError("rabbitmq qos fail", err)
		return
	}

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
	if err != nil {
		common.ZapError("rabbitmq queue dechalre fail", err)
		return
	}

	err = ch.QueueBind(
		q.Name,
		"aaa.*.ccc",
		"miaosha_demo_exchange",
		false,
		nil,
	)
	if err != nil {
		common.ZapError("rabbitmq queue bind fail", err)
		return
	}

	//备份队列
	qAe, err := ch.QueueDeclare(
		"miaosha_demo_queue_ae",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		common.ZapError("rabbitmq queue ae declare fail", err)
		return
	}

	err = ch.QueueBind(
		qAe.Name,
		"",
		"miaosha_demo_exchange_ae",
		false,
		nil,
	)
	if err != nil {
		common.ZapError("rabbitmq queue ae bind fail", err)
		return
	}

	//死信队列
	qDead, err := ch.QueueDeclare(
		"miaosha_demo_queue_dead",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		common.ZapError("rabbitmq queue dead declare fail", err)
		return
	}

	err = ch.QueueBind(
		qDead.Name,
		"",
		"miaosha_demo_exchange_dead",
		false,
		nil,
	)
	if err != nil {
		common.ZapError("rabbitmq queue dead bind fail", err)
		return
	}

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
	if err != nil {
		common.ZapError("rabbitmq register consumer fail", err)
		return
	}

	go func() {
		timeStart := time.Now()
		for d := range msgs {
			zap.L().Info("receive", zap.Int64("receive total ms", time.Now().Sub(timeStart).Milliseconds()))
			//消费后确认
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	forever := make(chan bool)
	<-forever
}






