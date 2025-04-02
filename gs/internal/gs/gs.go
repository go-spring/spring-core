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

//go:generate mockgen -build_flags="-mod=mod" -package=gs -source=gs.go -destination=gs_mock.go

package gs

import (
	"context"
	"reflect"
	"strings"
	"unsafe"

	"github.com/go-spring/spring-core/conf"
)

// anyType is the [reflect.Type] of the [any] type.
var anyType = reflect.TypeFor[any]()

// As returns the [reflect.Type] of the given interface type.
// It ensures that the provided generic type parameter T is an interface.
// If T is not an interface, the function panics.
func As[T any]() reflect.Type {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Interface {
		panic("T must be interface")
	}
	return t
}

// BeanSelector is an interface for selecting beans.
type BeanSelector interface {
	// TypeAndName returns the type and name of the bean.
	TypeAndName() (reflect.Type, string)
}

// BeanSelectorImpl is an implementation of BeanSelector.
type BeanSelectorImpl struct {
	Type reflect.Type // The type of the bean
	Name string       // The name of the bean
}

// BeanSelectorFor returns a BeanSelectorImpl for the given type.
// If a name is provided, it is set; otherwise, only the type is used.
func BeanSelectorFor[T any](name ...string) BeanSelector {
	if len(name) == 0 {
		return BeanSelectorImpl{Type: reflect.TypeFor[T]()}
	}
	return BeanSelectorImpl{Type: reflect.TypeFor[T](), Name: name[0]}
}

// TypeAndName returns the type and name of the bean.
func (s BeanSelectorImpl) TypeAndName() (reflect.Type, string) {
	return s.Type, s.Name
}

func (s BeanSelectorImpl) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	if s.Type != nil && s.Type != anyType {
		sb.WriteString("Type:")
		sb.WriteString(s.Type.String())
	}
	if s.Name != "" {
		if sb.Len() > 1 {
			sb.WriteString(",")
		}
		sb.WriteString("Name:")
		sb.WriteString(s.Name)
	}
	sb.WriteString("}")
	return sb.String()
}

/************************************ cond ***********************************/

// Condition is an interface used for defining conditional logic
// when registering beans in the IoC container.
type Condition interface {
	// Matches checks whether the condition is satisfied.
	Matches(ctx CondContext) (bool, error)
}

// CondBean represents a bean with Name and Type.
type CondBean interface {
	Name() string
	Type() reflect.Type
}

// CondContext defines methods for the IoC container used by conditions.
type CondContext interface {
	// Has checks whether the IoC container has a property with the given key.
	Has(key string) bool
	// Prop retrieves the value of a property from the IoC container.
	Prop(key string, def ...string) string
	// Find searches for bean definitions matching the given BeanSelector.
	Find(s BeanSelector) ([]CondBean, error)
}

// CondFunc is a function type that determines whether a condition is satisfied.
type CondFunc func(ctx CondContext) (bool, error)

/************************************* arg ***********************************/

// Arg is an interface for retrieving argument values in function parameter binding.
type Arg interface {
	// GetArgValue retrieves the argument value based on the type.
	GetArgValue(ctx ArgContext, t reflect.Type) (reflect.Value, error)
}

// ArgContext defines methods for the IoC container used by Arg types.
type ArgContext interface {
	// Check checks if the given condition is met.
	Check(c Condition) (bool, error)
	// Bind binds property values to the provided [reflect.Value].
	Bind(v reflect.Value, tag string) error
	// Wire wires dependent beans to the provided [reflect.Value].
	Wire(v reflect.Value, tag string) error
}

/*********************************** dync ************************************/

// Refreshable represents an object that can be dynamically refreshed.
type Refreshable interface {
	// OnRefresh is called to refresh the properties when they change.
	OnRefresh(prop conf.Properties, param conf.BindParam) error
}

/*********************************** app ************************************/

// Runner defines an interface for components that should run after
// all beans are injected but before the application servers start.
type Runner interface {
	Run() error
}

// Job defines an interface for components that run tasks with a given context
// after all beans are injected but before the application servers start.
type Job interface {
	Run(ctx context.Context) error
}

// ReadySignal defines an interface for components that can trigger a signal
// when the application is ready to serve requests.
type ReadySignal interface {
	TriggerAndWait() <-chan struct{}
}

// Server defines an interface for managing the lifecycle of application servers,
// such as HTTP, gRPC, Thrift, or MQ servers. It includes methods for starting
// and shutting down the server gracefully.
type Server interface {
	ListenAndServe(sig ReadySignal) error
	Shutdown(ctx context.Context) error
}

/*********************************** bean ************************************/

// BeanInitFunc defines the prototype for initialization functions.
// Examples: `func(bean)` or `func(bean) error`.
type BeanInitFunc = interface{}

// BeanDestroyFunc defines the prototype for destruction functions.
// Examples: `func(bean)` or `func(bean) error`.
type BeanDestroyFunc = interface{}

