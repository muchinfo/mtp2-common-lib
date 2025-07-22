# Socket TCP 网络通信组件

这是一个功能完整的TCP网络通信组件，提供TCP客户端和TCP服务器的完整实现，支持连接管理、自动重连、消息广播、回调处理等企业级功能。

## 📋 目录

- [功能概览] (#功能概览)
- [快速开始] (#快速开始)
- [API参考] (#api参考)
- [配置说明] (#配置说明)
- [使用示例] (#使用示例)
- [测试] (#测试)
- [最佳实践] (#最佳实践)

## 🚀 功能概览

### TCP客户端 (TCPClient)

- ✅ **连接管理** - 支持连接、断开连接、状态查询
- ✅ **自动重连** - 连接断开时自动尝试重连，支持重连次数限制
- ✅ **超时控制** - 可配置读取和写入超时时间
- ✅ **消息发送** - 支持发送字节数据和字符串消息
- ✅ **回调机制** - 提供连接、断开、消息接收、错误等事件回调
- ✅ **线程安全** - 使用读写锁保证并发操作安全
- ✅ **优雅关闭** - 支持优雅关闭和资源清理

### TCP服务器 (TCPServer)

- ✅ **服务监听** - 支持启动、停止TCP服务器
- ✅ **客户端管理** - 自动管理多个客户端连接，分配唯一ID
- ✅ **消息处理** - 异步处理客户端消息
- ✅ **消息广播** - 支持向所有客户端或指定客户端发送消息
- ✅ **连接限制** - 支持最大连接数限制
- ✅ **超时控制** - 可配置客户端读取和写入超时
- ✅ **回调机制** - 提供客户端连接、断开、消息接收、错误等事件回调
- ✅ **状态查询** - 可查询服务器状态、客户端列表等信息
- ✅ **线程安全** - 支持并发客户端连接处理

### 客户端连接 (ClientConnection)

- ✅ **连接信息** - 提供连接ID、远程地址、连接时间等信息
- ✅ **消息发送** - 支持向单个客户端发送消息
- ✅ **连接管理** - 支持关闭连接、查询连接状态
- ✅ **运行时统计** - 提供连接持续时间等统计信息

## 🎯 快速开始

### 1. TCP客户端基本用法

```go
package main

import (
    "log"
    "time"
    "github.com/muchinfo/mtp2-common-lib/socket"
)

func main() {
    // 创建客户端配置
    config := socket.TCPClientConfig{
        Address:        "localhost:8080",
        ReconnectDelay: 5 * time.Second,
        ReadTimeout:    30 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxReconnects:  0,    // 0表示无限重连
        AutoReconnect:  true, // 启用自动重连
    }

    // 创建TCP客户端
    client := socket.NewTCPClient(config)
    defer client.Close()

    // 设置回调函数
    client.SetCallbacks(
        func() {
            log.Println("✅ 连接成功")
        },
        func(err error) {
            log.Printf("❌ 连接断开: %v", err)
        },
        func(data []byte) {
            log.Printf("📨 收到消息: %s", string(data))
        },
        func(err error) {
            log.Printf("⚠️ 错误: %v", err)
        },
    )

    // 连接到服务器
    if err := client.Connect(); err != nil {
        log.Fatalf("连接失败: %v", err)
    }

    // 发送消息
    if err := client.SendString("Hello, Server!"); err != nil {
        log.Printf("发送失败: %v", err)
    }

    // 保持连接...
    select {}
}
```

### 2. TCP服务器基本用法

```go
package main

import (
    "log"
    "time"
    "github.com/muchinfo/mtp2-common-lib/socket"
)

func main() {
    // 创建服务器配置
    config := socket.TCPServerConfig{
        Address:        ":8080",          // 监听端口8080
        ReadTimeout:    30 * time.Second, // 读取超时
        WriteTimeout:   10 * time.Second, // 写入超时
        MaxConnections: 100,              // 最大连接数
    }

    // 创建TCP服务器
    server := socket.NewTCPServer(config)
    defer server.Stop()

    // 设置回调函数
    server.SetCallbacks(
        func(client *socket.ClientConnection) {
            log.Printf("✅ 客户端连接: %s", client.RemoteAddr)
            // 发送欢迎消息
            client.SendString("欢迎连接到服务器!")
        },
        func(client *socket.ClientConnection, err error) {
            log.Printf("❌ 客户端断开: %s", client.RemoteAddr)
        },
        func(client *socket.ClientConnection, data []byte) {
            message := string(data)
            log.Printf("📨 收到消息: %s", message)
            // 广播消息给所有客户端
            server.BroadcastString(fmt.Sprintf("[%s]: %s", client.RemoteAddr, message))
        },
        func(err error) {
            log.Printf("⚠️ 服务器错误: %v", err)
        },
    )

    // 启动服务器
    if err := server.Start(); err != nil {
        log.Fatalf("启动服务器失败: %v", err)
    }

    log.Printf("🚀 服务器启动成功，监听地址: %s", server.GetAddress())

    // 保持服务器运行...
    select {}
}
```

## 📚 API参考

### TCP客户端配置 (TCPClientConfig)

```go
type TCPClientConfig struct {
    Address        string        // 服务器地址，格式：host:port
    ReconnectDelay time.Duration // 重连延迟时间，默认5秒
    ReadTimeout    time.Duration // 读取超时时间，默认30秒
    WriteTimeout   time.Duration // 写入超时时间，默认10秒
    MaxReconnects  int           // 最大重连次数，0表示无限重连
    AutoReconnect  bool          // 是否启用自动重连，默认true
}
```

### TCP服务器配置 (TCPServerConfig)

```go
type TCPServerConfig struct {
    Address        string        // 监听地址，格式：:port 或 host:port
    ReadTimeout    time.Duration // 客户端读取超时，默认30秒
    WriteTimeout   time.Duration // 客户端写入超时，默认10秒
    MaxConnections int           // 最大客户端连接数，0表示无限制
}
```

### TCP客户端主要方法

| 方法 | 描述 |
|------|------|
| `NewTCPClient(config)` | 创建新的TCP客户端实例 |
| `Connect()` | 连接到服务器 |
| `Disconnect()` | 断开与服务器的连接 |
| `Close()` | 关闭客户端，清理所有资源 |
| `Send(data []byte)` | 发送字节数据到服务器 |
| `SendString(message)` | 发送字符串消息到服务器 |
| `IsConnected()` | 检查当前连接状态 |
| `GetAddress()` | 获取服务器地址 |
| `GetReconnectCount()` | 获取当前重连次数 |
| `SetCallbacks(...)` | 设置事件回调函数 |

### TCP服务器主要方法

| 方法 | 描述 |
|------|------|
| `NewTCPServer(config)` | 创建新的TCP服务器实例 |
| `Start()` | 启动服务器，开始监听连接 |
| `Stop()` | 停止服务器，关闭所有连接 |
| `IsRunning()` | 检查服务器是否正在运行 |
| `GetAddress()` | 获取服务器监听地址 |
| `GetClientCount()` | 获取当前客户端连接数 |
| `GetClients()` | 获取所有客户端连接列表 |
| `GetClient(id)` | 根据ID获取指定的客户端连接 |
| `Broadcast(data)` | 向所有客户端广播字节数据 |
| `BroadcastString(message)` | 向所有客户端广播字符串消息 |
| `SetCallbacks(...)` | 设置事件回调函数 |

### 客户端连接主要方法

| 方法 | 描述 |
|------|------|
| `Send(data []byte)` | 向客户端发送字节数据 |
| `SendString(message)` | 向客户端发送字符串消息 |
| `Close()` | 关闭客户端连接 |
| `IsClosed()` | 检查连接是否已关闭 |
| `GetUptime()` | 获取连接持续时间 |

### 回调函数

#### 客户端回调

```go
client.SetCallbacks(
    onConnect func(),                    // 连接成功回调
    onDisconnect func(error),           // 断开连接回调
    onMessage func([]byte),             // 消息接收回调
    onError func(error),                // 错误回调
)
```

#### 服务器回调

```go
server.SetCallbacks(
    onClientConnect func(*ClientConnection),           // 客户端连接回调
    onClientDisconnect func(*ClientConnection, error), // 客户端断开回调
    onMessage func(*ClientConnection, []byte),         // 消息接收回调
    onError func(error),                              // 错误回调
)
```

## ⚙️ 配置说明

### 重连配置

- **AutoReconnect**: 启用/禁用自动重连功能
- **ReconnectDelay**: 重连尝试之间的延迟时间
- **MaxReconnects**: 最大重连次数（0表示无限重连）

### 超时配置

- **ReadTimeout**: 读取操作的超时时间，防止读取阻塞
- **WriteTimeout**: 写入操作的超时时间，防止写入阻塞

### 连接限制

- **MaxConnections**: 服务器允许的最大并发连接数

## 📖 使用示例

### 聊天服务器示例

```go
// 启动聊天服务器
server := socket.NewTCPServer(socket.TCPServerConfig{
    Address: ":8080",
    MaxConnections: 50,
})

server.SetCallbacks(
    func(client *socket.ClientConnection) {
        // 通知所有用户有新用户加入
        server.BroadcastString(fmt.Sprintf("用户 %s 加入了聊天室", client.RemoteAddr))
    },
    func(client *socket.ClientConnection, err error) {
        // 通知所有用户有用户离开
        server.BroadcastString(fmt.Sprintf("用户 %s 离开了聊天室", client.RemoteAddr))
    },
    func(client *socket.ClientConnection, data []byte) {
        message := string(data)
        // 广播聊天消息
        server.BroadcastString(fmt.Sprintf("[%s]: %s", client.RemoteAddr, message))
    },
    nil,
)
```

### 回显服务器示例

```go
// 启动回显服务器
server := socket.NewTCPServer(socket.TCPServerConfig{
    Address: ":8080",
})

server.SetCallbacks(
    nil,
    nil,
    func(client *socket.ClientConnection, data []byte) {
        // 回显收到的消息
        client.SendString("Echo: " + string(data))
    },
    nil,
)
```

### 客户端重连示例

```go
// 创建支持自动重连的客户端
client := socket.NewTCPClient(socket.TCPClientConfig{
    Address:        "localhost:8080",
    AutoReconnect:  true,
    ReconnectDelay: 3 * time.Second,
    MaxReconnects:  10,
})

client.SetCallbacks(
    func() {
        log.Printf("重连成功，当前重连次数: %d", client.GetReconnectCount())
    },
    func(err error) {
        log.Printf("连接断开: %v", err)
    },
    nil,
    nil,
)
```

## 🧪 测试

项目包含完整的测试套件：

```bash
# 运行所有测试
go test ./socket -v

# 运行测试并显示覆盖率
go test ./socket -cover

# 运行集成测试
go test ./socket -v -run Integration

# 运行客户端测试
go test ./socket -v -run Client

# 运行服务器测试
go test ./socket -v -run Server
```

### 测试覆盖

- **单元测试**: 客户端和服务器的所有主要功能
- **集成测试**: 客户端与服务器的协同工作
- **并发测试**: 多客户端连接和消息广播
- **错误处理测试**: 各种错误场景的处理
- **代码覆盖率**: 85%+

## 💡 最佳实践

### 1. 错误处理

```go
// 总是检查错误返回值
if err := client.Connect(); err != nil {
    log.Printf("连接失败: %v", err)
    return
}

// 设置错误回调处理异步错误
client.SetCallbacks(nil, nil, nil, func(err error) {
    log.Printf("客户端错误: %v", err)
})
```

### 2. 资源管理

```go
// 使用defer确保资源被正确释放
client := socket.NewTCPClient(config)
defer client.Close()

server := socket.NewTCPServer(config)
defer server.Stop()
```

### 3. 超时设置

```go
// 根据网络环境设置合适的超时时间
config := socket.TCPClientConfig{
    ReadTimeout:  30 * time.Second, // 读取超时
    WriteTimeout: 10 * time.Second, // 写入超时
}
```

### 4. 重连策略

```go
// 为客户端设置合理的重连策略
config := socket.TCPClientConfig{
    AutoReconnect:  true,
    ReconnectDelay: 5 * time.Second,  // 重连延迟
    MaxReconnects:  10,               // 限制重连次数防止无限重连
}
```

### 5. 服务器容量规划

```go
// 根据服务器资源设置连接限制
config := socket.TCPServerConfig{
    MaxConnections: 1000, // 限制最大连接数
}
```

### 6. 消息协议

```go
// 建议使用结构化的消息格式
type Message struct {
    Type    string `json:"type"`
    Content string `json:"content"`
    From    string `json:"from"`
}

// 序列化后发送
data, _ := json.Marshal(message)
client.Send(data)
```
