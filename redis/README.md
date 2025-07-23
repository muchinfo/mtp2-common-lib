# Redis 组件

本组件为调用者提供 Redis 常用操作功能，支持字符串、哈希、列表、集合、有序集合等数据类型的操作，以及发布订阅、事务、JSON序列化等高级功能。

## 功能特性

### 🔧 核心功能

- **连接管理**: 自动连接、重连机制、连接池管理
- **数据类型**: 支持所有Redis数据类型（字符串、哈希、列表、集合、有序集合）
- **高级操作**: 事务、管道、发布订阅、Lua脚本
- **JSON支持**: 内置JSON序列化和反序列化
- **错误处理**: 完善的错误处理和回调机制
- **性能监控**: 连接池统计和性能指标

### 📊 支持的操作

#### 字符串操作 (String)

- `Set/Get/Del`: 基础读写删除
- `MSet/MGet`: 批量操作
- `IncrBy/DecrBy`: 计数器操作
- `Expire/TTL`: 过期时间管理
- `GetSet`: 原子获取并设置

#### 哈希操作 (Hash)

- `HSet/HGet/HDel`: 字段操作
- `HMSet/HMGet`: 批量字段操作
- `HGetAll/HKeys/HVals`: 获取所有数据
- `HExists/HLen`: 检查和统计
- `HIncrBy`: 字段数值操作

#### 列表操作 (List)

- `LPush/RPush`: 头尾插入
- `LPop/RPop`: 头尾弹出
- `LRange`: 范围获取
- `LTrim`: 列表裁剪
- `LLen`: 长度统计

#### 集合操作 (Set)

- `SAdd/SRem`: 添加删除成员
- `SMembers/SCard`: 获取成员和数量
- `SIsMember`: 成员检查
- `SRandMember`: 随机成员
- `SUnion/SInter/SDiff`: 集合运算

#### 有序集合操作 (Sorted Set)

- `ZAdd/ZRem`: 添加删除成员
- `ZRange/ZRangeWithScores`: 范围查询
- `ZRank/ZScore`: 排名和分数
- `ZCard/ZCount`: 统计操作
- `ZIncrBy`: 分数增减

#### 键管理操作 (Keys)

- `Exists/Type`: 键检查和类型
- `Keys/Scan`: 模式匹配查找
- `Rename/Del`: 重命名删除
- `Expire/TTL`: 过期管理

## 使用方法

### 基本使用

```go
package main

import (
    "time"
    "github.com/muchinfo/mtp2-common-lib/redis"
)

func main() {
    // 创建配置
    config := redis.RedisConfig{
        Address:         "localhost:6379",
        Password:        "",
        Database:        0,
        PoolSize:        10,
        MinIdleConns:    2,
        DialTimeout:     5 * time.Second,
        ReadTimeout:     3 * time.Second,
        WriteTimeout:    3 * time.Second,
        MaxRetries:      3,
        PoolTimeout:     4 * time.Second,
        IdleTimeout:     5 * time.Minute,
    }

    // 创建客户端
    client, err := redis.NewRedisClient(config)
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // 设置回调函数
    client.SetCallbacks(
        func(err error) {
            log.Printf("Redis Error: %v", err)
        },
        func() {
            log.Println("Redis Reconnected")
        },
    )

    // 基础操作
    err = client.Set("key", "value", time.Hour)
    if err != nil {
        panic(err)
    }

    value, err := client.Get("key")
    if err != nil {
        panic(err)
    }
    fmt.Println("Value:", value)
}
```

### 字符串操作示例

```go
// 设置键值对
err := client.Set("user:name", "Alice", 10*time.Minute)

// 获取值
name, err := client.Get("user:name")

// 计数器操作
views, err := client.IncrBy("page:views", 1)

// 批量操作
err = client.MSet("product:1", "Laptop", "product:2", "Phone")
products, err := client.MGet("product:1", "product:2")
```

### 哈希操作示例

```go
// 设置用户信息
userInfo := map[string]interface{}{
    "name":  "Bob",
    "email": "bob@example.com",
    "age":   30,
}
err := client.HMSet("user:123", userInfo)

// 获取特定字段
email, err := client.HGet("user:123", "email")

// 获取所有字段
allFields, err := client.HGetAll("user:123")
```

### 列表操作示例

```go
// 消息队列 (FIFO)
client.RPush("queue", "message1", "message2", "message3")
message, err := client.LPop("queue")

// 最新列表 (最新的在前面)
client.LPush("recent", "item1", "item2")
recent, err := client.LRange("recent", 0, 9)
```

### 集合操作示例

```go
// 添加标签
client.SAdd("article:tags", "golang", "redis", "database")

// 检查成员
exists, err := client.SIsMember("article:tags", "golang")

// 获取所有成员
tags, err := client.SMembers("article:tags")
```

### 有序集合操作示例

```go
// 排行榜
players := []*redisLib.Z{
    {Score: 1500, Member: "player1"},
    {Score: 2000, Member: "player2"},
}
client.ZAdd("leaderboard", players...)

// 获取前3名
top3, err := client.ZRangeWithScores("leaderboard", -3, -1)
```

### JSON操作示例

```go
// 存储结构体
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

user := User{ID: 1, Name: "Alice", Age: 25}
err := client.SetJSON("user:1", user, time.Hour)

// 读取结构体
var retrievedUser User
err = client.GetJSON("user:1", &retrievedUser)
```

### 发布订阅示例

```go
// 订阅频道
pubsub := client.Subscribe("notifications")
defer pubsub.Close()

// 发布消息
go func() {
    client.Publish("notifications", "Hello World!")
}()

// 接收消息
for {
    msg, err := pubsub.ReceiveMessage(client.GetContext())
    if err != nil {
        break
    }
    fmt.Println("Received:", msg.Payload)
}
```

