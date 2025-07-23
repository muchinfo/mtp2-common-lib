# WebSocket 组件

WebSocket 组件为调用者提供企业级的 WebSocket 客户端和服务器功能，支持自动重连、连接管理、消息广播等高级特性。

## 功能特性

### WebSocket 客户端 (WSClient)

- ✅ **自动连接管理**: 支持自动连接和优雅断开
- ✅ **智能重连机制**: 可配置的自动重连策略
- ✅ **多消息格式**: 支持文本和JSON消息发送
- ✅ **事件回调系统**: 连接、断开、消息、错误回调
- ✅ **并发安全**: 使用读写锁确保线程安全
- ✅ **上下文管理**: 支持优雅的取消和超时控制
- ✅ **心跳检测**: 内置 Ping/Pong 保活机制

### WebSocket 服务器 (WSServer)

- ✅ **多客户端管理**: 高效管理多个并发连接
- ✅ **消息广播**: 支持向所有或特定客户端广播消息
- ✅ **连接限制**: 可配置最大连接数限制
- ✅ **事件回调系统**: 客户端连接、断开、消息回调
- ✅ **HTTP升级**: 标准的WebSocket握手和协议升级
- ✅ **跨域支持**: 可自定义跨域检查逻辑
- ✅ **优雅关闭**: 支持优雅的服务器关闭和资源清理

## 快速开始

### 1. WebSocket 服务器

```go
package main

import (
    "log"
    "time"
    
    "github.com/muchinfo/mtp2-common-lib/websocket"
)

func main() {
    // 创建服务器配置
    config := websocket.WSServerConfig{
        Address:         ":8080",          // 监听端口
        Path:            "/ws",            // WebSocket路径
        MaxConnections:  100,              // 最大连接数
        PingInterval:    30 * time.Second, // ping间隔
        PongWait:        60 * time.Second, // pong等待时间
        WriteWait:       10 * time.Second, // 写入等待时间
        ReadBufferSize:  4096,             // 读取缓冲区
        WriteBufferSize: 4096,             // 写入缓冲区
    }
    
    // 创建服务器实例
    server := websocket.NewWSServer(config)
    
    // 设置回调函数
    server.SetCallbacks(
        // 客户端连接回调
        func(client *websocket.WSClientConnection) {
            log.Printf("Client connected: %s", client.RemoteAddr)
            // 发送欢迎消息
            client.SendText("Welcome to WebSocket Server!")
        },
        // 客户端断开回调
        func(client *websocket.WSClientConnection, err error) {
            log.Printf("Client disconnected: %s", client.RemoteAddr)
        },
        // 消息接收回调
        func(client *websocket.WSClientConnection, data []byte) {
            message := string(data)
            log.Printf("Received from %s: %s", client.RemoteAddr, message)
            
            // 广播消息到所有客户端
            server.BroadcastText("Broadcast: " + message)
        },
        // 错误回调
        func(err error) {
            log.Printf("Server error: %v", err)
        },
    )
    
    // 启动服务器
    if err := server.Start(); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
    
    log.Printf("WebSocket server started on %s%s", config.Address, config.Path)
    
    // 保持运行
    select {}
}
```

### 2. WebSocket 客户端

```go
package main

import (
    "log"
    "time"
    
    "github.com/muchinfo/mtp2-common-lib/websocket"
)

func main() {
    // 创建客户端配置
    config := websocket.WSClientConfig{
        URL:            "ws://localhost:8080/ws", // 服务器地址
        ReconnectDelay: 5 * time.Second,         // 重连延迟
        MaxReconnects:  0,                       // 0表示无限重连
        AutoReconnect:  true,                    // 启用自动重连
        ReadBufferSize: 4096,                    // 读取缓冲区
        WriteBufferSize: 4096,                   // 写入缓冲区
    }
    
    // 创建客户端实例
    client := websocket.NewWSClient(config)
    defer client.Close()
    
    // 设置回调函数
    client.SetCallbacks(
        // 连接成功回调
        func() {
            log.Println("Connected to server")
            // 发送消息
            client.SendText("Hello from client!")
        },
        // 断开连接回调
        func(err error) {
            if err != nil {
                log.Printf("Disconnected with error: %v", err)
            } else {
                log.Println("Disconnected gracefully")
            }
        },
        // 消息接收回调
        func(data []byte) {
            message := string(data)
            log.Printf("Received message: %s", message)
        },
        // 错误回调
        func(err error) {
            log.Printf("Client error: %v", err)
        },
    )
    
    // 连接到服务器
    if err := client.Connect(); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    // 定时发送消息
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                if client.IsConnected() {
                    client.SendText("Heartbeat message")
                }
            }
        }
    }()
    
    // 保持运行
    select {}
}
```