// Configuration holds parameters for bean setup configuration.
type Configuration struct {
	Includes []string // Methods to include
	Excludes []string // Methods to exclude
}

// BeanRegistration provides methods for configuring and registering bean metadata.
type BeanRegistration interface {
	Name() string
	Type() reflect.Type
	Value() reflect.Value
	SetName(name string)
	SetInit(fn BeanInitFunc)
	SetDestroy(fn BeanDestroyFunc)
	SetInitMethod(method string)
	SetDestroyMethod(method string)
	SetCondition(c ...Condition)
	SetDependsOn(selectors ...BeanSelector)
	SetExport(exports ...reflect.Type)
	SetConfiguration(c ...Configuration)
	SetRefreshable(tag string)
	SetCaller(skip int)
	OnProfiles(profiles string)
}

// beanBuilder helps configure a bean during its creation.
type beanBuilder[T any] struct {
	b BeanRegistration
}

// TypeAndName returns the type and name of the bean.
func (d *beanBuilder[T]) TypeAndName() (reflect.Type, string) {
	return d.b.Type(), d.b.Name()
}

// GetArgValue returns the value of the bean.
func (d *beanBuilder[T]) GetArgValue(ctx ArgContext, t reflect.Type) (reflect.Value, error) {
	return d.b.Value(), nil
}

// BeanRegistration returns the underlying BeanRegistration instance.
func (d *beanBuilder[T]) BeanRegistration() BeanRegistration {
	return d.b
}

// Name sets the name of the bean.
func (d *beanBuilder[T]) Name(name string) *T {
	d.b.SetName(name)
	return *(**T)(unsafe.Pointer(&d))
}

// Init sets the initialization function for the bean.
func (d *beanBuilder[T]) Init(fn BeanInitFunc) *T {
	d.b.SetInit(fn)
	return *(**T)(unsafe.Pointer(&d))
}

// Destroy sets the destruction function for the bean.
func (d *beanBuilder[T]) Destroy(fn BeanDestroyFunc) *T {
	d.b.SetDestroy(fn)
	return *(**T)(unsafe.Pointer(&d))
}

// InitMethod sets the initialization function for the bean by method name.
func (d *beanBuilder[T]) InitMethod(method string) *T {
	d.b.SetInitMethod(method)
	return *(**T)(unsafe.Pointer(&d))
}

// DestroyMethod sets the destruction function for the bean by method name.
func (d *beanBuilder[T]) DestroyMethod(method string) *T {
	d.b.SetDestroyMethod(method)
	return *(**T)(unsafe.Pointer(&d))
}

// Condition sets the conditions for the bean.
func (d *beanBuilder[T]) Condition(c ...Condition) *T {
	d.b.SetCondition(c...)
	return *(**T)(unsafe.Pointer(&d))
}

// DependsOn sets the beans that this bean depends on.
func (d *beanBuilder[T]) DependsOn(selectors ...BeanSelector) *T {
	d.b.SetDependsOn(selectors...)
	return *(**T)(unsafe.Pointer(&d))
}

// AsRunner marks the bean as a Runner.
func (d *beanBuilder[T]) AsRunner() *T {
	d.b.SetExport(As[Runner]())
	return *(**T)(unsafe.Pointer(&d))
}

// AsJob marks the bean as a Job.
func (d *beanBuilder[T]) AsJob() *T {
	d.b.SetExport(As[Job]())
	return *(**T)(unsafe.Pointer(&d))
}

// AsServer marks the bean as a Server.
func (d *beanBuilder[T]) AsServer() *T {
	d.b.SetExport(As[Server]())
	return *(**T)(unsafe.Pointer(&d))
}

// Export sets the interfaces that the bean will export.
func (d *beanBuilder[T]) Export(exports ...reflect.Type) *T {
	d.b.SetExport(exports...)
	return *(**T)(unsafe.Pointer(&d))
}

// Configuration applies the configuration parameters to the bean.
func (d *beanBuilder[T]) Configuration(c ...Configuration) *T {
	d.b.SetConfiguration(c...)
	return *(**T)(unsafe.Pointer(&d))
}

// Refreshable marks the bean as refreshable with the provided tag.
func (d *beanBuilder[T]) Refreshable(tag string) *T {
	d.b.SetRefreshable(tag)
	return *(**T)(unsafe.Pointer(&d))
}

// Caller sets the caller information for the bean.
func (d *beanBuilder[T]) Caller(skip int) *T {
	d.b.SetCaller(skip)
	return *(**T)(unsafe.Pointer(&d))
}

// OnProfiles sets the profiles that the bean will be active in.
func (d *beanBuilder[T]) OnProfiles(profiles string) *T {
	d.b.OnProfiles(profiles)
	return *(**T)(unsafe.Pointer(&d))
}

// RegisteredBean represents a bean that has been registered in the IoC container.
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
