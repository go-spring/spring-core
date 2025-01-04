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

// Package gs_ctx 实现了 go-spring 的核心骨架，包含 IoC 容器、基于 IoC 容器的 App
// 以及全局 App 对象封装三个部分，可以应用于多种使用场景。
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

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/dync"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
)

type BeanInit interface {
	OnInit(ctx gs.Context) error
}

type BeanDestroy interface {
	OnDestroy()
}

type refreshState int

const (
	Unrefreshed = refreshState(iota) // 未刷新
	RefreshInit                      // 准备刷新
	Refreshing                       // 正在刷新
	Refreshed                        // 已刷新
)

// ContextAware injects the Context into a struct as the field GSContext.
type ContextAware struct {
	GSContext gs.Context `autowire:""`
}

// Container 是 go-spring 框架的基石，实现了 Martin Fowler 在 << Inversion
// of Control Containers and the Dependency Injection pattern >> 一文中
// 提及的依赖注入的概念。但原文的依赖注入仅仅是指对象之间的依赖关系处理，而有些 IoC
// 容器在实现时比如 Spring 还引入了对属性 property 的处理。通常大家会用依赖注入统
// 述上面两种概念，但实际上使用属性绑定来描述对 property 的处理会更加合适，因此
// go-spring 严格区分了这两种概念，在描述对 bean 的处理时要么单独使用依赖注入或属
// 性绑定，要么同时使用依赖注入和属性绑定。
type Container struct {
	beans                   []*gs.BeanDefinition
	beansByName             map[string][]*gs.BeanDefinition
	beansByType             map[reflect.Type][]*gs.BeanDefinition
	groupFuncs              []gs.GroupFunc
	ctx                     context.Context
	cancel                  context.CancelFunc
	destroyers              []func()
	state                   refreshState
	wg                      sync.WaitGroup
	p                       *dync.Properties
	ContextAware            bool
	AllowCircularReferences bool `value:"${spring.main.allow-circular-references:=false}"`
}

// New 创建 IoC 容器。
func New() *Container {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Container{
		ctx:         ctx,
		cancel:      cancel,
		p:           dync.New(),
		beansByName: make(map[string][]*gs.BeanDefinition),
		beansByType: make(map[reflect.Type][]*gs.BeanDefinition),
	}
	c.Object(c).Export((*gs.Context)(nil))
	return c
}

// Object 注册对象形式的 bean ，需要注意的是该方法在注入开始后就不能再调用了。
func (c *Container) Object(i interface{}) *gs.BeanDefinition {
	return c.Accept(NewBean(reflect.ValueOf(i)))
}

// Provide 注册构造函数形式的 bean ，需要注意的是该方法在注入开始后就不能再调用了。
func (c *Container) Provide(ctor interface{}, args ...gs.Arg) *gs.BeanDefinition {
	return c.Accept(NewBean(ctor, args...))
}

func (c *Container) Accept(b *gs.BeanDefinition) *gs.BeanDefinition {
	if c.state >= Refreshing {
		panic(errors.New("should call before Refresh"))
	}
	c.beans = append(c.beans, b)
	return b
}

func (c *Container) Group(fn gs.GroupFunc) {
	c.groupFuncs = append(c.groupFuncs, fn)
}

// Context 返回 IoC 容器的 ctx 对象。
func (c *Container) Context() context.Context {
	return c.ctx
}

func (c *Container) Keys() []string {
	return c.p.Data().Keys()
}

func (c *Container) Has(key string) bool {
	return c.p.Data().Has(key)
}

func (c *Container) Prop(key string, opts ...conf.GetOption) string {
	return c.p.Data().Get(key, opts...)
}

func (c *Container) Resolve(s string) (string, error) {
	return c.p.Data().Resolve(s)
}

func (c *Container) Bind(i interface{}, opts ...conf.BindArg) error {
	return c.p.Data().Bind(i, opts...)
}

func (c *Container) RefreshProperties(p conf.ReadOnlyProperties) error {
	return c.p.Refresh(p)
}

