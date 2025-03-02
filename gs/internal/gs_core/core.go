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
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/syslog"
	"github.com/spf13/cast"
)

type refreshState int

const (
	RefreshDefault = refreshState(iota) // 未刷新
	RefreshInit                         // 准备刷新
	Refreshing                          // 正在刷新
	Refreshed                           // 已刷新
)

type BeanRuntime interface {
	Name() string
	Type() reflect.Type
	Value() reflect.Value
	Interface() interface{}
	Callable() gs.Callable
	Status() gs_bean.BeanStatus
	String() string
}

// Container 是 go-spring 框架的基石，实现了 Martin Fowler 在 << Inversion
// of Control Containers and the Dependency Injection pattern >> 一文中
// 提及的依赖注入的概念。但原文的依赖注入仅仅是指对象之间的依赖关系处理，而有些 IoC
// 容器在实现时比如 Spring 还引入了对属性 property 的处理。通常大家会用依赖注入统
// 述上面两种概念，但实际上使用属性绑定来描述对 property 的处理会更加合适，因此
// go-spring 严格区分了这两种概念，在描述对 bean 的处理时要么单独使用依赖注入或属
// 性绑定，要么同时使用依赖注入和属性绑定。
type Container struct {
	state     refreshState
	resolving *Resolving
	wiring    *Wiring
}

// New 创建 IoC 容器。
func New() gs.Container {
	return &Container{
		resolving: &Resolving{},
		wiring: &Wiring{
			p:           gs_dync.New(),
			beansByName: make(map[string][]BeanRuntime),
			beansByType: make(map[reflect.Type][]BeanRuntime),
		},
	}
}

func (c *Container) Wiring() *Wiring {
	return c.wiring
}

// Mock mocks the bean with the given object.
func (c *Container) Mock(obj interface{}, target gs.BeanSelectorInterface) {
	x := BeanMock{Object: obj, Target: target}
	c.resolving.mocks = append(c.resolving.mocks, x)
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
	if c.state >= Refreshing {
		return nil
	}
	x := b.BeanRegistration().(*gs_bean.BeanDefinition)
	c.resolving.beans = append(c.resolving.beans, x)
	return gs.NewRegisteredBean(b.BeanRegistration())
}

func (c *Container) GroupRegister(fn gs.GroupFunc) {
	c.resolving.group = append(c.resolving.group, fn)
}

// RefreshProperties updates the properties of the container.
func (c *Container) RefreshProperties(p gs.Properties) error {
	return c.wiring.p.Refresh(p)
}

