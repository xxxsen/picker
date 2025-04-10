package picker

import "fmt"

type IContainer interface {
	Register(name string, val interface{}) error
	Get(name string) (interface{}, bool)
}

type containerImpl struct {
	m map[string]interface{}
}

func (c *containerImpl) Register(name string, val interface{}) error {
	if _, ok := c.m[name]; ok {
		return fmt.Errorf("plugin:%s exists", name)
	}
	c.m[name] = val
	return nil
}

func (c *containerImpl) Get(name string) (interface{}, bool) {
	val, ok := c.m[name]
	return val, ok
}

func NewContainer() IContainer {
	return &containerImpl{
		m: make(map[string]interface{}),
	}
}
