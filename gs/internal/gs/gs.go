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

package gs

import (
	"context"
	"reflect"
	"unsafe"

	"github.com/go-spring/spring-core/conf"
)

// A BeanSelector can be the ID of a bean, a `reflect.Type`, an interface{}
// pointer such as `(*error)(nil)`, or `new(error)`.
type BeanSelector interface{}

/********************************** condition ********************************/

type CondBean interface {
	ID() string
	Name() string
	TypeName() string
	Type() reflect.Type
}

// CondContext defines some methods of IoC container that conditions use.
type CondContext interface {
	// Has returns whether the IoC container has a property.
	Has(key string) bool
	// Prop returns the property's value when the IoC container has it, or
	// returns empty string when the IoC container doesn't have it.
	Prop(key string, opts ...conf.GetOption) string
	// Find returns bean definitions that matched with the bean selector.
	Find(selector BeanSelector) ([]CondBean, error)
}

// Condition is used when registering a bean to determine whether it's valid.
type Condition interface {
	Matches(ctx CondContext) (bool, error)
}

/************************************* arg ***********************************/

// Arg 用于为函数参数提供绑定值。可以是 bean.Selector 类型，表示注入 bean ；
// 可以是 ${X:=Y} 形式的字符串，表示属性绑定或者注入 bean ；可以是 ValueArg
// 类型，表示不从 IoC 容器获取而是用户传入的普通值；可以是 IndexArg 类型，表示
// 带有下标的参数绑定；可以是 *optionArg 类型，用于为 Option 方法提供参数绑定。
type Arg interface{}

// ArgContext defines some methods of IoC container that Callable use.
type ArgContext interface {
	// Matches returns true when the Condition returns true,
	// and returns false when the Condition returns false.
	Matches(c Condition) (bool, error)
	// Bind binds properties value by the "value" tag.
	Bind(v reflect.Value, tag string) error
	// Wire wires dependent beans by the "autowire" tag.
	Wire(v reflect.Value, tag string) error
}

type Callable interface {
	Arg(i int) (Arg, bool)
	In(i int) (reflect.Type, bool)
	Call(ctx ArgContext) ([]reflect.Value, error)
}

/*********************************** conf ************************************/

type Properties interface {
	Data() map[string]string
	Keys() []string
	Has(key string) bool
	SubKeys(key string) ([]string, error)
	Get(key string, opts ...conf.GetOption) string
	Resolve(s string) (string, error)
	Bind(i interface{}, args ...conf.BindArg) error
	CopyTo(out *conf.Properties) error
}

// Refreshable 可动态刷新的对象
type Refreshable interface {
	OnRefresh(prop Properties, param conf.BindParam) error
}

/*********************************** bean ************************************/

type ConfigurationParam struct {
	Enable  bool     // 是否扫描成员方法
	Include []string // 包含哪些成员方法
	Exclude []string // 排除那些成员方法
}

type BeanRegistration interface {
	ID() string
	Type() reflect.Type
	SetCaller(skip int)
	SetName(name string)
	SetCondition(cond Condition)
	SetDependsOn(selectors ...BeanSelector)
	SetPrimary()
	SetInit(fn interface{})
	SetDestroy(fn interface{})
	SetExport(exports ...interface{})
	SetConfiguration(param ...ConfigurationParam)
	SetEnableRefresh(tag string)
}

type beanBuilder[T any] struct {
	b BeanRegistration
}

func (d *beanBuilder[T]) BeanRegistration() BeanRegistration {
	return d.b
}

func (d *beanBuilder[T]) ID() string {
	return d.b.ID()
}

func (d *beanBuilder[T]) Type() reflect.Type {
	return d.b.Type()
}

func (d *beanBuilder[T]) Name(name string) *T {
	d.b.SetName(name)
	return *(**T)(unsafe.Pointer(&d))
}

// Condition 设置 bean 的 Condition。
func (d *beanBuilder[T]) Condition(cond Condition) *T {
	d.b.SetCondition(cond)
	return *(**T)(unsafe.Pointer(&d))
}

// DependsOn 设置 bean 的间接依赖项。
func (d *beanBuilder[T]) DependsOn(selectors ...BeanSelector) *T {
	d.b.SetDependsOn(selectors...)
	return *(**T)(unsafe.Pointer(&d))
}

// Primary 设置 bean 为主版本。
func (d *beanBuilder[T]) Primary() *T {
	d.b.SetPrimary()
	return *(**T)(unsafe.Pointer(&d))
}

// Init 设置 bean 的初始化函数。
func (d *beanBuilder[T]) Init(fn interface{}) *T {
	d.b.SetInit(fn)
	return *(**T)(unsafe.Pointer(&d))
}

// Destroy 设置 bean 的销毁函数。
func (d *beanBuilder[T]) Destroy(fn interface{}) *T {
	d.b.SetDestroy(fn)
	return *(**T)(unsafe.Pointer(&d))
}

// Export 设置 bean 的导出接口。
func (d *beanBuilder[T]) Export(exports ...interface{}) *T {
	d.b.SetExport(exports...)
	return *(**T)(unsafe.Pointer(&d))
}

// Configuration 设置 bean 为配置类。
func (d *beanBuilder[T]) Configuration(param ...ConfigurationParam) *T {
	d.b.SetConfiguration(param...)
	return *(**T)(unsafe.Pointer(&d))
}

// EnableRefresh 设置 bean 为可刷新的。
func (d *beanBuilder[T]) EnableRefresh(tag string) *T {
	d.b.SetEnableRefresh(tag)
	return *(**T)(unsafe.Pointer(&d))
}

type RegisteredBean struct {
	beanBuilder[RegisteredBean]
}

func NewRegisteredBean(d BeanRegistration) *RegisteredBean {
	return &RegisteredBean{
		beanBuilder: beanBuilder[RegisteredBean]{d},
	}
}

type BeanDefinition struct {
	beanBuilder[BeanDefinition]
}

func NewBeanDefinition(d BeanRegistration) *BeanDefinition {
	return &BeanDefinition{
		beanBuilder: beanBuilder[BeanDefinition]{d},
	}
}

/*********************************** ioc ************************************/

// Container ...
type Container interface {
	Object(i interface{}) *RegisteredBean
	Provide(ctor interface{}, args ...Arg) *RegisteredBean
	Register(b *BeanDefinition) *RegisteredBean
	GroupRegister(fn func(p Properties) ([]*BeanDefinition, error))
	RefreshProperties(p Properties) error
	Refresh() error
	SimplifyMemory()
	Close()
}

// Context ...
type Context interface {
	Context() context.Context
	Keys() []string
	Has(key string) bool
	SubKeys(key string) ([]string, error)
	Prop(key string, opts ...conf.GetOption) string
	Resolve(s string) (string, error)
	Bind(i interface{}, opts ...conf.BindArg) error
	Get(i interface{}, selectors ...BeanSelector) error
	Wire(objOrCtor interface{}, ctorArgs ...Arg) (interface{}, error)
	Invoke(fn interface{}, args ...Arg) ([]interface{}, error)
	Go(fn func(ctx context.Context))
}

// ContextAware injects the Context into a struct as the field GSContext.
type ContextAware struct {
	GSContext Context `autowire:""`
}
