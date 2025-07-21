package mq

import (
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRabbitMQClient_PublishConsume_DirectExchange(t *testing.T) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
	logger, _ := zap.NewDevelopment()
	exchange := "entry"
	exchangeType := "direct"
	queue := "test_queue_unit"
	routingKey := "test_key"
	msg := []byte("unit_test_message")
	ch := make(chan string, 1)

	// 高性能模式
	client, err := NewRabbitMQClientWithExchange(url, exchange, exchangeType, logger)
	if err != nil {
		t.Fatalf("连接 RabbitMQ 失败: %v", err)
	}
	defer client.Close()

	err = client.ConsumeWithExchange(exchange, exchangeType, queue, routingKey, 2, func(m string) {
		ch <- m
	})
	if err != nil {
		t.Fatalf("启动消费者失败: %v", err)
	}

	err = client.PublishWithExchange(exchange, exchangeType, routingKey, msg)
	if err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	select {
	case got := <-ch:
		if got != string(msg) {
			t.Errorf("消费消息内容不符, got=%s want=%s", got, string(msg))
		}
	case <-time.After(5 * time.Second):
		t.Error("消费消息超时")
	}

	// 兼容模式
	client2, err := NewRabbitMQClient(url, logger)
	if err != nil {
		t.Fatalf("连接 RabbitMQ 失败: %v", err)
	}
	defer client2.Close()
	err = client2.PublishWithExchange(exchange, exchangeType, routingKey, []byte("兼容模式消息"))
	if err != nil {
		t.Fatalf("兼容模式发送消息失败: %v", err)
	}
}
