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

// Package gsioc 实现了 go-spring 的核心骨架，包含 IoC 容器、基于 IoC 容器的 App
// 以及全局 App 对象封装三个部分，可以应用于多种使用场景。
package gs_ctx

import (
	"bytes"
	"container/list"
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/dync"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_util"
	"github.com/go-spring/spring-core/util"
)

type refreshState int

const (
	Unrefreshed = refreshState(iota) // 未刷新
	RefreshInit                      // 准备刷新
	Refreshing                       // 正在刷新
	Refreshed                        // 已刷新
)

var (
	contextType = reflect.TypeOf((*gs.Context)(nil)).Elem()
)

// ContextAware injects the Context into a struct as the field GSContext.
type ContextAware struct {
	GSContext gs.Context `autowire:""`
}

type tempContainer struct {
	beans       []*gs.BeanDefinition
	beansByName map[string][]*gs.BeanDefinition
	beansByType map[reflect.Type][]*gs.BeanDefinition
	groupFuncs  []gs.GroupFunc
}

// Container 是 go-spring 框架的基石，实现了 Martin Fowler 在 << Inversion
// of Control Containers and the Dependency Injection pattern >> 一文中
// 提及的依赖注入的概念。但原文的依赖注入仅仅是指对象之间的依赖关系处理，而有些 IoC
// 容器在实现时比如 Spring 还引入了对属性 property 的处理。通常大家会用依赖注入统
// 述上面两种概念，但实际上使用属性绑定来描述对 property 的处理会更加合适，因此
// go-spring 严格区分了这两种概念，在描述对 bean 的处理时要么单独使用依赖注入或属
// 性绑定，要么同时使用依赖注入和属性绑定。
type Container struct {
	*tempContainer
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
	return &Container{
		ctx:    ctx,
		cancel: cancel,
		p:      dync.New(),
		tempContainer: &tempContainer{
			beansByName: make(map[string][]*gs.BeanDefinition),
			beansByType: make(map[reflect.Type][]*gs.BeanDefinition),
		},
	}
}

// Context 返回 IoC 容器的 ctx 对象。
func (c *Container) Context() context.Context {
	return c.ctx
}

func (c *Container) Properties() *dync.Properties {
	return c.p
}

type BeanInit interface {
	OnInit(ctx gs.Context) error
}

type BeanDestroy interface {
	OnDestroy()
}

// NewBean 普通函数注册时需要使用 reflect.ValueOf(fn) 形式以避免和构造函数发生冲突。
func NewBean(objOrCtor interface{}, ctorArgs ...gs.Arg) *gs.BeanDefinition {

	var v reflect.Value
	var fromValue bool
	var method bool
	var name string

	switch i := objOrCtor.(type) {
	case reflect.Value:
		fromValue = true
		v = i
	default:
		v = reflect.ValueOf(i)
	}

	t := v.Type()
	if !gs_util.IsBeanType(t) {
		panic(errors.New("bean must be ref type"))
	}

	if !v.IsValid() || v.IsNil() {
		panic(errors.New("bean can't be nil"))
	}

	const skip = 2
	var f gs.Callable
	_, file, line, _ := runtime.Caller(skip)

	// 以 reflect.ValueOf(fn) 形式注册的函数被视为函数对象 bean 。
	if !fromValue && t.Kind() == reflect.Func {

		if !gs_util.IsConstructor(t) {
			t1 := "func(...)bean"
			t2 := "func(...)(bean, error)"
			panic(fmt.Errorf("constructor should be %s or %s", t1, t2))
		}

		var err error
		f, err = gs_arg.Bind(objOrCtor, ctorArgs, skip)
		if err != nil {
			panic(err)
		}

		out0 := t.Out(0)
		v = reflect.New(out0)
		if gs_util.IsBeanType(out0) {
			v = v.Elem()
		}

		t = v.Type()
		if !gs_util.IsBeanType(t) {
			panic(errors.New("bean must be ref type"))
		}

		// 成员方法一般是 xxx/gs_test.(*Server).Consumer 形式命名
		fnPtr := reflect.ValueOf(objOrCtor).Pointer()
		fnInfo := runtime.FuncForPC(fnPtr)
		funcName := fnInfo.Name()
		name = funcName[strings.LastIndex(funcName, "/")+1:]
		name = name[strings.Index(name, ".")+1:]
		if name[0] == '(' {
			name = name[strings.Index(name, ".")+1:]
		}
		method = strings.LastIndexByte(fnInfo.Name(), ')') > 0
	}

	if t.Kind() == reflect.Ptr && !conf.IsValueType(t.Elem()) {
		panic(errors.New("bean should be *val but not *ref"))
	}

	// Type.String() 一般返回 *pkg.Type 形式的字符串，
	// 我们只取最后的类型名，如有需要请自定义 bean 名称。
	if name == "" {
		s := strings.Split(t.String(), ".")
		name = strings.TrimPrefix(s[len(s)-1], "*")
	}

	return gs.NewBean(t, v, f, name, method, file, line)
}

