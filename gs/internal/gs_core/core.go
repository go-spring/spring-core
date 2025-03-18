/*
 * Copyright 2024 The Go-Spring Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gs_core

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_core/resolving"
	"github.com/go-spring/spring-core/gs/internal/gs_core/wiring"
	"github.com/go-spring/spring-core/util/syslog"
)

type refreshState int

const (
	RefreshDefault = refreshState(iota) // 未刷新
	RefreshInit                         // 准备刷新
	Refreshing                          // 正在刷新
	Refreshed                           // 已刷新
)

// Container 是 go-spring 框架的基石，实现了 Martin Fowler 在 << Inversion
// of Control Containers and the Dependency Injection pattern >> 一文中
// 提及的依赖注入的概念。但原文的依赖注入仅仅是指对象之间的依赖关系处理，而有些 IoC
// 容器在实现时比如 Spring 还引入了对属性 property 的处理。通常大家会用依赖注入统
// 述上面两种概念，但实际上使用属性绑定来描述对 property 的处理会更加合适，因此
// go-spring 严格区分了这两种概念，在描述对 bean 的处理时要么单独使用依赖注入或属
// 性绑定，要么同时使用依赖注入和属性绑定。
type Container struct {
	state     refreshState
	resolving *resolving.Resolving
	wiring    *wiring.Wiring
}

// New 创建 IoC 容器。
func New() *Container {
	return &Container{
		resolving: resolving.New(),
		wiring:    wiring.New(),
	}
}

// Mock mocks the bean with the given object.
func (c *Container) Mock(obj interface{}, target gs.BeanSelector) {
	c.resolving.Mock(obj, target)
}

// Object 注册对象形式的 bean ，需要注意的是该方法在注入开始后就不能再调用了。
func (c *Container) Object(i interface{}) *gs.RegisteredBean {
	b := NewBean(reflect.ValueOf(i))
	return c.Register(b)
}

// Provide 注册构造函数形式的 bean ，需要注意的是该方法在注入开始后就不能再调用了。
func (c *Container) Provide(ctor interface{}, args ...gs.Arg) *gs.RegisteredBean {
	b := NewBean(ctor, args...)
	return c.Register(b)
}

func (c *Container) Register(b *gs.BeanDefinition) *gs.RegisteredBean {
	x := b.BeanRegistration().(*gs_bean.BeanDefinition)
	r := gs.NewRegisteredBean(b.BeanRegistration())
	if c.state < Refreshing {
		c.resolving.Register(x)
	}
	return r
}

func (c *Container) GroupRegister(fn resolving.GroupFunc) {
	c.resolving.GroupRegister(fn)
}

// RefreshProperties updates the properties of the container.
func (c *Container) RefreshProperties(p conf.ReadOnlyProperties) error {
	return c.wiring.RefreshProperties(p)
}

// Refresh initializes and wires all beans in the container.
func (c *Container) Refresh() (err error) {
	if c.state != RefreshDefault {
		return errors.New("container is refreshing or refreshed")
	}
	c.state = RefreshInit
	start := time.Now()

	err = c.resolving.RefreshInit(c.wiring.Properties())
	if err != nil {
		return err
	}

	c.state = Refreshing

	beans, err := c.resolving.Refresh(c.wiring.Properties())
	if err != nil {
		return err
	}

	err = c.wiring.Refresh(beans)
	if err != nil {
		return err
	}

	c.resolving = nil
	c.state = Refreshed
	syslog.Debugf("container is refreshed successfully, %d beans cost %v",
		len(beans), time.Now().Sub(start))
	return nil
}

// Wire wires the bean with the given object.
func (c *Container) Wire(obj interface{}) error {

	if !testing.Testing() {
		return errors.New("not allowed to call Wire method in non-test mode")
	}

	stack := wiring.NewStack()
	defer func() {
		if len(stack.Beans) > 0 {
			syslog.Infof("wiring path %s", stack.Path())
		}
	}()

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	return c.wiring.WireBeanValue(v, t, false, stack)
}

// Close closes the container and cleans up resources.
func (c *Container) Close() {
	c.wiring.Close()
}