func (c *Container) Refresh() (err error) {

	if c.state != Unrefreshed {
		return errors.New("Container already refreshed")
	}
	c.state = RefreshInit

	// start := time.Now()

	// 处理 group 逻辑
	for _, fn := range c.groupFuncs {
		var beans []*gs.BeanDefinition
		beans, err = fn(c.p.Data())
		if err != nil {
			return err
		}
		c.beans = append(c.beans, beans...)
	}

	// 处理 configuration 逻辑
	for _, bd := range c.beans {
		if !bd.IsConfiguration() {
			continue
		}
		var newBeans []*gs.BeanDefinition
		newBeans, err = c.scanConfiguration(bd)
		if err != nil {
			return err
		}
		c.beans = append(c.beans, newBeans...)
	}

	c.state = Refreshing

	for _, b := range c.beans {
		c.registerBean(b)
	}

	for _, b := range c.beans {
		if err = c.resolveBean(b); err != nil {
			return err
		}
	}

	beansById := make(map[string]*gs.BeanDefinition)
	{
		for _, b := range c.beans {
			if b.GetStatus() == gs.Deleted {
				continue
			}
			if b.GetStatus() != gs.Resolved {
				return fmt.Errorf("unexpected status %d", b.GetStatus())
			}
			beanID := b.ID()
			if d, ok := beansById[beanID]; ok {
				return fmt.Errorf("found duplicate beans [%s] [%s]", b, d)
			}
			beansById[beanID] = b
		}
	}

	stack := newWiringStack()

	// defer func() {
	// 	if err != nil || len(stack.beans) > 0 {
	// 		err = fmt.Errorf("%s ↩\n%s", err, stack.path())
	// 		c.logger.Error(err)
	// 	}
	// }()

	// 按照 bean id 升序注入，保证注入过程始终一致。
	{
		var keys []string
		for s := range beansById {
			keys = append(keys, s)
		}
		sort.Strings(keys)
		for _, s := range keys {
			b := beansById[s]
			if err = c.wireBean(b, stack); err != nil {
				return err
			}
		}
	}

	if c.AllowCircularReferences {
		// 处理被标记为延迟注入的那些 bean 字段
		for _, f := range stack.lazyFields {
			tag := strings.TrimSuffix(f.tag, ",lazy")
			if err := c.wireByTag(f.v, tag, stack); err != nil {
				return fmt.Errorf("%q wired error: %s", f.path, err.Error())
			}
		}
	} else if len(stack.lazyFields) > 0 {
		return errors.New("remove the dependency cycle between beans")
	}

	c.destroyers = stack.sortDestroyers()
	c.state = Refreshed

	// cost := time.Now().Sub(start)
	// c.logger.Infof("refresh %d beans cost %v", len(beansById), cost)

	// if autoClear && !c.ContextAware {
	// 	c.clear()
	// }

	// c.logger.Info("Container refreshed successfully")
	return nil
}

func (c *Container) scanConfiguration(bd *gs.BeanDefinition) ([]*gs.BeanDefinition, error) {
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
	var newBeans []*gs.BeanDefinition
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
			if x.MatchString(m.Name) {
				var f gs.Callable
				f, err := gs_arg.Bind(m.Func.Interface(), []gs.Arg{bd}, 0)
				if err != nil {
					return nil, err
				}
				name := bd.GetName() + "_" + m.Name
				out0 := m.Type.Out(0)
				v := reflect.New(out0)
				if gs.IsBeanType(out0) {
					v = v.Elem()
				}
				t := v.Type()
				b := gs.NewBean(t, v, f, name, true, bd.File(), bd.Line()).On(gs_cond.OnBean(bd))
				newBeans = append(newBeans, b)
				break
			}
		}
	}
	return newBeans, nil
}

func (c *Container) registerBean(b *gs.BeanDefinition) {
	// c.logger.Debugf("register %s name:%q type:%q %s", b.getClass(), b.BeanName(), b.Type(), b.FileLine())
	c.beansByName[b.GetName()] = append(c.beansByName[b.GetName()], b)
	c.beansByType[b.Type()] = append(c.beansByType[b.Type()], b)
	for _, t := range b.GetExports() {
		// c.logger.Debugf("register %s name:%q type:%q %s", b.getClass(), b.BeanName(), t, b.FileLine())
		c.beansByType[t] = append(c.beansByType[t], b)
	}
}

// resolveBean 判断 bean 的有效性，如果 bean 是无效的则被标记为已删除。
func (c *Container) resolveBean(b *gs.BeanDefinition) error {

	if b.GetStatus() >= gs.Resolving {
		return nil
	}

	b.SetStatus(gs.Resolving)

	// method bean 先确定 parent bean 是否存在
	if b.IsMethod() {
		selector, ok := b.Callable().Arg(0)
		if !ok || selector == "" {
			selector, _ = b.Callable().In(0)
		}
		parents, err := c.Find(selector)
		if err != nil {
			return err
		}
		n := len(parents)
		if n > 1 {
			msg := fmt.Sprintf("found %d parent beans, bean:%q type:%q [", n, selector, b.Type().In(0))
			for _, b := range parents {
				msg += "( " + b.String() + " ), "
			}
			msg = msg[:len(msg)-2] + "]"
			return errors.New(msg)
		} else if n == 0 {
			b.SetStatus(gs.Deleted)
			return nil
		}
	}

	if b.GetCond() != nil {
		if ok, err := b.GetCond().Matches(c); err != nil {
			return err
		} else if !ok {
			b.SetStatus(gs.Deleted)
			return nil
		}
	}

	b.SetStatus(gs.Resolved)
	return nil
}

