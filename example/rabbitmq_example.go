package main

import (
	"fmt"
	"time"

	"github.com/muchinfo/mtp2-common-lib/mq"

	"go.uber.org/zap"
)

func RunRabbitMQExample() {
	url := "amqp://guest:guest@localhost:5672/"
	logger, _ := zap.NewDevelopment()

	// 高性能模式：初始化声明 Exchange，后续 Publish/Consume 不再声明 Exchange
	exchange := "entry"
	exchangeType := "direct"
	queue := "demo_queue"
	routingKey := "demo_key"

	client, err := mq.NewRabbitMQClientWithExchange(url, exchange, exchangeType, logger)
	if err != nil {
		panic(fmt.Sprintf("RabbitMQ 连接失败: %v", err))
	}
	defer client.Close()

	// 启动并发消费者（高性能模式）
	client.ConsumeWithExchange(exchange, exchangeType, queue, routingKey, 2, func(msg string) {
		logger.Info("[RabbitMQ Demo] Received", zap.String("body", msg))
	})

	// 并发发送消息到 direct 类型 exchange（高性能模式）
	for i := 0; i < 3; i++ {
		go func(i int) {
			body := []byte(fmt.Sprintf("Demo Message #%d %s", i+1, time.Now().Format("15:04:05")))
			client.PublishWithExchange(exchange, exchangeType, routingKey, body)
		}(i)
	}
	time.Sleep(2 * time.Second)

	// 兼容模式：每次声明 Exchange
	// client2, _ := mq.NewRabbitMQClient(url, logger)
	// client2.PublishWithExchange(exchange, exchangeType, routingKey, []byte("兼容模式消息"))
}
