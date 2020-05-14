package main

import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"miaosha-demo/common"
)

func main() {
	flagNum := flag.Int("num", 0, "message num")
	flag.Parse()
	num := *flagNum

	config, err := common.NewConfigConsul()
	common.FailOnError(err, "fail to new config")

	cache := common.NewFreeCacheClient(10)
	consul, err := common.NewConsulClient(config, cache)
	common.FailOnError(err, "fail to new consul")

	mqPool, err := common.NewRabbitmqPool(consul)
	common.FailOnError(err, "fail to new mq pool")

	conn, err := mqPool.Get()
	defer mqPool.Put(conn)
	common.FailOnError(err, "fail to get conn from pool")

	ch, err := conn.Channel()
	common.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	/*
	//设置此交换器的备份交换器，用于收集此发到此交换器的不匹配队列的消息
	argTable := amqp.Table{"alternate-exchange":"my_exchange_ae"}
	//完全相同的交换器（参数都相同）已存在，则不重复创建
	err = ch.ExchangeDeclare(
		"my_exchange", // name
		"topic",      // type
		true,         // 重启后是否删除
		false,        //没有绑定是否删除
		false,        // 是否接受publish
		false,        //true，声明不需要等待server的确认，channel可能因为声明结果出错而关闭，使用NotifyClose 处理异常关闭
		argTable,          // 其他参数
	)
	common.FailOnError(err, "Failed to declare an exchange")

	err = ch.ExchangeDeclare("my_exchange_ae", "fanout", true, false, false, false, nil)
	common.FailOnError(err, "Failed to declare an exchange2")

	*/

	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	if err := ch.Confirm(false); err != nil {
		common.FailOnError(err, "confirm mode fail")
	}

	channelReturn := make(chan amqp.Return)
	ch.NotifyReturn(channelReturn)

	go func() {
		for ret := range channelReturn {
			log.Printf("return: [x] %v", ret)
		}
	}()

	go func() {
		//confirmed := <-confirms
		for confirmed := range confirms {
			log.Printf("confirm: [x] %v", confirmed)
		}
	}()

	for i := 0; i < num; i++ {
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
}


