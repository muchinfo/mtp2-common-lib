# mtp2-common-lib

Muchinfo MTP2 通用 Go 组件库，适用于企业级微服务、后台系统等场景，涵盖日志、配置、消息队列、数据库等常用基础能力。

## 目录结构

- logger/   —— 高性能 zap 日志组件，支持日志轮转、结构化、糖化日志
- config/   —— 基于 viper 的多文件配置加载与热更新
- mq/       —— RabbitMQ 并发安全客户端，支持 zap 日志、断网重连
- database/ —— Oracle 数据库 xorm 封装，支持 zap 日志、慢SQL、熔断
- example/  —— 各模块独立示例

## 快速开始

### 1. 日志 logger

高性能 zap 日志，支持自定义配置、开发/生产模式、日志轮转。

```go
import "mtp2-common-lib/logger"
logger.Init(nil) // 或 InitDevelopment/InitProduction
logger.Info("hello", zap.String("key", "value"))
```

详见 [logger/README.md](logger/README.md)

### 2. 配置 config

多文件合并、热更新，基于 viper。

```go
import "mtp2-common-lib/config"
var cfg struct{ Name string }
config.InitViper([]string{"config.yaml"}, &cfg, func(e fsnotify.Event) { /* ... */ })
```

详见 [config/README.md](config/README.md)

### 3. 消息队列 mq

RabbitMQ 客户端，支持 zap.Logger 注入、并发消费、断网重连。

```go
import (
    "mtp2-common-lib/mq"
    "go.uber.org/zap"
)
logger, _ := zap.NewProduction()
client, _ := mq.NewRabbitMQClient("amqp://guest:guest@localhost:5672/", logger)
client.Publish("queue", []byte("hello"))
```

详见 [mq/README.md](mq/README.md)

### 4. 数据库 database

Oracle 数据库 xorm 封装，支持 zap.Logger、慢SQL统计、熔断、健康检查。

```go
import (
    "mtp2-common-lib/database"
    "go.uber.org/zap"
    "time"
)
logger, _ := zap.NewProduction()
breaker := database.NewCircuitBreaker(3)
engine, _ := database.NewOracleEngine("user/pwd@host:port/sid", logger, 100*time.Millisecond, breaker)
database.AutoMigrate(engine)
```

详见 [database/README.md](database/README.md)

### 5. 示例

所有模块均有独立 example 文件，见 [example/](example/)

## 依赖

- [go.uber.org/zap](https://github.com/uber-go/zap)
- [github.com/spf13/viper](https://github.com/spf13/viper)
- [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)
- [xorm.io/xorm](https://xorm.io/)
- [github.com/godror/godror](https://github.com/godror/godror)

## 贡献

欢迎 issue、PR 与建议！
