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

	// 新增：可选的 exchange 预声明配置
	predeclareExchange     bool
	predeclaredExchange    string
	predeclaredExchangeTyp string
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

// NewRabbitMQClientWithExchange 高性能模式：初始化时声明 Exchange，后续 Publish/Consume 不再声明 Exchange
func NewRabbitMQClientWithExchange(url, exchange, exchangeType string, logger *zap.Logger) (*RabbitMQClient, error) {
	client := &RabbitMQClient{
		url: url, logger: logger,
		predeclareExchange:     true,
		predeclaredExchange:    exchange,
		predeclaredExchangeTyp: exchangeType,
	}
	err := client.connectWithRetry()
	if err != nil {
		return nil, err
	}
	// 初始化时声明 Exchange
	ch, err := client.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()
	err = ch.ExchangeDeclare(exchange, exchangeType, false, false, false, false, nil)
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

// Publish 支持自定义 exchange 和 routingKey，兼容原有用法
func (c *RabbitMQClient) PublishWithExchange(exchange, exchangeType, routingKey string, body []byte) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		ch, err := c.Channel()
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		defer ch.Close()
		// 高性能模式：只在初始化声明 Exchange
		if c.predeclareExchange && exchange == c.predeclaredExchange {
			// 不再声明 Exchange
		} else if exchange != "" {
			// 兼容模式：每次声明 Exchange
			err = ch.ExchangeDeclare(
				exchange, exchangeType, false, false, false, false, nil,
			)
			if err != nil {
				lastErr = err
				time.Sleep(500 * time.Millisecond)
				continue
			}
		}
		err = ch.PublishWithContext(context.Background(),
			exchange, routingKey, false, false,
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

// 兼容原有用法：Publish(queueName, body) 等价于 PublishWithExchange("", "", queueName, body)
func (c *RabbitMQClient) Publish(queueName string, body []byte) error {
	return c.PublishWithExchange("", "", queueName, body)
}

// Consume 支持自定义 exchange 和 routingKey，兼容原有用法
func (c *RabbitMQClient) ConsumeWithExchange(exchange, exchangeType, queueName, routingKey string, concurrency int, handler func(msg string)) error {
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
				// 高性能模式：只在初始化声明 Exchange
				if c.predeclareExchange && exchange == c.predeclaredExchange {
					// 不再声明 Exchange
				} else if exchange != "" {
					// 兼容模式：每次声明 Exchange
					err = ch.ExchangeDeclare(
						exchange, exchangeType, false, false, false, false, nil,
					)
					if err != nil {
						ch.Close()
						if c.logger != nil {
							c.logger.Error("[RabbitMQ] Exchange声明失败，1秒后重试", zap.Error(err))
						}
						time.Sleep(time.Second)
						continue
					}
				}
				// 声明队列并绑定到 exchange
				_, err = ch.QueueDeclare(
					queueName, false, false, false, false, nil,
				)
				if err != nil {
					ch.Close()
					if c.logger != nil {
						c.logger.Error("[RabbitMQ] 队列声明失败，1秒后重试", zap.Error(err))
					}
					time.Sleep(time.Second)
					continue
				}
				if exchange != "" {
					err = ch.QueueBind(
						queueName, routingKey, exchange, false, nil,
					)
					if err != nil {
						ch.Close()
						if c.logger != nil {
							c.logger.Error("[RabbitMQ] 队列绑定失败，1秒后重试", zap.Error(err))
						}
						time.Sleep(time.Second)
						continue
					}
				}
				msgs, err := ch.Consume(
					queueName, "", false, false, false, false, nil,
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

// 兼容原有用法：Consume(queueName, concurrency, handler) 等价于 ConsumeWithExchange("", "", queueName, queueName, concurrency, handler)
func (c *RabbitMQClient) Consume(queueName string, concurrency int, handler func(msg string)) error {
	return c.ConsumeWithExchange("", "", queueName, queueName, concurrency, handler)
}