// Refresh initializes and wires all beans in the container.
func (c *Container) Refresh() (err error) {
	if c.state != RefreshDefault {
		return errors.New("container is refreshing or refreshed")
	}
	c.state = RefreshInit
	start := time.Now()

	err = c.resolving.RefreshInit(c.wiring.p.Data())
	if err != nil {
		return err
	}

	c.state = Refreshing

	c.wiring.AllowCircularReferences = cast.ToBool(c.wiring.p.Data().Get("spring.allow-circular-references"))
	c.wiring.ForceAutowireIsNullable = cast.ToBool(c.wiring.p.Data().Get("spring.force-autowire-is-nullable"))

	err = c.resolving.Refresh(c.wiring.p.Data())
	if err != nil {
		return err
	}

	// registers all beans
	var beans []*gs_bean.BeanDefinition
	for _, b := range c.resolving.beans {
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		c.wiring.beansByName[b.Name()] = append(c.wiring.beansByName[b.Name()], b)
		c.wiring.beansByType[b.Type()] = append(c.wiring.beansByType[b.Type()], b)
		for _, t := range b.Exports() {
			c.wiring.beansByType[t] = append(c.wiring.beansByType[t], b)
		}
		beans = append(beans, b)
	}

	stack := NewWiringStack()
	defer func() {
		if err != nil || len(stack.beans) > 0 {
			err = fmt.Errorf("%s ↩\n%s", err, stack.path())
			syslog.Errorf("%s", err.Error())
		}
	}()

	// injects all beans
	c.wiring.state = Refreshing
	for _, b := range beans {
		if err = c.wiring.wireBean(b, stack); err != nil {
			return err
		}
	}
	c.wiring.state = Refreshed

	if c.wiring.AllowCircularReferences {
		// processes the bean fields that are marked for lazy injection.
		for _, f := range stack.lazyFields {
			tag := strings.TrimSuffix(f.tag, ",lazy")
			if err = c.wiring.wireStructField(f.v, tag, stack); err != nil {
				return fmt.Errorf("%q wired error: %s", f.path, err.Error())
			}
		}
	} else if len(stack.lazyFields) > 0 {
		return errors.New("found circular references in beans")
	}

	c.wiring.destroyers, err = stack.getSortedDestroyers()
	if err != nil {
		return err
	}

	// registers all beans
	c.wiring.beansByName = make(map[string][]BeanRuntime)
	c.wiring.beansByType = make(map[reflect.Type][]BeanRuntime)
	for _, b := range c.resolving.beans {
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		c.wiring.beansByName[b.Name()] = append(c.wiring.beansByName[b.Name()], b.BeanRuntime)
		c.wiring.beansByType[b.Type()] = append(c.wiring.beansByType[b.Type()], b.BeanRuntime)
		for _, t := range b.Exports() {
			c.wiring.beansByType[t] = append(c.wiring.beansByType[t], b.BeanRuntime)
		}
	}

	if !testing.Testing() {
		if c.wiring.p.ObjectsCount() == 0 {
			c.wiring.p = nil
		}
		c.wiring.beansByName = nil
		c.wiring.beansByType = nil
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

	stack := NewWiringStack()
	defer func() {
		if len(stack.beans) > 0 {
			syslog.Infof("wiring path %s", stack.path())
		}
	}()

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	return c.wiring.wireBeanValue(v, t, false, stack)
}

// Close closes the container and cleans up resources.
func (c *Container) Close() {
	for _, f := range c.wiring.destroyers {
		f()
	}
}

type BeanMock struct {
	Object interface{}
	Target gs.BeanSelectorInterface
}

type Resolving struct {
	mocks []BeanMock
	beans []*gs_bean.BeanDefinition
	group []gs.GroupFunc
}

func (c *Resolving) RefreshInit(p gs.Properties) error {
	// processes all group functions to register beans.
	for _, fn := range c.group {
		beans, err := fn(p)
		if err != nil {
			return err
		}
		for _, b := range beans {
			d := b.BeanRegistration().(*gs_bean.BeanDefinition)
			c.beans = append(c.beans, d)
		}
	}

	// processes configuration beans to register beans.
	for _, b := range c.beans {
		if !b.ConfigurationBean() {
			continue
		}
		newBeans, err := c.scanConfiguration(b)
		if err != nil {
			return err
		}
		c.beans = append(c.beans, newBeans...)
	}
	return nil
}

func (c *Resolving) Refresh(p gs.Properties) error {

	// resolves all beans on their condition.
	ctx := &CondContext{p: p, c: c}
	for _, b := range c.beans {
		if err := ctx.resolveBean(b); err != nil {
			return err
		}
	}

	type BeanID struct {
		s string
		t reflect.Type
	}

	// caches all beans by id and checks for duplicates.
	beansByID := make(map[BeanID]*gs_bean.BeanDefinition)
	for _, b := range c.beans {
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		if b.Status() != gs_bean.StatusResolved {
			return fmt.Errorf("unexpected status %d", b.Status())
		}
		beanID := BeanID{b.Name(), b.Type()}
		if d, ok := beansByID[beanID]; ok {
			return fmt.Errorf("found duplicate beans [%s] [%s]", b, d)
		}
		beansByID[beanID] = b
	}
	return nil
}

func (c *Resolving) scanConfiguration(bd *gs_bean.BeanDefinition) ([]*gs_bean.BeanDefinition, error) {
	var (
		includes []*regexp.Regexp
		excludes []*regexp.Regexp
	)
	param := bd.ConfigurationParam()
	ss := param.Includes
	if len(ss) == 0 {
		ss = []string{"New*"}
	}
	for _, s := range ss {
		var x *regexp.Regexp
		x, err := regexp.Compile(s)
		if err != nil {
			return nil, err
		}
		includes = append(includes, x)
	}
	ss = param.Excludes
	for _, s := range ss {
		var x *regexp.Regexp
		x, err := regexp.Compile(s)
		if err != nil {
			return nil, err
		}
		excludes = append(excludes, x)
	}
	var newBeans []*gs_bean.BeanDefinition
	n := bd.Type().NumMethod()
	for i := 0; i < n; i++ {
		m := bd.Type().Method(i)
		skip := false
		for _, x := range excludes {
			if x.MatchString(m.Name) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		for _, x := range includes {
			if !x.MatchString(m.Name) {
				continue
			}
			fnType := m.Func.Type()
			out0 := fnType.Out(0)
			file, line, _ := util.FileLine(m.Func.Interface())
			f, err := gs_arg.Bind(m.Func.Interface(), []gs.Arg{
				gs_arg.Tag(bd.Name()),
			})
			if err != nil {
				return nil, err
			}
			f.SetFileLine(file, line)
			v := reflect.New(out0)
			if util.IsBeanType(out0) {
				v = v.Elem()
			}
			name := bd.Name() + "_" + m.Name
			b := gs_bean.NewBean(v.Type(), v, f, name)
			b.SetFileLine(file, line)
			b.SetCondition(gs_cond.OnBean(bd))
			newBeans = append(newBeans, b)
			break
		}
	}
	return newBeans, nil
}

type CondContext struct {
	c *Resolving
	p gs.Properties
}

// resolveBean determines the validity of the bean.
func (c *CondContext) resolveBean(b *gs_bean.BeanDefinition) error {
	if b.Status() >= gs_bean.StatusResolving {
		return nil
	}
	b.SetStatus(gs_bean.StatusResolving)
	for _, cond := range b.Conditions() {
		if ok, err := cond.Matches(c); err != nil {
			return err
		} else if !ok {
			b.SetStatus(gs_bean.StatusDeleted)
			return nil
		}
	}
	b.SetStatus(gs_bean.StatusResolved)
	return nil
}

func (c *CondContext) Has(key string) bool {
	return c.p.Has(key)
}

func (c *CondContext) Prop(key string, def ...string) string {
	return c.p.Get(key, def...)
}

// Find 查找符合条件的 bean 对象，注意该函数只能保证返回的 bean 是有效的，
// 即未被标记为删除的，而不能保证已经完成属性绑定和依赖注入。
func (c *CondContext) Find(s gs.BeanSelectorInterface) ([]gs.CondBean, error) {
	t, name := s.TypeAndName()
	var result []gs.CondBean
	for _, b := range c.c.beans {
		if b.Status() == gs_bean.StatusResolving || b.Status() == gs_bean.StatusDeleted {
			continue
		}
		if t != nil {
			if b.Type() != t {
				foundType := false
				for _, typ := range b.Exports() {
					if typ == t {
						foundType = true
						break
					}
				}
				if !foundType {
					continue
				}
			}
		}
		if name != "" && name != b.Name() {
			continue
		}
		if err := c.resolveBean(b); err != nil {
			return nil, err
		}
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		result = append(result, b)
	}
	return result, nil
}
