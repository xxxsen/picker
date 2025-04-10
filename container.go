package picker

import "fmt"

type IContainer interface {
	Register(name string, val interface{}) error
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

func asContainer(m map[string]interface{}) IContainer {
	return &containerImpl{
		m: m,
	}
}
