# RabbitMQ 并发安全客户端

本组件基于 [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go) 封装，支持并发安全、断网自动重连、可注入 zap.Logger 的 RabbitMQ 生产者与消费者。

## 主要特性

- 支持多 goroutine 并发安全收发消息
- 断网/异常自动重连，生产和消费均自动重试
- 消费端支持指定并发数
- 支持 zap.Logger 日志注入，日志统一管理
- 简单易用的 API

## 快速开始

### 1. 注入 zap.Logger

```go
import (
    "go.uber.org/zap"
    "mtp2-common-lib/mq"
)

logger, _ := zap.NewProduction()
client, _ := mq.NewRabbitMQClient("amqp://guest:guest@localhost:5672/", logger)
defer client.Close()
```

### 2. 生产消息

```go
client.Publish("queue_name", []byte("hello world"))
```

### 3. 消费消息

```go
client.Consume("queue_name", 2, func(msg string) {
    logger.Info("收到消息", zap.String("body", msg))
})
```

### 4. 断网重连

- 组件内部自动处理，无需手动干预。

### 5. 示例

见 `example/rabbitmq_example.go`。

## 日志说明

- 所有内部日志均通过 zap.Logger 输出，便于统一管理和格式化。
- logger 允许为 nil，若为 nil 则不输出内部日志。

## 单元测试

```shell
go test ./mq
```

## 依赖

- [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)
- [go.uber.org/zap](https://github.com/uber-go/zap)
