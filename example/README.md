
# example 目录说明

本目录用于存放各模块的独立示例代码，便于快速体验和验证 mtp2-common-lib 的各项功能。

## 运行方式

你可以通过如下命令按需运行不同模块的示例：

```sh
# 进入 example 目录
cd example

# 只运行 Logger 示例（推荐外部项目用法，见 logger_example.go）
go run . logger

# 只运行 Viper 配置示例
go run . viper

# 只运行 RabbitMQ 示例
go run . rabbitmq

# 只运行 XORM Oracle 示例
go run . xorm

# 只运行 ULID 生成示例
go run . ulidgen

# 只运行 HTTP 工具示例
go run . http

# 只运行 Socket 客户端示例
go run . socket_client

# 只运行 Socket 服务器端示例
go run . socket_server

# 运行全部示例
go run . all
```

支持的模块有：

- viper      —— 配置加载与热更新
- logger     —— 高性能 zap 日志，支持 callerSkip
- rabbitmq   —— RabbitMQ 客户端
- xorm       —— Oracle 数据库 xorm 封装
- ulidgen    —— ULID 唯一ID生成
- http       —— HTTP 请求、签名、加解密
- socket_client     —— Socket 客户端
- socket_server     —— Socket 服务器端

无参数或参数错误时会输出帮助信息。

## 各模块示例说明

- `logger_example.go`：演示日志组件的自定义配置、callerSkip 用法，推荐外部项目使用 `InitWithCallerSkip(config, 2)`。
- `viper_example.go`：演示多文件配置加载、热更新。
- `rabbitmq_example.go`：演示 RabbitMQ 消息收发、断线重连。
- `xorm_oracle_example.go`：演示 Oracle 数据库连接、慢 SQL、熔断。
- `ulidgen_example.go`：演示 ULID 唯一ID生成。
- `httpcall_example.go`：演示 HTTP 请求、签名、加解密。
- `socket_client_example.go`：演示 Socket 客户端功能。
- `socket_server_example.go`：演示 Socket 服务器端功能。

## 测试

部分模块提供 go test 测试用例，例如：

```sh
go test -v httpcall_test.go
go test -v ../logger/zap_test.go
```

## 依赖

示例依赖于主库的 go.mod 统一管理，无需单独安装。

## 常见问题

- 日志调用栈定位不准确？请参考 logger_example.go，外部项目建议用 `InitWithCallerSkip(config, 2)`。
- 数据库/消息队列等需本地服务支持，建议先启动对应服务。

## 说明

- 各模块的具体示例代码见本目录下对应的 `*_example.go` 文件。
- main.go 作为统一入口，便于管理和扩展。
