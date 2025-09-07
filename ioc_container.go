package inject

import (
	"time"
)

// Container IoC容器
type Container struct {
	graph Graph
}

// NewContainer 创建一个新的IoC容器
func NewContainer() *Container {
	return &Container{
		graph: Graph{},
	}
}

// Provides 使用默认名称提供一些bean
func (c *Container) Provides(beans ...interface{}) error {
	for _, bean := range beans {
		if err := c.graph.Provide(&Object{Value: bean}); err != nil {
			return err
		}
	}
	return nil
}

// ProvideWithName 使用指定名称提供bean
func (c *Container) ProvideWithName(name string, bean interface{}) error {
	return c.graph.Provide(&Object{Name: name, Value: bean})
}

// Populate 为所有bean填充依赖字段。
// 此函数必须在提供所有bean后调用
func (c *Container) Populate() error {
	start := time.Now()
	defer func() {
		if c.graph.Logger != nil {
			c.graph.Logger.Info("populate the bean container toke time %s", time.Now().Sub(start))
		}
	}()
	return c.graph.Populate()
}