### 事务操作示例

```go
// 创建管道事务
pipe := client.TxPipeline()

// 添加操作
pipe.Set(ctx, "key1", "value1", 0)
pipe.IncrBy(ctx, "counter", 1)
pipe.HSet(ctx, "hash", "field", "value")

// 执行事务
results, err := pipe.Exec(ctx)
```

## 配置选项

### RedisConfig 结构

```go
type RedisConfig struct {
    // 连接配置
    Address  string        // Redis服务器地址 (host:port)
    Password string        // 密码，如果没有则为空
    Database int           // 数据库索引 (0-15)
    
    // 连接池配置
    PoolSize        int           // 连接池大小
    MinIdleConns    int           // 最小空闲连接数
    DialTimeout     time.Duration // 连接超时时间
    ReadTimeout     time.Duration // 读取超时时间
    WriteTimeout    time.Duration // 写入超时时间
    MaxRetries      int           // 最大重试次数
    PoolTimeout     time.Duration // 获取连接超时时间
    IdleTimeout     time.Duration // 空闲连接超时时间
    
    // TLS配置 (可选)
    TLSConfig *tls.Config
}
```

### 推荐配置

```go
// 开发环境
config := redis.RedisConfig{
    Address:         "localhost:6379",
    Database:        0,
    PoolSize:        5,
    MinIdleConns:    1,
    DialTimeout:     5 * time.Second,
    ReadTimeout:     3 * time.Second,
    WriteTimeout:    3 * time.Second,
}

// 生产环境
config := redis.RedisConfig{
    Address:         "redis.example.com:6379",
    Password:        "your-password",
    Database:        0,
    PoolSize:        20,
    MinIdleConns:    5,
    DialTimeout:     10 * time.Second,
    ReadTimeout:     5 * time.Second,
    WriteTimeout:    5 * time.Second,
    MaxRetries:      3,
    PoolTimeout:     10 * time.Second,
    IdleTimeout:     10 * time.Minute,
}
```

## 错误处理

### 回调函数

```go
client.SetCallbacks(
    func(err error) {
        // 错误回调 - 记录日志，发送告警等
        log.Printf("Redis Error: %v", err)
        // 可以在这里实现自定义错误处理逻辑
    },
    func() {
        // 重连回调 - 记录重连事件
        log.Println("Redis Reconnected")
        // 可以在这里实现重连后的逻辑
    },
)
```

### 常见错误处理

```go
value, err := client.Get("key")
if err != nil {
    if err == redis.Nil {
        // 键不存在
        fmt.Println("Key not found")
    } else {
        // 其他错误
        log.Printf("Redis error: %v", err)
    }
}
```

## 性能优化

### 连接池优化

```go
// 根据应用负载调整连接池大小
config.PoolSize = 20        // 并发连接数
config.MinIdleConns = 5     // 预热连接
config.PoolTimeout = 10 * time.Second // 获取连接超时
```

### 批量操作

```go
// 使用批量操作提高性能
// 避免：循环中多次调用Set
for i := 0; i < 1000; i++ {
    client.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), 0)
}

// 推荐：使用MSet批量操作
args := make([]interface{}, 2000)
for i := 0; i < 1000; i++ {
    args[i*2] = fmt.Sprintf("key%d", i)
    args[i*2+1] = fmt.Sprintf("value%d", i)
}
client.MSet(args...)
```

### 管道操作

```go
// 使用管道减少网络往返
pipe := client.TxPipeline()
for i := 0; i < 100; i++ {
    pipe.Set(ctx, fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), 0)
}
pipe.Exec(ctx)
```

## 最佳实践

### 1. 键命名规范

```go
// 使用分隔符组织键名
"user:123:profile"
"session:abc123:data"
"cache:product:456"
```

### 2. 过期时间管理

```go
// 为缓存数据设置合适的过期时间
client.Set("cache:user:123", userData, 30*time.Minute)
client.Set("session:token", sessionData, 24*time.Hour)
```

### 3. 内存优化

```go
// 使用哈希存储相关字段
client.HMSet("user:123", map[string]interface{}{
    "name": "Alice",
    "email": "alice@example.com",
    "age": 25,
})

// 而不是分散的键
client.Set("user:123:name", "Alice", 0)
client.Set("user:123:email", "alice@example.com", 0)
client.Set("user:123:age", "25", 0)
```

### 4. 错误恢复

```go
// 实现重试逻辑
func setWithRetry(client *redis.RedisClient, key, value string, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := client.Set(key, value, time.Hour)
        if err == nil {
            return nil
        }
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    return errors.New("max retries exceeded")
}
```

## 监控和调试

### 连接池监控

```go
stats := client.PoolStats()
fmt.Printf("Total Connections: %d\n", stats.TotalConns)
fmt.Printf("Idle Connections: %d\n", stats.IdleConns)
fmt.Printf("Stale Connections: %d\n", stats.StaleConns)
```

### 健康检查

```go
// 定期健康检查
func healthCheck(client *redis.RedisClient) bool {
    err := client.Ping()
    return err == nil
}
```

## 依赖

- `github.com/redis/go-redis/v9` - Redis Go客户端
- `encoding/json` - JSON序列化支持

## 注意事项

1. **连接管理**: 确保正确关闭客户端连接
2. **错误处理**: 始终检查返回的错误
3. **过期时间**: 为缓存数据设置合适的TTL
4. **内存使用**: 监控Redis内存使用情况
5. **网络延迟**: 考虑网络延迟对性能的影响
6. **数据序列化**: JSON操作会增加CPU开销
7. **连接池**: 根据应用负载调整连接池大小
8. **事务**: 注意事务操作的原子性保证
