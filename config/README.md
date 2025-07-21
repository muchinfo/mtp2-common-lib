# config 组件说明

本目录基于 [spf13/viper](https://github.com/spf13/viper) 封装了配置文件加载与热更新的工具函数，支持多配置文件合并和变更回调。

## 主要功能

- 支持 YAML 配置文件加载
- 支持多个配置文件依次合并，后者覆盖前者
- 支持配置热更新，变更时自动回调

## 使用方法

### 1. 定义配置结构体

```go
// 定义你的配置结构体（示例）
 type AppConfig struct {
     Name string `yaml:"name"`
     Age  int    `yaml:"age"`
 }
```

### 2. 调用 InitViper 加载配置

```go
import (
    "github.com/fsnotify/fsnotify"
    "github.com/muchinfo/mtp2-common-lib/config"
)

func main() {
    var cfg AppConfig
    files := []string{"config.yaml", "override.yaml"}
    err := config.InitViper(files, &cfg, func(e fsnotify.Event) {
        // 配置变更时的回调，注意：cfg 会自动更新为最新配置
        fmt.Println("配置变更:", e.Name)
        fmt.Printf("新配置: %+v\n", cfg)
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("初始配置: %+v\n", cfg)
    // 注意：上面这行仅输出首次加载的配置，后续配置变更需依赖回调函数中的输出
}
```

### 3. 配置文件示例

`config.yaml`:

```yaml
name: Alice
age: 20
```

`override.yaml`:

```yaml
age: 30
```

最终合并后，`cfg.Name == "Alice"`，`cfg.Age == 30`。

## 测试

可参考 `viper_test.go` 进行单元测试，主要覆盖配置文件加载、合并、热更新和回调等功能：
可参考 `viper_test.go` 进行单元测试：

```shell
go test ./config
```

## 依赖

- [spf13/viper](https://github.com/spf13/viper)
- [fsnotify/fsnotify](https://github.com/fsnotify/fsnotify)
