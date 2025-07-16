# Oracle 数据库操作（xorm 版）

本组件基于 [xorm.io/xorm](https://xorm.io/) 和 [github.com/godror/godror](https://github.com/godror/godror) 实现 Oracle 数据库的连接与常用 CRUD 操作。

## 主要特性

- 支持 Oracle 连接与自动建表
- 提供简单的实体定义与 CRUD 示例
- 适合 Go 业务系统快速集成

## 快速开始

### 1. 安装依赖

```shell
go get xorm.io/xorm
go get github.com/godror/godror
```

### 2. 连接数据库

本组件基于 [xorm.io/xorm](https://xorm.io/) 和 [github.com/godror/godror](https://github.com/godror/godror) 实现 Oracle 数据库的连接与企业级特性。

## 进阶特性

- 支持 Oracle 连接与自动建表
- zap.Logger 日志注入，SQL/慢SQL/错误统一输出
- 连接池、健康检查、断线重连
- 慢 SQL 统计（可自定义阈值，超时自动告警）
- 熔断机制，支持自动监控数据库健康
- 适合 Go 业务系统快速集成

```go
database.AutoMigrate(engine)
```

### 4. CRUD 示例

go get go.uber.org/zap

```go
user := &database.User{Name: "张三", Age: 20}
engine.Insert(user)

var got database.User
engine.ID(user.Id).Get(&got)

user.Age = 21
engine.ID(user.Id).Update(user)

engine.ID(user.Id).Delete(new(database.User))
```

### 5. 示例

见 `example/xorm_oracle_example.go`

### 6. 单元测试

```shell
go test ./database
```

## 依赖

- [xorm.io/xorm](https://xorm.io/)
- [github.com/godror/godror](https://github.com/godror/godror)
