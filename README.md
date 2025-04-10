picker
===

emmm, 包装了下yaegi, 方便动态执行golang函数...

## 使用方式

```golang
func main() {
	pgs := &Plugins{
		Plugins: []*PluginConfig{
			{
				Name:     "p1", //单纯执行代码
				Function: `func(ctx context.Context) {fmt.Printf("hello world\n")}`,
			},
			{
				Name:     "p2",
                Import:   "fmt", // 导入一个自定义包
				Function: `func(ctx context.Context) {fmt.Printf("test\n")}`,
			},
			{
				Name:     "p3",
				Define:   "a := 3.14", //定义相关的变量, 之后可以被Function引用, 通常用于代码执行前的初始化
				Function: `func(ctx context.Context) {fmt.Printf("r u ok? value a:%f\n", a)}`,
			},
		},
	}
	pk, _ := Load[func(ctx context.Context)](pgs)
	for _, name := range pk.List() { //按插件定义的顺序遍历插件列表
		fn, _ := pk.Get(name) //从容器获取插件func并进行执行
		fn(context.Background())
	}
}
```

**NOTE: 函数签名没有限制一定是几个参数, 可以根据需要自定义, 但是需要确保函数签名跟泛型参数的类型是一致的。**

