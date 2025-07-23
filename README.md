# mtp2-common-lib

Muchinfo MTP2 通用 Go 组件库，适用于企业级微服务、后台系统等场景，涵盖日志、配置、消息队列、数据库等常用基础能力。

## 目录结构

- logger/    —— 高性能 zap 日志组件，支持日志轮转、结构化、糖化日志
- config/    —— 基于 viper 的多文件配置加载与热更新
- mq/        —— RabbitMQ 并发安全客户端，支持 zap 日志、断网重连
- database/  —— Oracle 数据库 xorm 封装，支持 zap 日志、慢SQL、熔断
- http/      —— 标准 HTTP 请求、签名、加解密等工具
- socket/    —— TCP 网络通信组件，支持客户端、服务器、自动重连、消息广播
- websocket/ —— WebSocket 网络通信组件，支持客户端、服务器、自动重连、消息广播
- example/   —— 各模块独立示例

## 快速开始

### 1. 日志 logger

高性能 zap 日志，支持自定义配置、开发/生产模式、日志轮转。

```go
import "github.com/muchinfo/mtp2-common-lib/logger"
config := &logger.Config{ /* ... */ }
if err := logger.Init(config); err != nil {
    panic(err)
}
logger.Info("hello", zap.String("key", "value"))
```

// 兼容老用法（仅本项目内部用）
// logger.Init(nil) // 或 InitDevelopment/InitProduction

详见 [logger/README.md](logger/README.md)

### 2. 配置 config

多文件合并、热更新，基于 viper。

```go
import "github.com/muchinfo/mtp2-common-lib/config"
var cfg struct{ Name string }
config.InitViper([]string{"config.yaml"}, &cfg, func(e fsnotify.Event) { /* ... */ })
```

详见 [config/README.md](config/README.md)

### 3. 消息队列 mq

RabbitMQ 客户端，支持 zap.Logger 注入、并发消费、断网重连。

```go
import (
    "github.com/muchinfo/mtp2-common-lib/mq"
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
    "github.com/muchinfo/mtp2-common-lib/database"
    "go.uber.org/zap"
    "time"
)
logger, _ := zap.NewProduction()
breaker := database.NewCircuitBreaker(3)
engine, _ := database.NewOracleEngine("user/pwd@host:port/sid", logger, 100*time.Millisecond, breaker)
database.AutoMigrate(engine)
```

详见 [database/README.md](database/README.md)

### 5. HTTP 工具 http

标准 HTTP 请求工具，支持 context、自定义 client、logger，内置常用签名算法（MD5、HMAC-SHA256、RSA）及 RSA 加解密。

```go
import "github.com/muchinfo/mtp2-common-lib/http"
// GET 请求
resp, status, header, err := http.HttpCall("GET", "https://httpbin.org/get", nil, nil, nil)
// POST JSON
data := map[string]any{"foo": "bar"}
resp, status, _, err := http.HttpCall("POST", "https://httpbin.org/post", data, nil, nil)
```

详见 [http/README.md](http/README.md)

### 6. TCP 网络通信 socket

TCP 客户端和服务器组件，支持自动重连、消息广播、回调处理等企业级功能。

```go
import "github.com/muchinfo/mtp2-common-lib/socket"
// TCP 客户端
config := socket.TCPClientConfig{
    Address: "localhost:8080",
    AutoReconnect: true,
}
client := socket.NewTCPClient(config)
client.Connect()
client.SendString("Hello, Server!")

// TCP 服务器
serverConfig := socket.TCPServerConfig{
    Address: ":8080",
    MaxConnections: 100,
}
server := socket.NewTCPServer(serverConfig)
server.Start()
server.BroadcastString("Hello, All Clients!")
```

详见 [socket/README.md](socket/README.md)

### 7. WebSocket 网络通信

WebSocket 组件为调用者提供企业级的 WebSocket 客户端和服务器功能，支持自动重连、连接管理、消息广播等高级特性。

```go
import "github.com/muchinfo/mtp2-common-lib/websocket"

// WebSocket客户端
clientConfig := websocket.WSClientConfig{
    URL:            "ws://localhost:8080/ws",
    AutoReconnect:  true,
    ReconnectDelay: 5 * time.Second,
}
client := websocket.NewWSClient(clientConfig)

// WebSocket服务器  
serverConfig := websocket.WSServerConfig{
    Address:        ":8080",
    Path:           "/ws",
    MaxConnections: 100,
}
server := websocket.NewWSServer(serverConfig)
```

详见 [websocket/README.md](websocket/README.md)

### 8. 示例

所有模块均有独立 example 文件，见 [example/](example/)

## 依赖

- [go.uber.org/zap](https://github.com/uber-go/zap)
- [github.com/spf13/viper](https://github.com/spf13/viper)
- [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)
- [xorm.io/xorm](https://xorm.io/)
- [github.com/godror/godror](https://github.com/godror/godror)
- [github.com/fsnotify/fsnotify](https://github.com/fsnotify/fsnotify)
- [github.com/gorilla/websocket](https://github.com/gorilla/websocket)