// Find 查找符合条件的 bean 对象，注意该函数只能保证返回的 bean 是有效的，
// 即未被标记为删除的，而不能保证已经完成属性绑定和依赖注入。
func (c *Container) Find(selector gs.BeanSelector) ([]*gs.BeanDefinition, error) {

	finder := func(fn func(*gs.BeanDefinition) bool) ([]*gs.BeanDefinition, error) {
		var result []*gs.BeanDefinition
		for _, b := range c.beans {
			if b.GetStatus() == gs.Resolving || b.GetStatus() == gs.Deleted || !fn(b) {
				continue
			}
			if err := c.resolveBean(b); err != nil {
				return nil, err
			}
			if b.GetStatus() == gs.Deleted {
				continue
			}
			result = append(result, b)
		}
		return result, nil
	}

	var t reflect.Type
	switch st := selector.(type) {
	case string, gs.BeanDefinition, *gs.BeanDefinition:
		tag, err := c.toWireTag(selector)
		if err != nil {
			return nil, err
		}
		return finder(func(b *gs.BeanDefinition) bool {
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

	return finder(func(b *gs.BeanDefinition) bool {
		if b.Type() == t {
			return true
		}
		for _, typ := range b.GetExports() {
			if typ == t {
				return true
			}
		}
		return false
	})
}

// Get 根据类型和选择器获取符合条件的 bean 对象。当 i 是一个基础类型的 bean 接收
// 者时，表示符合条件的 bean 对象只能有一个，没有找到或者多于一个时会返回 error。
// 当 i 是一个 map 类型的 bean 接收者时，表示获取任意数量的 bean 对象，map 的
// key 是 bean 的名称，map 的 value 是 bean 的地址。当 i 是一个 array 或者
// slice 时，也表示获取任意数量的 bean 对象，但是它会对获取到的 bean 对象进行排序，
// 如果没有传入选择器或者传入的选择器是 * ，则根据 bean 的 order 值进行排序，这种
// 工作模式称为自动模式，否则根据传入的选择器列表进行排序，这种工作模式成为指派模式。
// 该方法和 Find 方法的区别是该方法保证返回的所有 bean 对象都已经完成属性绑定和依
// 赖注入，而 Find 方法只能保证返回的 bean 对象是有效的，即未被标记为删除的。
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
			// c.logger.Infof("wiring path %s", stack.path())
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

// Wire 如果传入的是 bean 对象，则对 bean 对象进行属性绑定和依赖注入，如果传入的
// 是构造函数，则立即执行该构造函数，然后对返回的结果进行属性绑定和依赖注入。无论哪
// 种方式，该函数执行完后都会返回 bean 对象的真实值。
func (c *Container) Wire(objOrCtor interface{}, ctorArgs ...gs.Arg) (interface{}, error) {

	stack := newWiringStack()

	// defer func() {
	// 	if len(stack.beans) > 0 {
	// 		c.logger.Infof("wiring path %s", stack.path())
	// 	}
	// }()

	b := NewBean(objOrCtor, ctorArgs...)
	err := c.wireBean(b, stack)
	if err != nil {
		return nil, err
	}
	return b.Interface(), nil
}

// Invoke 调用函数，函数的参数会自动注入，函数的返回值也会自动注入。
func (c *Container) Invoke(fn interface{}, args ...gs.Arg) ([]interface{}, error) {

	if !gs.IsFuncType(reflect.TypeOf(fn)) {
		return nil, errors.New("fn should be func type")
	}

	stack := newWiringStack()

	// defer func() {
	// 	if len(stack.beans) > 0 {
	// 		c.logger.Infof("wiring path %s", stack.path())
	// 	}
	// }()

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

// Go 创建安全可等待的 goroutine，fn 要求的 ctx 对象由 IoC 容器提供，当 IoC 容
// 器关闭时 ctx会 发出 Done 信号， fn 在接收到此信号后应当立即退出。
func (c *Container) Go(fn func(ctx context.Context)) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				// c.logger.Panic(r)
			}
		}()
		fn(c.ctx)
	}()
}

// Close 关闭容器，此方法必须在 Refresh 之后调用。该方法会触发 ctx 的 Done 信
// 号，然后等待所有 goroutine 结束，最后按照被依赖先销毁的原则执行所有的销毁函数。
func (c *Container) Close() {

	c.cancel()
	c.wg.Wait()

	// c.logger.Info("goroutines exited")

	for _, f := range c.destroyers {
		f()
	}

	// c.logger.Info("Container closed")
}
