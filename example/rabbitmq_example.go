package main

import (
	"fmt"
	"mtp2-common-lib/mq"
	"time"

	"go.uber.org/zap"
)

func RunRabbitMQExample() {
	url := "amqp://guest:guest@localhost:5672/"
	logger, _ := zap.NewDevelopment()
	client, err := mq.NewRabbitMQClient(url, logger)
	if err != nil {
		panic(fmt.Sprintf("RabbitMQ 连接失败: %v", err))
	}
	defer client.Close()

	queue := "demo_queue"

	// 启动并发消费者
	client.Consume(queue, 2, func(msg string) {
		logger.Info("[RabbitMQ Demo] Received", zap.String("body", msg))
	})

	// 并发发送消息
	for i := 0; i < 3; i++ {
		go func(i int) {
			body := []byte(fmt.Sprintf("Demo Message #%d %s", i+1, time.Now().Format("15:04:05")))
			client.Publish(queue, body)
		}(i)
	}
	time.Sleep(2 * time.Second)
}