## API 参考

### WSServer 配置

```go
type WSServerConfig struct {
    Address         string                           // 监听地址 (如 ":8080")
    Path            string                           // WebSocket路径 (如 "/ws")
    MaxConnections  int                              // 最大连接数，0表示无限制
    PingInterval    time.Duration                    // ping消息间隔，默认30秒
    PongWait        time.Duration                    // pong响应等待时间，默认60秒
    WriteWait       time.Duration                    // 写入超时时间，默认10秒
    ReadBufferSize  int                              // 读取缓冲区大小，默认4096
    WriteBufferSize int                              // 写入缓冲区大小，默认4096
    CheckOrigin     func(r *http.Request) bool       // 跨域检查函数
}
```

### WSClient 配置

```go
type WSClientConfig struct {
    URL             string        // WebSocket服务器URL
    ReconnectDelay  time.Duration // 重连延迟，默认5秒
    MaxReconnects   int           // 最大重连次数，0表示无限重连
    AutoReconnect   bool          // 是否启用自动重连
    ReadBufferSize  int           // 读取缓冲区大小，默认4096
    WriteBufferSize int           // 写入缓冲区大小，默认4096
    Headers         http.Header   // 连接时的HTTP头
}
```

### WSServer 方法

```go
// 创建服务器
func NewWSServer(config WSServerConfig) *WSServer

// 设置回调函数
func (s *WSServer) SetCallbacks(
    onClientConnect    func(*WSClientConnection),
    onClientDisconnect func(*WSClientConnection, error),
    onMessage          func(*WSClientConnection, []byte),
    onError            func(error),
)

// 服务器控制
func (s *WSServer) Start() error                    // 启动服务器
func (s *WSServer) Stop() error                     // 停止服务器
func (s *WSServer) IsRunning() bool                 // 检查运行状态
func (s *WSServer) GetAddress() string              // 获取监听地址

// 客户端管理
func (s *WSServer) GetClientCount() int             // 获取连接数
func (s *WSServer) GetClients() []*WSClientConnection // 获取所有客户端
func (s *WSServer) GetClient(id string) *WSClientConnection // 获取指定客户端

// 消息发送
func (s *WSServer) BroadcastText(message string) error      // 广播文本消息
func (s *WSServer) BroadcastJSON(data interface{}) error   // 广播JSON消息
func (s *WSServer) BroadcastBinary(data []byte) error      // 广播二进制消息
```

### WSClient 方法

```go
// 创建客户端
func NewWSClient(config WSClientConfig) *WSClient

// 设置回调函数
func (c *WSClient) SetCallbacks(
    onConnect    func(),
    onDisconnect func(error),
    onMessage    func([]byte),
    onError      func(error),
)

// 连接控制
func (c *WSClient) Connect() error           // 连接到服务器
func (c *WSClient) Disconnect()              // 断开连接
func (c *WSClient) Close()                   // 关闭客户端
func (c *WSClient) IsConnected() bool        // 检查连接状态

// 消息发送
func (c *WSClient) SendText(message string) error      // 发送文本消息
func (c *WSClient) SendJSON(data interface{}) error   // 发送JSON消息
func (c *WSClient) SendBinary(data []byte) error      // 发送二进制消息
```

### WSClientConnection 方法

```go
// 客户端信息
type WSClientConnection struct {
    ID          string        // 唯一连接ID
    RemoteAddr  string        // 远程地址
    ConnectedAt time.Time     // 连接时间
    UserAgent   string        // 用户代理
    Headers     http.Header   // HTTP头信息
}

// 消息发送
func (c *WSClientConnection) SendText(message string) error      // 发送文本消息
func (c *WSClientConnection) SendJSON(data interface{}) error   // 发送JSON消息
func (c *WSClientConnection) SendBinary(data []byte) error      // 发送二进制消息

// 连接信息
func (c *WSClientConnection) GetUptime() time.Duration          // 获取连接时长
func (c *WSClientConnection) Close() error                      // 关闭连接
```

