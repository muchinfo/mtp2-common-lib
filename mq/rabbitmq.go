package mq

import (
	"context"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// RabbitMQClient 支持并发安全的连接和通道管理
type RabbitMQClient struct {
	url    string
	conn   *amqp.Connection
	mutex  sync.Mutex
	closed bool
	logger *zap.Logger
}

// NewRabbitMQClient 创建客户端
// logger 允许为 nil，若为 nil 则不输出日志
func NewRabbitMQClient(url string, logger *zap.Logger) (*RabbitMQClient, error) {
	client := &RabbitMQClient{url: url, logger: logger}
	err := client.connectWithRetry()
	if err != nil {
		return nil, err
	}
	return client, nil
}

// 尝试连接，断网时自动重试
func (c *RabbitMQClient) connectWithRetry() error {
	var err error
	for i := range 10 { // 最多重试10次
		c.conn, err = amqp.Dial(c.url)
		if err == nil {
			return nil
		}
		if c.logger != nil {
			c.logger.Warn("[RabbitMQ] 连接失败，稍后重试", zap.Int("retry", i+1), zap.Error(err))
		}
		time.Sleep(time.Duration(2<<i) * time.Millisecond) // 指数退避
	}
	return err
}

// Channel 获取新通道（每个 goroutine 独立使用）
func (c *RabbitMQClient) Channel() (*amqp.Channel, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed {
		return nil, amqp.ErrClosed
	}
	if c.conn == nil || c.conn.IsClosed() {
		if err := c.connectWithRetry(); err != nil {
			return nil, err
		}
	}
	return c.conn.Channel()
}

// Close 关闭连接
func (c *RabbitMQClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.closed = true
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// 并发安全的生产者
func (c *RabbitMQClient) Publish(queueName string, body []byte) error {
	var lastErr error
	for i := 0; i < 3; i++ { // 最多重试3次
		ch, err := c.Channel()
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		defer ch.Close()
		_, err = ch.QueueDeclare(
			queueName, true, false, false, false, nil,
		)
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		err = ch.PublishWithContext(context.Background(),
			"", queueName, false, false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        body,
			},
		)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}
	return lastErr
}

// 并发消费者，支持指定并发数
func (c *RabbitMQClient) Consume(queueName string, concurrency int, handler func(msg string)) error {
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				ch, err := c.Channel()
				if err != nil {
					if c.logger != nil {
						c.logger.Error("[RabbitMQ] 获取通道失败，1秒后重试", zap.Error(err))
					}
					time.Sleep(time.Second)
					continue
				}
				_, err = ch.QueueDeclare(
					queueName, true, false, false, false, nil,
				)
				if err != nil {
					ch.Close()
					if c.logger != nil {
						c.logger.Error("[RabbitMQ] 队列声明失败，1秒后重试", zap.Error(err))
					}
					time.Sleep(time.Second)
					continue
				}
				msgs, err := ch.Consume(
					queueName, "", true, false, false, false, nil,
				)
				if err != nil {
					ch.Close()
					if c.logger != nil {
						c.logger.Error("[RabbitMQ] 消费失败，1秒后重试", zap.Error(err))
					}
					time.Sleep(time.Second)
					continue
				}
				for d := range msgs {
					handler(string(d.Body))
				}
				ch.Close()
				// 如果 msgs 被关闭，说明连接断开，自动重连
				if c.logger != nil {
					c.logger.Warn("[RabbitMQ] 消费通道关闭，1秒后重连")
				}
				time.Sleep(time.Second)
			}
		}()
	}
	go func() { wg.Wait() }()
	return nil
}
