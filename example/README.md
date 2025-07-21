# http 包示例

请参考 httpcall_example.go 了解 http 包的常用用法。

运行测试：

```bash
go test -v httpcall_test.go
```

## example 目录说明

本目录用于存放各模块的示例代码，便于快速体验和验证功能。

## 运行方式

你可以通过如下命令按需运行不同模块的示例：

```sh
# 只运行 Logger 示例
cd example
go run . logger

# 只运行 ULID 生成示例
go run . ulidgen

# 运行全部示例
go run . all
```

支持的模块有：

- viper
- logger
- rabbitmq
- xorm
- ulidgen

无参数或参数错误时会输出帮助信息。

## 说明

- 各模块的具体示例代码见本目录下对应的 `*_example.go` 文件。
- main.go 作为统一入口，便于管理和扩展。