func (c *Container) Group(fn gs.GroupFunc) {
	c.groupFuncs = append(c.groupFuncs, fn)
}

func (c *Container) Accept(b *gs.BeanDefinition) *gs.BeanDefinition {
	if c.state >= Refreshing {
		panic(errors.New("should call before Refresh"))
	}
	c.beans = append(c.beans, b)
	return b
}

// Object 注册对象形式的 bean ，需要注意的是该方法在注入开始后就不能再调用了。
func (c *Container) Object(i interface{}) *gs.BeanDefinition {
	return c.Accept(NewBean(reflect.ValueOf(i)))
}

// Provide 注册构造函数形式的 bean ，需要注意的是该方法在注入开始后就不能再调用了。
func (c *Container) Provide(ctor interface{}, args ...gs.Arg) *gs.BeanDefinition {
	return c.Accept(NewBean(ctor, args...))
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

// destroyer 保存具有销毁函数的 bean 以及销毁函数的调用顺序。
type destroyer struct {
	current *gs.BeanDefinition
	earlier []*gs.BeanDefinition
}

func (d *destroyer) foundEarlier(b *gs.BeanDefinition) bool {
	for _, c := range d.earlier {
		if c == b {
			return true
		}
	}
	return false
}

// after 添加一个需要在该 bean 的销毁函数执行之前调用销毁函数的 bean 。
func (d *destroyer) after(b *gs.BeanDefinition) {
	if d.foundEarlier(b) {
		return
	}
	d.earlier = append(d.earlier, b)
}

// getBeforeDestroyers 获取排在 i 前面的 destroyer，用于 sort.Triple 排序。
func getBeforeDestroyers(destroyers *list.List, i interface{}) *list.List {
	d := i.(*destroyer)
	result := list.New()
	for e := destroyers.Front(); e != nil; e = e.Next() {
		c := e.Value.(*destroyer)
		if d.foundEarlier(c.current) {
			result.PushBack(c)
		}
	}
	return result
}

type lazyField struct {
	v    reflect.Value
	path string
	tag  string
}

// wiringStack 记录 bean 的注入路径。
type wiringStack struct {
	destroyers   *list.List
	destroyerMap map[string]*destroyer
	beans        []*gs.BeanDefinition
	lazyFields   []lazyField
}

func newWiringStack() *wiringStack {
	return &wiringStack{
		destroyers:   list.New(),
		destroyerMap: make(map[string]*destroyer),
	}
}

// pushBack 添加一个即将注入的 bean 。
func (s *wiringStack) pushBack(b *gs.BeanDefinition) {
	// s.logger.Tracef("push %s %s", b, getStatusString(b.status))
	s.beans = append(s.beans, b)
}

// popBack 删除一个已经注入的 bean 。
func (s *wiringStack) popBack() {
	n := len(s.beans)
	// b := s.beans[n-1]
	s.beans = s.beans[:n-1]
	// s.logger.Tracef("pop %s %s", b, getStatusString(b.status))
}

// path 返回 bean 的注入路径。
func (s *wiringStack) path() (path string) {
	for _, b := range s.beans {
		path += fmt.Sprintf("=> %s ↩\n", b)
	}
	return path[:len(path)-1]
}

// saveDestroyer 记录具有销毁函数的 bean ，因为可能有多个依赖，因此需要排重处理。
func (s *wiringStack) saveDestroyer(b *gs.BeanDefinition) *destroyer {
	d, ok := s.destroyerMap[b.ID()]
	if !ok {
		d = &destroyer{current: b}
		s.destroyerMap[b.ID()] = d
	}
	return d
}

// sortDestroyers 对具有销毁函数的 bean 按照销毁函数的依赖顺序进行排序。
func (s *wiringStack) sortDestroyers() []func() {

	destroy := func(v reflect.Value, fn interface{}) func() {
		return func() {
			if fn == nil {
				v.Interface().(BeanDestroy).OnDestroy()
			} else {
				fnValue := reflect.ValueOf(fn)
				out := fnValue.Call([]reflect.Value{v})
				if len(out) > 0 && !out[0].IsNil() {
					// s.logger.Error(out[0].Interface().(error))
				}
			}
		}
	}

	destroyers := list.New()
	for _, d := range s.destroyerMap {
		destroyers.PushBack(d)
	}
	destroyers = util.TripleSort(destroyers, getBeforeDestroyers)

	var ret []func()
	for e := destroyers.Front(); e != nil; e = e.Next() {
		d := e.Value.(*destroyer).current
		ret = append(ret, destroy(d.Value(), d.GetDestroy()))
	}
	return ret
}

func (c *Container) clear() {
	c.tempContainer = nil
}

func (c *Container) RefreshProperties(p *conf.Properties) error {
	return c.p.Refresh(p)
}

func (c *Container) Refresh(autoClear bool) (err error) {

	if c.state != Unrefreshed {
		return errors.New("Container already refreshed")
	}
	c.state = RefreshInit

	// start := time.Now()
	c.Object(c).Export((*gs.Context)(nil))

	for _, fn := range c.groupFuncs {
		var beans []*gs.BeanDefinition
		beans, err = fn(c.p.Data())
		if err != nil {
			return err
		}
		c.beans = append(c.beans, beans...)
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

	if autoClear && !c.ContextAware {
		c.clear()
	}

	// c.logger.Info("Container refreshed successfully")
	return nil
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
		selector, ok := b.F.Arg(0)
		if !ok || selector == "" {
			selector, _ = b.F.In(0)
		}
		parents, err := c.Find(selector)
		if err != nil {
			return err
		}
		n := len(parents)
		if n > 1 {
			msg := fmt.Sprintf("found %d parent beans, bean:%q type:%q [", n, selector, b.T.In(0))
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

// wireTag 注入语法的 tag 分解式，字符串形式的完整格式为 TypeName:BeanName? 。
// 注入语法的字符串表示形式分为三个部分，TypeName 是原始类型的全限定名，BeanName
// 是 bean 注册时设置的名称，? 表示注入结果允许为空。
type wireTag struct {
	typeName string
	beanName string
	nullable bool
}

func parseWireTag(str string) (tag wireTag) {

	if str == "" {
		return
	}

	if n := len(str) - 1; str[n] == '?' {
		tag.nullable = true
		str = str[:n]
	}

	i := strings.Index(str, ":")
	if i < 0 {
		tag.beanName = str
		return
	}

	tag.typeName = str[:i]
	tag.beanName = str[i+1:]
	return
}

func (tag wireTag) String() string {
	b := bytes.NewBuffer(nil)
	if tag.typeName != "" {
		b.WriteString(tag.typeName)
		b.WriteString(":")
	}
	b.WriteString(tag.beanName)
	if tag.nullable {
		b.WriteString("?")
	}
	return b.String()
}

func toWireTag(selector gs.BeanSelector) wireTag {
	switch s := selector.(type) {
	case string:
		return parseWireTag(s)
	case gs.BeanDefinition:
		return parseWireTag(s.ID())
	case *gs.BeanDefinition:
		return parseWireTag(s.ID())
	default:
		return parseWireTag(gs_util.TypeName(s) + ":")
	}
}

func toWireString(tags []wireTag) string {
	var buf bytes.Buffer
	for i, tag := range tags {
		buf.WriteString(tag.String())
		if i < len(tags)-1 {
			buf.WriteByte(',')
		}
	}
	return buf.String()
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
		tag := toWireTag(selector)
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

// wireBean 对 bean 进行属性绑定和依赖注入，同时追踪其注入路径。如果 bean 有初始
// 化函数，则在注入完成之后执行其初始化函数。如果 bean 依赖了其他 bean，则首先尝试
// 实例化被依赖的 bean 然后对它们进行注入。
func (c *Container) wireBean(b *gs.BeanDefinition, stack *wiringStack) error {

	if b.GetStatus() == gs.Deleted {
		return fmt.Errorf("bean:%q have been deleted", b.ID())
	}

	// 运行时 Get 或者 Wire 会出现下面这种情况。
	if c.state == Refreshed && b.GetStatus() == gs.Wired {
		return nil
	}

	haveDestroy := false

	defer func() {
		if haveDestroy {
			stack.destroyers.Remove(stack.destroyers.Back())
		}
	}()

	// 记录注入路径上的销毁函数及其执行的先后顺序。
	if _, ok := b.Interface().(BeanDestroy); ok || b.GetDestroy() != nil {
		haveDestroy = true
		d := stack.saveDestroyer(b)
		if i := stack.destroyers.Back(); i != nil {
			d.after(i.Value.(*gs.BeanDefinition))
		}
		stack.destroyers.PushBack(b)
	}

	stack.pushBack(b)

	if b.GetStatus() == gs.Creating && b.F != nil {
		prev := stack.beans[len(stack.beans)-2]
		if prev.GetStatus() == gs.Creating {
			return errors.New("found circle autowire")
		}
	}

	if b.GetStatus() >= gs.Creating {
		stack.popBack()
		return nil
	}

	b.SetStatus(gs.Creating)

	// 对当前 bean 的间接依赖项进行注入。
	for _, s := range b.GetDepends() {
		beans, err := c.Find(s)
		if err != nil {
			return err
		}
		for _, d := range beans {
			err = c.wireBean(d, stack)
			if err != nil {
				return err
			}
		}
	}

	v, err := c.getBeanValue(b, stack)
	if err != nil {
		return err
	}

	b.SetStatus(gs.Created)

	t := v.Type()
	for _, typ := range b.GetExports() {
		if !t.Implements(typ) {
			return fmt.Errorf("%s doesn't implement interface %s", b, typ)
		}
	}

	err = c.wireBeanValue(v, t, stack)
	if err != nil {
		return err
	}

	if b.GetInit() != nil {
		fnValue := reflect.ValueOf(b.GetInit())
		out := fnValue.Call([]reflect.Value{b.Value()})
		if len(out) > 0 && !out[0].IsNil() {
			return out[0].Interface().(error)
		}
	}

	if f, ok := b.Interface().(BeanInit); ok {
		if err = f.OnInit(c); err != nil {
			return err
		}
	}

	b.SetStatus(gs.Wired)
	stack.popBack()
	return nil
}

type argContext struct {
	c     *Container
	stack *wiringStack
}

func (a *argContext) Matches(c gs.Condition) (bool, error) {
	return c.Matches(a.c)
}

func (a *argContext) Bind(v reflect.Value, tag string) error {
	return a.c.p.Data().Bind(v, conf.Tag(tag))
}

func (a *argContext) Wire(v reflect.Value, tag string) error {
	return a.c.wireByTag(v, tag, a.stack)
}

// getBeanValue 获取 bean 的值，如果是构造函数 bean 则执行其构造函数然后返回执行结果。
func (c *Container) getBeanValue(b *gs.BeanDefinition, stack *wiringStack) (reflect.Value, error) {

	if b.F == nil {
		return b.Value(), nil
	}

	out, err := b.F.Call(&argContext{c: c, stack: stack})
	if err != nil {
		return reflect.Value{}, err /* fmt.Errorf("%s:%s return error: %v", b.getClass(), b.ID(), err) */
	}

	// 构造函数的返回值为值类型时 b.Type() 返回其指针类型。
	if val := out[0]; gs_util.IsBeanType(val.Type()) {
		// 如果实现接口的是值类型，那么需要转换成指针类型然后再赋值给接口。
		if !val.IsNil() && val.Kind() == reflect.Interface && conf.IsValueType(val.Elem().Type()) {
			v := reflect.New(val.Elem().Type())
			v.Elem().Set(val.Elem())
			b.Value().Set(v)
		} else {
			b.Value().Set(val)
		}
	} else {
		b.Value().Elem().Set(val)
	}

	if b.Value().IsNil() {
		return reflect.Value{}, fmt.Errorf("%s:%q return nil", b.GetClass(), b.FileLine())
	}

	v := b.Value()
	// 结果以接口类型返回时需要将原始值取出来才能进行注入。
	if b.Type().Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v, nil
}

// wireBeanValue 对 v 进行属性绑定和依赖注入，v 在传入时应该是一个已经初始化的值。
func (c *Container) wireBeanValue(v reflect.Value, t reflect.Type, stack *wiringStack) error {

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	// 如整数指针类型的 bean 是无需注入的。
	if v.Kind() != reflect.Struct {
		return nil
	}

	typeName := t.Name()
	if typeName == "" { // 简单类型没有名字
		typeName = t.String()
	}

	param := conf.BindParam{Path: typeName}
	return c.wireStruct(v, t, param, stack)
}

// wireStruct 对结构体进行依赖注入，需要注意的是这里不需要进行属性绑定。
func (c *Container) wireStruct(v reflect.Value, t reflect.Type, opt conf.BindParam, stack *wiringStack) error {

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)

		if !fv.CanInterface() {
			fv = gs_util.PatchValue(fv)
			if !fv.CanInterface() {
				continue
			}
		}

		fieldPath := opt.Path + "." + ft.Name

		// 支持 autowire 和 inject 两个标签。
		tag, ok := ft.Tag.Lookup("autowire")
		if !ok {
			tag, ok = ft.Tag.Lookup("inject")
		}
		if ok {
			if strings.HasSuffix(tag, ",lazy") {
				f := lazyField{v: fv, path: fieldPath, tag: tag}
				stack.lazyFields = append(stack.lazyFields, f)
			} else {
				if ft.Type == contextType {
					c.ContextAware = true
				}
				if err := c.wireByTag(fv, tag, stack); err != nil {
					return fmt.Errorf("%q wired error: %w", fieldPath, err)
				}
			}
			continue
		}

		subParam := conf.BindParam{
			Key:  opt.Key,
			Path: fieldPath,
		}

		if tag, ok = ft.Tag.Lookup("value"); ok {
			// validateTag, _ := ft.Tag.Lookup(validate.TagName())
			if err := subParam.BindTag(tag, ""); err != nil {
				return err
			}
			if ft.Anonymous {
				err := c.wireStruct(fv, ft.Type, subParam, stack)
				if err != nil {
					return err
				}
			} else {
				err := c.p.BindValue(fv.Addr(), subParam)
				if err != nil {
					return err
				}
			}
			continue
		}

		if ft.Anonymous && ft.Type.Kind() == reflect.Struct {
			if err := c.wireStruct(fv, ft.Type, subParam, stack); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Container) wireByTag(v reflect.Value, tag string, stack *wiringStack) error {

	// tag 预处理，可能通过属性值进行指定。
	if strings.HasPrefix(tag, "${") {
		s, err := c.p.Data().Resolve(tag)
		if err != nil {
			return err
		}
		tag = s
	}

	if tag == "" {
		return c.autowire(v, nil, false, stack)
	}

	var tags []wireTag
	if tag != "?" {
		for _, s := range strings.Split(tag, ",") {
			tags = append(tags, toWireTag(s))
		}
	}
	return c.autowire(v, tags, tag == "?", stack)
}

func (c *Container) autowire(v reflect.Value, tags []wireTag, nullable bool, stack *wiringStack) error {
	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return c.collectBeans(v, tags, nullable, stack)
	default:
		var tag wireTag
		if len(tags) > 0 {
			tag = tags[0]
		} else if nullable {
			tag.nullable = true
		}
		return c.getBean(v, tag, stack)
	}
}

// getBean 获取 tag 对应的 bean 然后赋值给 v，因此 v 应该是一个未初始化的值。
func (c *Container) getBean(v reflect.Value, tag wireTag, stack *wiringStack) error {

	if !v.IsValid() {
		return fmt.Errorf("receiver must be ref type, bean:%q", tag)
	}

	t := v.Type()
	if !gs_util.IsBeanReceiver(t) {
		return fmt.Errorf("%s is not valid receiver type", t.String())
	}

	var foundBeans []*gs.BeanDefinition
	for _, b := range c.beansByType[t] {
		if b.GetStatus() == gs.Deleted {
			continue
		}
		if !b.Match(tag.typeName, tag.beanName) {
			continue
		}
		foundBeans = append(foundBeans, b)
	}

	// 指定 bean 名称时通过名称获取，防止未通过 Export 方法导出接口。
	if t.Kind() == reflect.Interface && tag.beanName != "" {
		for _, b := range c.beansByName[tag.beanName] {
			if b.GetStatus() == gs.Deleted {
				continue
			}
			if !b.Type().AssignableTo(t) {
				continue
			}
			if !b.Match(tag.typeName, tag.beanName) {
				continue
			}

			found := false // 对结果排重
			for _, r := range foundBeans {
				if r == b {
					found = true
					break
				}
			}
			if !found {
				foundBeans = append(foundBeans, b)
				// c.logger.Warnf("you should call Export() on %s", b)
			}
		}
	}

	if len(foundBeans) == 0 {
		if tag.nullable {
			return nil
		}
		return fmt.Errorf("can't find bean, bean:%q type:%q", tag, t)
	}

	// 优先使用设置成主版本的 bean
	var primaryBeans []*gs.BeanDefinition

	for _, b := range foundBeans {
		if b.IsPrimary() {
			primaryBeans = append(primaryBeans, b)
		}
	}

	if len(primaryBeans) > 1 {
		msg := fmt.Sprintf("found %d primary beans, bean:%q type:%q [", len(primaryBeans), tag, t)
		for _, b := range primaryBeans {
			msg += "( " + b.String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return errors.New(msg)
	}

	if len(primaryBeans) == 0 && len(foundBeans) > 1 {
		msg := fmt.Sprintf("found %d beans, bean:%q type:%q [", len(foundBeans), tag, t)
		for _, b := range foundBeans {
			msg += "( " + b.String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return errors.New(msg)
	}

	var result *gs.BeanDefinition
	if len(primaryBeans) == 1 {
		result = primaryBeans[0]
	} else {
		result = foundBeans[0]
	}

	// 确保找到的 bean 已经完成依赖注入。
	err := c.wireBean(result, stack)
	if err != nil {
		return err
	}

	v.Set(result.Value())
	return nil
}

// filterBean 返回 tag 对应的 bean 在数组中的索引，找不到返回 -1。
func filterBean(beans []*gs.BeanDefinition, tag wireTag, t reflect.Type) (int, error) {

	var found []int
	for i, b := range beans {
		if b.Match(tag.typeName, tag.beanName) {
			found = append(found, i)
		}
	}

	if len(found) > 1 {
		msg := fmt.Sprintf("found %d beans, bean:%q type:%q [", len(found), tag, t)
		for _, i := range found {
			msg += "( " + beans[i].String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return -1, errors.New(msg)
	}

	if len(found) > 0 {
		i := found[0]
		return i, nil
	}

	if tag.nullable {
		return -1, nil
	}

	return -1, fmt.Errorf("can't find bean, bean:%q type:%q", tag, t)
}

type byOrder []*gs.BeanDefinition

func (b byOrder) Len() int           { return len(b) }
func (b byOrder) Less(i, j int) bool { return b[i].GetOrder() < b[j].GetOrder() }
func (b byOrder) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

func (c *Container) collectBeans(v reflect.Value, tags []wireTag, nullable bool, stack *wiringStack) error {

	t := v.Type()
	if t.Kind() != reflect.Slice && t.Kind() != reflect.Map {
		return fmt.Errorf("should be slice or map in collection mode")
	}

	et := t.Elem()
	if !gs_util.IsBeanReceiver(et) {
		return fmt.Errorf("%s is not valid receiver type", t.String())
	}

	var beans []*gs.BeanDefinition
	if et.Kind() == reflect.Interface && et.NumMethod() == 0 {
		beans = c.beans
	} else {
		beans = c.beansByType[et]
	}

	{
		var arr []*gs.BeanDefinition
		for _, b := range beans {
			if b.GetStatus() == gs.Deleted {
				continue
			}
			arr = append(arr, b)
		}
		beans = arr
	}

	if len(tags) > 0 {

		var (
			anyBeans  []*gs.BeanDefinition
			afterAny  []*gs.BeanDefinition
			beforeAny []*gs.BeanDefinition
		)

		foundAny := false
		for _, item := range tags {

			// 是否遇到了"无序"标记
			if item.beanName == "*" {
				if foundAny {
					return fmt.Errorf("more than one * in collection %q", tags)
				}
				foundAny = true
				continue
			}

			index, err := filterBean(beans, item, et)
			if err != nil {
				return err
			}
			if index < 0 {
				continue
			}

			if foundAny {
				afterAny = append(afterAny, beans[index])
			} else {
				beforeAny = append(beforeAny, beans[index])
			}

			tmpBeans := append([]*gs.BeanDefinition{}, beans[:index]...)
			beans = append(tmpBeans, beans[index+1:]...)
		}

		if foundAny {
			anyBeans = append(anyBeans, beans...)
		}

		n := len(beforeAny) + len(anyBeans) + len(afterAny)
		arr := make([]*gs.BeanDefinition, 0, n)
		arr = append(arr, beforeAny...)
		arr = append(arr, anyBeans...)
		arr = append(arr, afterAny...)
		beans = arr
	}

	if len(beans) == 0 && !nullable {
		if len(tags) == 0 {
			return fmt.Errorf("no beans collected for %q", toWireString(tags))
		}
		for _, tag := range tags {
			if !tag.nullable {
				return fmt.Errorf("no beans collected for %q", toWireString(tags))
			}
		}
		return nil
	}

	for _, b := range beans {
		if err := c.wireBean(b, stack); err != nil {
			return err
		}
	}

	var ret reflect.Value
	switch t.Kind() {
	case reflect.Slice:
		sort.Sort(byOrder(beans))
		ret = reflect.MakeSlice(t, 0, 0)
		for _, b := range beans {
			ret = reflect.Append(ret, b.Value())
		}
	case reflect.Map:
		ret = reflect.MakeMap(t)
		for _, b := range beans {
			ret.SetMapIndex(reflect.ValueOf(b.GetName()), b.Value())
		}
	default:
	}
	v.Set(ret)
	return nil
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
		tags = append(tags, toWireTag(s))
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

func (c *Container) Invoke(fn interface{}, args ...gs.Arg) ([]interface{}, error) {

	if !gs_util.IsFuncType(reflect.TypeOf(fn)) {
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
