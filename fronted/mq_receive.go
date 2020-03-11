package main

import (
	"github.com/streadway/amqp"
	"log"
	"miaosha-demo/common"
)

func main() {
	conn, err := amqp.Dial("amqp://root:root@192.168.125.128:5672/")
	common.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.FailOnError(err, "Failed to open a channel")
	defer ch.Close()


	//设置此交换器的备份交换器，用于收集此发到此交换器的不匹配队列的消息
	argTable := amqp.Table{"alternate-exchange":"my_exchange_ae"}	//报错就删掉重新执行
	//完全相同的交换器（参数都相同）已存在，则不重复创建
	err = ch.ExchangeDeclare(
		"my_exchange",
		"topic",
		true,         // 重启后是否删除
		false,        //没有绑定是否删除
		false,        // 是否接受publish
		false,        //true，声明不需要等待server的确认，channel可能因为声明结果出错而关闭，使用NotifyClose 处理异常关闭
		argTable,          // 其他参数
	)
	common.FailOnError(err, "Failed to declare an exchange1")

	//备份交换器
	err = ch.ExchangeDeclare("my_exchange_ae", "fanout", true, false, false, false, nil)
	common.FailOnError(err, "Failed to declare an exchange2")


	//死信交换器
	err = ch.ExchangeDeclare("my_exchange_dead", "fanout", true, false, false, false, nil)
	common.FailOnError(err, "Failed to declare an exchange3")

	/*
	err = ch.Qos(
		1, //当前消费者一次能接受的最大消息数量
		0, //服务器传递的最大容量（以八位字节为单位）
		false, //如果设置为true 对channel可用
	)
	common.FailOnError(err, "qos fail")
	*/


	//设置队列的死信交换器，如果投入死信队列，则使用指定的route key(可用于区分业务)
	myQueueArgs := amqp.Table{"x-dead-letter-exchange" : "my_exchange_dead", "x-dead-letter-routing-key" : "aaa"}
	//每个新创建的队列都和空交换器""（type=direct）绑定，路由键是队列名
	//相同参数队列不存在则创建
	q, err := ch.QueueDeclare(
		"my_queue",    // 不指定则自动生成
		true, // 重启后是否消失
		false, // 没有消费者时是否删除
		false,  //只有在声明队列的连接中有效，声明队列的连接断开则删除队列，在其他连接上操作报错
		false, //=true时，队列已存在，或在其他连接中修改已存在队列，则报错
		myQueueArgs,   // arguments
	)
	common.FailOnError(err, "Failed to declare a queue")


	//多次绑定不报错，多个绑定规则匹配一个队列也只发一次
	//durable队列只能绑定durable交换器
	err = ch.QueueBind(
		q.Name,       // queue name
		"aaa.*.ccc",            // routing key   绑定键可以模糊，发送消息的路由键不能模糊
		"my_exchange", // exchange
		false, //=false 且 队列不能被绑定，则channel报错
		nil)
	common.FailOnError(err, "Failed to bind a queue")




	//备份队列
	q2, err := ch.QueueDeclare(
		"my_queue_ae",
		true,
		false,
		false,
		false,
		nil,
	)
	common.FailOnError(err, "Failed to declare a queue")
	err = ch.QueueBind(q2.Name, "", "my_exchange_ae", false, nil)
	common.FailOnError(err, "Failed to bind a queue")


	//死信队列
	q3, err := ch.QueueDeclare(
		"my_queue_dead",
		true,
		false,
		false,
		false,
		nil,
	)
	common.FailOnError(err, "Failed to declare a queue")
	//死信队列绑定死信交换器
	err = ch.QueueBind(
		q3.Name,
		"",
		"my_exchange_dead",
		false,
		nil)
	common.FailOnError(err, "Failed to bind a queue")



	//所有收到的消息处理成功后都应该ack
	//如果消费者取消，或连接,channel断开，则未被ack的消息会回到队列尾
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer，同一channel上的唯一消费者名，指定名称可以方便取消消费者，不指定则自动生成
		false,   //不需要ack，自动确认，可能消费没成功
		false,  //队列只能有一个消费者
		false,  //不支持
		false,  // =true不等待server确认请求就开始传递消息，如果不能消费，channel报错并关闭
		nil,    // args
	)
	common.FailOnError(err, "Failed to register a consumer1")

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)
			d.Ack(false) ////手动确认2，同一channel
		}
	}()


msgs2, err := ch.Consume(
	q2.Name, // queue
	"",     // consumer，同一channel上的唯一消费者名，指定名称可以方便取消消费者，不指定则自动生成
	false,   //不需要ack，自动确认，可能消费没成功
	false,  //队列只能有一个消费者
	false,  //不支持
	false,  // =true不等待server确认请求就开始传递消息，如果不能消费，channel报错并关闭
	nil,    // args
)
common.FailOnError(err, "Failed to register a consumer2")
go func() {
	for d2 := range msgs2 {
		log.Printf("ae: [x] %v", d2)
		d2.Ack(false)
	}
}()


msgs3, err := ch.Consume(
	q3.Name,
	"",
	false,
	false,
	false,
	false,
	nil,
)
common.FailOnError(err, "Failed to register a consumer3")
go func() {
	for d3 := range msgs3 {
		log.Printf("dead: [x] %v", d3)
		d3.Ack(false)
	}
}()




	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	forever := make(chan bool)
	<-forever

}

