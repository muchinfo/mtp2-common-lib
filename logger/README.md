# Logger 日志组件

基于 `go.uber.org/zap` 的高性能日志组件，支持日志轮转、结构化日志和多种输出格式。

## 功能特性

- 🚀 高性能日志记录 (基于 zap)
- 📁 自动日志轮转和清理
- 🔧 灵活的配置选项
- 📝 支持结构化日志和糖化日志
- 🎯 同时输出到文件和控制台
- ⚡ 开箱即用的预设配置

## 快速开始

### 1. 基本使用

```go
package main

import (
    "mtp2-common-lib/logger"
    "go.uber.org/zap"
)

func main() {
    // 使用默认配置初始化
    if err := logger.Init(nil); err != nil {
        panic(err)
    }
    defer logger.Close()

    // 记录日志
    logger.Info("应用启动", zap.String("version", "1.0.0"))
    logger.Infof("用户 %s 登录成功", "张三")
}
```

### 2. 自定义配置

```go
config := &logger.Config{
    Level:      "debug",           // 日志级别
    Format:     "json",            // 输出格式: json 或 console
    OutputPath: "./logs/app.log",  // 日志文件路径
    MaxAge:     7,                 // 保留天数
    Rotation:   24,                // 轮转间隔(小时)
}

if err := logger.Init(config); err != nil {
    panic(err)
}
defer logger.Close()
```

### 3. 快速初始化

```go
// 开发环境
if err := logger.InitDevelopment(); err != nil {
    panic(err)
}

// 生产环境
if err := logger.InitProduction(); err != nil {
    panic(err)
}
```

## API 说明

### 配置选项

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Level | string | "info" | 日志级别: debug, info, warn, error |
| Format | string | "json" | 输出格式: json, console |
| OutputPath | string | "./logs/app.log" | 日志文件路径 |
| MaxAge | int | 7 | 日志保留天数 |
| Rotation | int | 24 | 日志轮转间隔(小时) |

### 结构化日志

使用 `zap.Field` 记录结构化数据：

```go
logger.Info("用户操作",
    zap.String("user_id", "12345"),
    zap.String("action", "login"),
    zap.Duration("latency", time.Millisecond*150),
    zap.Bool("success", true),
)
```

### 糖化日志

使用类似 `printf` 的格式：

```go
logger.Infof("用户 %s 在 %s 执行了 %s 操作", "张三", "2024-01-01", "登录")
logger.Debugf("处理请求耗时: %dms", 150)
logger.Errorf("连接数据库失败: %v", err)
```

### 全局实例

直接使用全局 Logger 实例：

```go
// 原生 zap.Logger
logger.Logger.Info("消息", zap.String("key", "value"))

// 糖化 Logger
logger.SugarLogger.Infow("消息", "key", "value")
```

## 日志级别

- **debug**: 调试信息
- **info**: 一般信息  
- **warn**: 警告信息
- **error**: 错误信息
- **fatal**: 致命错误 (会导致程序退出)

## 日志轮转

日志文件会根据配置自动轮转：

- 按时间轮转 (默认24小时)
- 自动清理过期日志 (默认保留7天)
- 轮转文件命名格式: `app.log.2024010115`
- 当前日志文件: `app.log` (软链接)

## 输出示例

### JSON 格式 (生产环境推荐)

```json
{
  "level": "info",
  "ts": "2024-01-01 15:04:05",
  "caller": "main/main.go:15",
  "msg": "用户操作",
  "user_id": "12345",
  "action": "login",
  "success": true
}
```

### Console 格式 (开发环境推荐)

```text
2024-01-01 15:04:05  INFO  main/main.go:15  用户操作  {"user_id": "12345", "action": "login", "success": true}
```

## 最佳实践

1. **使用结构化日志**: 便于日志分析和监控
2. **合理设置日志级别**: 生产环境建议使用 info 级别
3. **及时关闭日志**: 在程序退出时调用 `logger.Close()`
4. **避免敏感信息**: 不要记录密码、密钥等敏感数据
5. **统一错误处理**: 在记录错误时提供足够的上下文信息

## 性能优化

- zap 是高性能日志库，比标准库快 4-10 倍
- 使用结构化日志避免字符串拼接
- 在高频场景下考虑使用异步日志
- 合理设置日志轮转和清理策略

## 依赖

- `go.uber.org/zap`: 高性能日志库
- `github.com/lestrrat/go-file-rotatelogs`: 日志轮转
