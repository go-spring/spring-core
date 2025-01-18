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
	"github.com/go-spring/spring-core/util"
)

// A BeanSelector can be the ID of a bean, a `reflect.Type`, an interface{}
// pointer such as `(*error)(nil)`, or `new(error)`.
type BeanSelector interface{}

// BeanSelectorToString returns the string representation of the bean selector.
func BeanSelectorToString(s BeanSelector) string {
	switch v := s.(type) {
	case string:
		return v
	default:
		return util.TypeName(s) + ":"
	}
}

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

// CondFunc is a function that returns true when the condition is met.
type CondFunc func(ctx CondContext) (bool, error)

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

// BeanRegistration provides methods for configuring bean metadata.
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

// beanBuilder helps configure a bean during its creation.
type beanBuilder[T any] struct {
	b BeanRegistration
}

// BeanRegistration returns the underlying BeanRegistration instance.
func (d *beanBuilder[T]) BeanRegistration() BeanRegistration {
	return d.b
}

// ID returns the unique identifier of the bean.
func (d *beanBuilder[T]) ID() string {
	return d.b.ID()
}

// Type returns the [reflect.Type] of the bean.
func (d *beanBuilder[T]) Type() reflect.Type {
	return d.b.Type()
}

// Name sets the name of the bean.
func (d *beanBuilder[T]) Name(name string) *T {
	d.b.SetName(name)
	return *(**T)(unsafe.Pointer(&d))
}

// Caller sets the caller information for the bean.
func (d *beanBuilder[T]) Caller(skip int) *T {
	d.b.SetCaller(skip)
	return *(**T)(unsafe.Pointer(&d))
}

// Condition sets the condition of the bean.
func (d *beanBuilder[T]) Condition(cond Condition) *T {
	d.b.SetCondition(cond)
	return *(**T)(unsafe.Pointer(&d))
}

// DependsOn sets the dependencies for the bean.
func (d *beanBuilder[T]) DependsOn(selectors ...BeanSelector) *T {
	d.b.SetDependsOn(selectors...)
	return *(**T)(unsafe.Pointer(&d))
}

// Primary marks the bean as primary.
func (d *beanBuilder[T]) Primary() *T {
	d.b.SetPrimary()
	return *(**T)(unsafe.Pointer(&d))
}

// Init sets the initialization function.
func (d *beanBuilder[T]) Init(fn interface{}) *T {
	d.b.SetInit(fn)
	return *(**T)(unsafe.Pointer(&d))
}

// Destroy sets the destroy function.
func (d *beanBuilder[T]) Destroy(fn interface{}) *T {
	d.b.SetDestroy(fn)
	return *(**T)(unsafe.Pointer(&d))
}

// Export sets the interfaces to export.
func (d *beanBuilder[T]) Export(exports ...interface{}) *T {
	d.b.SetExport(exports...)
	return *(**T)(unsafe.Pointer(&d))
}

func (d *beanBuilder[T]) Configuration(param ...ConfigurationParam) *T {
	d.b.SetConfiguration(param...)
	return *(**T)(unsafe.Pointer(&d))
}

func (d *beanBuilder[T]) EnableRefresh(tag string) *T {
	d.b.SetEnableRefresh(tag)
	return *(**T)(unsafe.Pointer(&d))
}

// RegisteredBean represents a bean that has been registered.
type RegisteredBean struct {
	beanBuilder[RegisteredBean]
}

// NewRegisteredBean creates a new RegisteredBean instance.
func NewRegisteredBean(d BeanRegistration) *RegisteredBean {
	return &RegisteredBean{
		beanBuilder: beanBuilder[RegisteredBean]{d},
	}
}

// BeanDefinition represents a bean that has not yet been registered.
type BeanDefinition struct {
	beanBuilder[BeanDefinition]
}

// NewBeanDefinition creates a new BeanDefinition instance.
func NewBeanDefinition(d BeanRegistration) *BeanDefinition {
	return &BeanDefinition{
		beanBuilder: beanBuilder[BeanDefinition]{d},
	}
}

/************************************ ioc ************************************/

// Container represents the modifiable aspects of an IoC container.
type Container interface {

	// Object registers a bean with the given object instance.
	Object(i interface{}) *RegisteredBean

	// Provide registers a bean using the given constructor function.
	Provide(ctor interface{}, args ...Arg) *RegisteredBean

	// Register registers a bean using the given bean definition.
	Register(b *BeanDefinition) *RegisteredBean

	// GroupRegister registers beans by executing the given function.
	GroupRegister(fn func(p Properties) ([]*BeanDefinition, error))

	// RefreshProperties updates the properties of the container.
	RefreshProperties(p Properties) error

	// Refresh initializes and wires all beans in the container.
	Refresh() error

	// ReleaseUnusedMemory releases unused memory by cleaning up unnecessary resources.
	ReleaseUnusedMemory()

	// Close closes the container and cleans up resources.
	Close()
}

// Context represents the unmodifiable (or runtime) aspects of an IoC container.
type Context interface {

	// Context returns the root context.Context of the container.
	Context() context.Context

	// Keys returns all keys present in the container's properties.
	Keys() []string

	// Has checks if a key exists in the container's properties.
	Has(key string) bool

	// SubKeys returns sub-keys under the specified key in the container's properties.
	SubKeys(key string) ([]string, error)

	// Prop retrieves the value of the specified key from the container's properties.
	Prop(key string, opts ...conf.GetOption) string

	// Resolve resolves placeholders or references in the given string.
	Resolve(s string) (string, error)

	// Bind binds the value of the specified key to the provided struct or variable.
	Bind(i interface{}, opts ...conf.BindArg) error

	// Get retrieves a bean of the specified type using the provided selectors.
	Get(i interface{}, selectors ...BeanSelector) error

	// Wire creates and returns a wired bean using the provided object or constructor function.
	Wire(objOrCtor interface{}, ctorArgs ...Arg) (interface{}, error)

	// Invoke calls the provided function with the specified arguments.
	Invoke(fn interface{}, args ...Arg) ([]interface{}, error)

	// Go runs the provided function in a new goroutine. When the container is closed,
	// the context.Context will be canceled.
	Go(fn func(ctx context.Context))
}

// ContextAware is used to inject the gs.Context into a bean.
type ContextAware struct {
	GSContext Context `autowire:""`
}
