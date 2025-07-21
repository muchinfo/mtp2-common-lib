# ulidgen 组件

用于生成唯一、可排序、长度可控的 ULID 字符串，适用于订单号、流水号等业务场景。

## 特点

- 基于 [oklog/ulid](https://github.com/oklog/ulid) 实现，唯一性强，26位标准长度。
- 支持自定义长度（最短18位），可带业务前缀，兼容第三方API长度要求。
- 可自定义随机源，便于测试。

## 安装依赖

```bash
go get github.com/oklog/ulid/v2
```

## 用法示例

```go
import "yourmodule/ulidgen"

// 生成标准 ULID
id, err := ulidgen.GenerateULID()
// 生成18位短ULID
id, err := ulidgen.GenerateShortULID(18)
// 生成带前缀且总长20位的ULID
id, err := ulidgen.GenerateULIDWithPrefix("JYI", 20)
```

## 接口说明

- `GenerateULID() (string, error)`：生成26位标准ULID。
- `GenerateShortULID(length int) (string, error)`：生成指定长度（18~26位）ULID。
- `GenerateULIDWithPrefix(prefix string, maxLen int) (string, error)`：生成带前缀且总长不超过maxLen的ULID。
- `GenerateULIDWithRandSource(t time.Time, r *rand.Rand) string`：自定义时间和随机源，便于测试。

## 测试

```bash
go test ./ulidgen
```

## 典型场景

- 订单号、流水号、分布式唯一ID
- 兼容外部API长度要求（如18~20位）

## 注意事项

- 截断ULID会略微降低唯一性，但18位已足够大多数业务场景。
- 如需更短ID，建议加业务前缀+时间戳+自增号等混合方案。
