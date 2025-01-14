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

func (c *Container) SubKeys(key string) ([]string, error) {
	return c.p.Data().SubKeys(key)
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

func (c *Container) RefreshProperties(p gs.Properties) error {
	return c.p.Refresh(p)
}

// Refresh refreshes the container, registering, resolving and injecting beans.
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
			syslog.Error("%s", err.Error())
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
	syslog.Info("container is refreshed successfully, %d beans cost %v",
		len(beansById), time.Now().Sub(start))
	return nil
}

// ReleaseUnusedMemory releases unused memory.
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
			syslog.Info("wiring path %s", stack.path())
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

	defer func() {
		if len(stack.beans) > 0 {
			syslog.Info("wiring path %s", stack.path())
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

// Invoke 调用函数，函数的参数会自动注入，函数的返回值也会自动注入。
func (c *Container) Invoke(fn interface{}, args ...gs.Arg) ([]interface{}, error) {

	if !util.IsFuncType(reflect.TypeOf(fn)) {
		return nil, errors.New("fn should be func type")
	}

	stack := newWiringStack()

	defer func() {
		if len(stack.beans) > 0 {
			syslog.Info("wiring path %s", stack.path())
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

// Go 创建安全可等待的 goroutine，fn 要求的 ctx 对象由 IoC 容器提供，当 IoC 容
// 器关闭时 ctx会 发出 Done 信号， fn 在接收到此信号后应当立即退出。
func (c *Container) Go(fn func(ctx context.Context)) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				syslog.Error("%v", r)
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

	syslog.Info("goroutines exited")

	for _, f := range c.destroyers {
		f()
	}

	syslog.Info("container closed")
}
