/*
 * Copyright 2012-2024 the original author or authors.
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
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
	"github.com/go-spring/spring-core/gs/syslog"
	"github.com/go-spring/spring-core/util"
)

type refreshState int

const (
	RefreshDefault = refreshState(iota) // 未刷新
	RefreshInit                         // 准备刷新
	Refreshing                          // 正在刷新
	Refreshed                           // 已刷新
)

var BeanDefinitionType = reflect.TypeOf((*gs.BeanDefinition)(nil))

type GroupFunc = func(p gs.Properties) ([]*gs.BeanDefinition, error)

type BeanRuntime interface {
	Name() string
	Type() reflect.Type
	Value() reflect.Value
	Interface() interface{}
	Callable() gs.Callable
	Match(typeName string, beanName string) bool
	Status() gs_bean.BeanStatus
	IsPrimary() bool
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
	beans        []*gs_bean.BeanDefinition
	beansByName  map[string][]BeanRuntime
	beansByType  map[reflect.Type][]BeanRuntime
	groupFuncs   []GroupFunc
	p            *gs_dync.Properties
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	state        refreshState
	destroyers   []func()
	ContextAware bool

	AllowCircularReferences bool `value:"${spring.allow-circular-references:=false}"`
	ForceAutowireIsNullable bool `value:"${spring.force-autowire-is-nullable:=false}"`
}

// New 创建 IoC 容器。
func New() gs.Container {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Container{
		ctx:         ctx,
		cancel:      cancel,
		p:           gs_dync.New(),
		beansByName: make(map[string][]BeanRuntime),
		beansByType: make(map[reflect.Type][]BeanRuntime),
	}
	c.Object(c).Export((*gs.Context)(nil))
	return c
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
		panic(errors.New("should call before Refresh"))
	}
	c.beans = append(c.beans, b.BeanRegistration().(*gs_bean.BeanDefinition))
	return gs.NewRegisteredBean(b.BeanRegistration())
}

func (c *Container) GroupRegister(fn GroupFunc) {
	c.groupFuncs = append(c.groupFuncs, fn)
}

// Context returns the root context.Context of the container.
func (c *Container) Context() context.Context {
	return c.ctx
}

// Keys returns all keys present in the container's properties.
func (c *Container) Keys() []string {
	return c.p.Data().Keys()
}

// Has checks if a key exists in the container's properties.
func (c *Container) Has(key string) bool {
	return c.p.Data().Has(key)
}

// SubKeys returns sub-keys under the specified key in the container's properties.
func (c *Container) SubKeys(key string) ([]string, error) {
	return c.p.Data().SubKeys(key)
}

// Prop retrieves the value of the specified key from the container's properties.
func (c *Container) Prop(key string, opts ...conf.GetOption) string {
	return c.p.Data().Get(key, opts...)
}

// Resolve resolves placeholders or references in the given string.
func (c *Container) Resolve(s string) (string, error) {
	return c.p.Data().Resolve(s)
}

// Bind binds the value of the specified key to the provided struct or variable.
func (c *Container) Bind(i interface{}, opts ...conf.BindArg) error {
	return c.p.Data().Bind(i, opts...)
}

// RefreshProperties updates the properties of the container.
func (c *Container) RefreshProperties(p gs.Properties) error {
	return c.p.Refresh(p)
}

// Refresh initializes and wires all beans in the container.
func (c *Container) Refresh() (err error) {
	if c.state != RefreshDefault {
		return errors.New("container is refreshing or refreshed")
	}
	c.state = RefreshInit
	start := time.Now()

	// processes all group functions to register beans.
	for _, fn := range c.groupFuncs {
		var beans []*gs.BeanDefinition
		beans, err = fn(c.p.Data())
		if err != nil {
			return err
		}
		for _, b := range beans {
			d := b.BeanRegistration().(*gs_bean.BeanDefinition)
			c.beans = append(c.beans, d)
		}
	}
	c.groupFuncs = nil

	// processes configuration beans to register beans.
	for _, b := range c.beans {
		if !b.IsConfiguration() {
			continue
		}
		var newBeans []*gs_bean.BeanDefinition
		newBeans, err = c.scanConfiguration(b)
		if err != nil {
			return err
		}
		c.beans = append(c.beans, newBeans...)
	}

	c.state = Refreshing

	// registers all beans by name and type.
	for _, b := range c.beans {
		c.registerBean(b)
	}

	// resolves all beans on their condition.
	for _, b := range c.beans {
		if err = c.resolveBean(b); err != nil {
			return err
		}
	}

	// caches all beans by id and checks for duplicates.
	beansById := make(map[string]*gs_bean.BeanDefinition)
	for _, b := range c.beans {
		if b.Status() == gs_bean.Deleted {
			continue
		}
		if b.Status() != gs_bean.Resolved {
			return fmt.Errorf("unexpected status %d", b.Status())
		}
		beanID := b.ID()
		if d, ok := beansById[beanID]; ok {
			return fmt.Errorf("found duplicate beans [%s] [%s]", b, d)
		}
		beansById[beanID] = b
	}

	stack := newWiringStack()
	defer func() {
		if err != nil || len(stack.beans) > 0 {
			err = fmt.Errorf("%s ↩\n%s", err, stack.path())
			syslog.Errorf("%s", err.Error())
		}
	}()

	// injects all beans in ascending order of their IDs
	{
		var keys []string
		for s := range beansById {
			keys = append(keys, s)
		}
		sort.Strings(keys)
		for _, s := range keys {
			if err = c.wireBeanInRefreshing(beansById[s], stack); err != nil {
				return err
			}
		}
	}

	if c.AllowCircularReferences {
		// processes the bean fields that are marked for lazy injection.
		for _, f := range stack.lazyFields {
			tag := strings.TrimSuffix(f.tag, ",lazy")
			if err = c.wireByTag(f.v, tag, stack); err != nil {
				return fmt.Errorf("%q wired error: %s", f.path, err.Error())
			}
		}
	} else if len(stack.lazyFields) > 0 {
		return errors.New("found circular references in beans")
	}

	c.destroyers = stack.sortDestroyers()

	// retains only the runtime essential content to simplify memory.
	c.beansByName = make(map[string][]BeanRuntime)
	c.beansByType = make(map[reflect.Type][]BeanRuntime)
	for _, b := range c.beans {
		c.registerBean(b)
	}

	c.state = Refreshed
	syslog.Debugf("container is refreshed successfully, %d beans cost %v",
		len(beansById), time.Now().Sub(start))
	return nil
}

// ReleaseUnusedMemory releases unused memory by cleaning up unnecessary resources.
func (c *Container) ReleaseUnusedMemory() {
	if !c.ContextAware { // 保留核心数据
		if c.p.ObjectsCount() == 0 {
			c.p = nil
		}
		c.beansByName = nil
		c.beansByType = nil
	}
	c.beans = nil
}

func (c *Container) scanConfiguration(bd *gs_bean.BeanDefinition) ([]*gs_bean.BeanDefinition, error) {
	var (
		includes []*regexp.Regexp
		excludes []*regexp.Regexp
	)
	ss := bd.GetIncludeMethod()
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
	ss = bd.GetExcludeMethod()
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
			if out0 == BeanDefinitionType {
				ret := m.Func.Call([]reflect.Value{bd.Value()})
				if len(ret) > 1 {
					if err := ret[1].Interface().(error); err != nil {
						return nil, err
					}
				}
				b := ret[0].Interface().(*gs.BeanDefinition)
				newBeans = append(newBeans, b.BeanRegistration().(*gs_bean.BeanDefinition))
				retBeans, err := c.scanConfiguration(b.BeanRegistration().(*gs_bean.BeanDefinition))
				if err != nil {
					return nil, err
				}
				newBeans = append(newBeans, retBeans...)
			} else {
				var f gs.Callable
				f, err := gs_arg.Bind(m.Func.Interface(), []gs.Arg{bd.ID()}, 0)
				if err != nil {
					return nil, err
				}
				v := reflect.New(out0)
				if util.IsBeanType(out0) {
					v = v.Elem()
				}
				name := bd.Name() + "_" + m.Name
				b := gs_bean.NewBean(v.Type(), v, f, name) // todo
				gs.NewBeanDefinition(b).Condition(gs_cond.OnBean(bd))
				newBeans = append(newBeans, b)
			}
			break
		}
	}
	return newBeans, nil
}

// registerBean registers a bean by name and type.
func (c *Container) registerBean(b *gs_bean.BeanDefinition) {
	if b.Status() == gs_bean.Deleted {
		return
	}
	c.beansByName[b.Name()] = append(c.beansByName[b.Name()], b)
	c.beansByType[b.Type()] = append(c.beansByType[b.Type()], b)
	for _, t := range b.Exports() {
		c.beansByType[t] = append(c.beansByType[t], b)
	}
}

// resolveBean determines the validity of the bean.
func (c *Container) resolveBean(b *gs_bean.BeanDefinition) error {
	if b.Status() >= gs_bean.Resolving {
		return nil
	}
	b.SetStatus(gs_bean.Resolving)
	if cond := b.Condition(); cond != nil {
		if ok, err := cond.Matches(c); err != nil {
			return err
		} else if !ok {
			b.SetStatus(gs_bean.Deleted)
			return nil
		}
	}
	b.SetStatus(gs_bean.Resolved)
	return nil
}

// Find 查找符合条件的 bean 对象，注意该函数只能保证返回的 bean 是有效的，
// 即未被标记为删除的，而不能保证已经完成属性绑定和依赖注入。
func (c *Container) Find(selector gs.BeanSelector) ([]gs.CondBean, error) {

	finder := func(fn func(*gs_bean.BeanDefinition) bool) ([]gs.CondBean, error) {
		var result []gs.CondBean
		for _, b := range c.beans {
			if b.Status() == gs_bean.Resolving || b.Status() == gs_bean.Deleted || !fn(b) {
				continue
			}
			if err := c.resolveBean(b); err != nil {
				return nil, err
			}
			if b.Status() == gs_bean.Deleted {
				continue
			}
			result = append(result, b)
		}
		return result, nil
	}

	var t reflect.Type
	switch st := selector.(type) {
	case string, *gs_bean.BeanDefinition:
		tag, err := c.toWireTag(selector)
		if err != nil {
			return nil, err
		}
		return finder(func(b *gs_bean.BeanDefinition) bool {
			return b.Match(tag.typeName, tag.beanName)
		})
	case reflect.Type:
		t = st
	default:
		t = reflect.TypeOf(st)
	}

	if t.Kind() == reflect.Ptr {
		if e := t.Elem(); e.Kind() == reflect.Interface {
			t = e // 指 (*error)(nil) 形式的 bean 选择器
		}
	}

	return finder(func(b *gs_bean.BeanDefinition) bool {
		if b.Type() == t {
			return true
		}
		return false
	})
}

// Get retrieves a bean of the specified type using the provided selectors.
func (c *Container) Get(i interface{}, selectors ...gs.BeanSelector) error {

	if i == nil {
		return errors.New("i can't be nil")
	}

	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return errors.New("i must be pointer")
	}

	stack := newWiringStack()

	defer func() {
		if len(stack.beans) > 0 {
			syslog.Infof("wiring path %s", stack.path())
		}
	}()

	var tags []wireTag
	for _, s := range selectors {
		g, err := c.toWireTag(s)
		if err != nil {
			return err
		}
		tags = append(tags, g)
	}
	return c.autowire(v.Elem(), tags, false, stack)
}

// Wire creates and returns a wired bean using the provided object or constructor function.
func (c *Container) Wire(objOrCtor interface{}, ctorArgs ...gs.Arg) (interface{}, error) {

	stack := newWiringStack()

	defer func() {
		if len(stack.beans) > 0 {
			syslog.Infof("wiring path %s", stack.path())
		}
	}()

	b := NewBean(objOrCtor, ctorArgs...)
	var err error
	switch c.state {
	case Refreshing:
		err = c.wireBeanInRefreshing(b.BeanRegistration().(*gs_bean.BeanDefinition), stack)
	case Refreshed:
		err = c.wireBeanAfterRefreshed(b.BeanRegistration().(*gs_bean.BeanDefinition), stack)
	default:
		err = errors.New("state is error for wiring")
	}
	if err != nil {
		return nil, err
	}
	return b.BeanRegistration().(*gs_bean.BeanDefinition).Interface(), nil
}

// Invoke calls the provided function with the specified arguments.
func (c *Container) Invoke(fn interface{}, args ...gs.Arg) ([]interface{}, error) {

	if !util.IsFuncType(reflect.TypeOf(fn)) {
		return nil, errors.New("fn should be func type")
	}

	stack := newWiringStack()

	defer func() {
		if len(stack.beans) > 0 {
			syslog.Infof("wiring path %s", stack.path())
		}
	}()

	r, err := gs_arg.Bind(fn, args, 1)
	if err != nil {
		return nil, err
	}

	ret, err := r.Call(&argContext{c: c, stack: stack})
	if err != nil {
		return nil, err
	}

	var a []interface{}
	for _, v := range ret {
		a = append(a, v.Interface())
	}
	return a, nil
}

// Go runs the provided function in a new goroutine. When the container is closed,
// the context.Context will be canceled.
func (c *Container) Go(fn func(ctx context.Context)) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				syslog.Errorf("%v", r)
			}
		}()
		fn(c.ctx)
	}()
}

// Close closes the container and cleans up resources.
func (c *Container) Close() {

	c.cancel()
	c.wg.Wait()

	syslog.Infof("goroutines exited")

	for _, f := range c.destroyers {
		f()
	}

	syslog.Infof("container closed")
}