## 高级用法

### 1. 消息类型处理

```go
// JSON消息示例
type ChatMessage struct {
    Type    string `json:"type"`
    User    string `json:"user"`
    Content string `json:"content"`
    Time    int64  `json:"timestamp"`
}

// 发送JSON消息
msg := ChatMessage{
    Type:    "chat",
    User:    "Alice",
    Content: "Hello everyone!",
    Time:    time.Now().Unix(),
}
client.SendJSON(msg)

// 接收JSON消息
client.SetCallbacks(nil, nil, func(data []byte) {
    var msg ChatMessage
    if err := json.Unmarshal(data, &msg); err == nil {
        log.Printf("[%s] %s: %s", msg.Type, msg.User, msg.Content)
    }
}, nil)
```

### 2. 选择性广播

```go
// 向特定客户端发送消息
clients := server.GetClients()
for _, client := range clients {
    if client.UserAgent == "特定用户代理" {
        client.SendText("特定消息")
    }
}

// 基于连接时间过滤
cutoff := time.Now().Add(-1 * time.Hour)
for _, client := range clients {
    if client.ConnectedAt.After(cutoff) {
        client.SendText("新用户消息")
    }
}
```

### 3. 连接限制和管理

```go
server.SetCallbacks(
    func(client *WSClientConnection) {
        // 检查连接数限制
        if server.GetClientCount() > 50 {
            client.SendText("服务器繁忙，请稍后再试")
            client.Close()
            return
        }
        
        // IP地址限制示例
        if strings.Contains(client.RemoteAddr, "192.168.") {
            log.Printf("内网用户连接: %s", client.RemoteAddr)
        }
        
        client.SendJSON(map[string]interface{}{
            "type": "welcome",
            "your_id": client.ID,
            "server_time": time.Now().Unix(),
        })
    },
    // ... 其他回调
)
```

### 4. 错误处理和监控

```go
client.SetCallbacks(
    func() {
        log.Println("✅ 连接成功")
    },
    func(err error) {
        if err != nil {
            log.Printf("❌ 连接断开: %v", err)
        }
    },
    func(data []byte) {
        log.Printf("📨 收到消息: %s", string(data))
    },
    func(err error) {
        log.Printf("⚠️ 客户端错误: %v", err)
        
        // 根据错误类型处理
        if strings.Contains(err.Error(), "connection refused") {
            log.Println("🔄 服务器不可达，将尝试重连...")
        }
    },
)
```

## 测试

运行WebSocket组件的测试：

```bash
# 运行所有测试
go test ./websocket -v

# 运行测试并显示覆盖率
go test ./websocket -v -cover

# 运行特定测试
go test ./websocket -v -run TestWSServer_StartStop
```

当前测试覆盖率: **69.7%**

测试包括：

- 服务器启动/停止测试
- 客户端连接/断开测试  
- 消息交换测试
- 广播功能测试
- 自动重连测试

## 性能特性

### 并发支持

- 服务器支持数百个并发WebSocket连接
- 使用goroutine处理每个客户端连接
- 读写锁保证并发安全

### 内存管理  

- 自动管理连接生命周期
- 及时清理断开的连接资源
- 可配置的缓冲区大小

### 网络优化

- 内置心跳机制检测连接状态
- 支持二进制和文本消息压缩
- 可配置的超时和重试策略

## 注意事项

1. **端口占用**: 确保配置的端口未被其他程序占用
2. **防火墙**: 确保WebSocket端口在防火墙中开放
3. **跨域**: 生产环境建议配置适当的跨域检查
4. **资源限制**: 根据系统资源合理设置最大连接数
5. **日志记录**: 建议在生产环境启用适当的日志记录

## 示例项目

完整的使用示例请参考：

- [WebSocket客户端示例](../example/websocket_client_example.go)
- [WebSocket服务器示例](../example/websocket_server_example.go)

运行示例：

```bash
# 运行服务器示例
go run example/. websocket_server

# 运行客户端示例  
go run example/. websocket_client
```
