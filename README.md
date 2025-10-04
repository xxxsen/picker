picker
===

一个基于 [yaegi](https://github.com/traefik/yaegi) 的轻量插件容器，支持在运行时按配置加载并执行 Go 函数。

目前内置 YAML / JSON / TOML 三种配置格式。推荐在 TOML 中使用 `'''` 多行字面量存放函数体，可避免 `\n` 这类转义在解析过程中被处理。

## Quick Start

更多示例见 `picker_test.go`。

```go
package main

import (
    "context"
    "log"

    "github.com/xxxsen/picker"
)

func main() {
    pgs := &picker.Plugins{
        Plugins: []*picker.PluginConfig{
            {
                Name:     "p1",
                Function: `func(ctx context.Context) {fmt.Println("hello world")}`,
            },
            {
                Name:     "p2",
                Import:   []string{"fmt"},
                Function: `func(ctx context.Context) {fmt.Println("hello again")}`,
            },
            {
                Name:   "p3",
                Define: "a := 3.14",
                Function: `func(ctx context.Context) {fmt.Printf("value a:%f\n", a)}`,
            },
        },
    }
    pk, err := picker.Load[func(context.Context)](pgs)
    if err != nil {
        log.Fatalf("load plugins failed: %v", err)
    }
    for _, name := range pk.List() {
        fn, _ := pk.Get(name)
        fn(context.Background())
    }
}
```

> 提示：泛型参数决定了插件函数签名，确保配置中的函数类型与之匹配即可；参数数量、类型都可以自行定义。
