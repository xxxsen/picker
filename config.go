package picker

type config struct {
	objs     map[string]interface{}
	safeWrap bool
}

type Option func(c *config)

// WithCustomObject 导入自定义结构, 用于函数运行时访问
func WithCustomObject(m map[string]interface{}) Option {
	return func(c *config) {
		c.objs = m
	}
}

// WithSafeFuncWrap 将函数进行重新包装, 运行时自动catch panic, 基于反射实现, 应该会比较慢
// 但是, 你都已经动态执行了...
func WithSafeFuncWrap(v bool) Option {
	return func(c *config) {
		c.safeWrap = v
	}
}

func applyOpts(opts ...Option) *config {
	c := &config{
		objs:     make(map[string]interface{}),
		safeWrap: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
