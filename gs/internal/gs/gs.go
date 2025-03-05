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

// Package gs provides all the concepts required for Go-Spring implementation.
package gs

import (
	"context"
	"reflect"
	"strings"
	"unsafe"

	"github.com/go-spring/spring-core/conf"
)

// As returns the [reflect.Type] of the given interface type.
func As[T any]() reflect.Type {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Interface {
		panic("T must be interface")
	}
	return t
}

// BeanInitFunc defines the prototype for initialization functions.
// For example: `func(bean)` or `func(bean) error`.
type BeanInitFunc = interface{}

// BeanDestroyFunc defines the prototype for destroy functions.
// For example: `func(bean)` or `func(bean) error`.
type BeanDestroyFunc = interface{}

// BeanSelector is an interface for selecting beans.
type BeanSelector interface {
	TypeAndName() (reflect.Type, string)
}

// BeanSelectorImpl is an identifier for a bean.
type BeanSelectorImpl struct {
	Type reflect.Type // Type of the bean
	Name string       // Name of the bean
}

// BeanSelectorFor returns a BeanSelectorImpl for the given type.
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
	if s.Type != nil {
		sb.WriteString("Type:")
		sb.WriteString(s.Type.String())
	}
	if s.Name != "" {
		if s.Type != nil {
			sb.WriteString(",")
		}
		sb.WriteString("Name:")
		sb.WriteString(s.Name)
	}
	sb.WriteString("}")
	return sb.String()
}

/********************************** condition ********************************/

// Condition is a conditional logic interface used when registering beans.
type Condition interface {
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
	// Find searches for bean definitions that match the provided BeanSelector.
	Find(s BeanSelector) ([]CondBean, error)
}

// CondFunc is a function type that determines whether a condition is satisfied.
type CondFunc func(ctx CondContext) (bool, error)

/************************************* arg ***********************************/

// Arg is used to provide binding values for function parameters.
type Arg interface {
	GetArgValue(ctx ArgContext, t reflect.Type) (reflect.Value, error)
}

// ArgContext defines methods for the IoC container used by Callable types.
type ArgContext interface {
	// Matches checks if the given condition is met.
	Matches(c Condition) (bool, error)
	// Bind binds property values to the provided [reflect.Value].
	Bind(v reflect.Value, tag string) error
	// Wire wires dependent beans to the provided [reflect.Value].
	Wire(v reflect.Value, tag string) error
}

// Callable represents an entity that can be invoked with an ArgContext.
type Callable interface {
	Call(ctx ArgContext) ([]reflect.Value, error)
}

/*********************************** conf ************************************/

// Properties represents read-only configuration properties.
type Properties = conf.ReadOnlyProperties

// Refreshable represents an object that can be dynamically refreshed.
type Refreshable interface {
	// OnRefresh is called to refresh the properties when they change.
	OnRefresh(prop Properties, param conf.BindParam) error
}

/*********************************** app ************************************/

// Runner defines an interface for runners that should be executed after all
// beans are injected but before the application's servers are started.
type Runner interface {
	Run()
}

// Job defines an interface for jobs that should be executed after all
// beans are injected but before the application's servers are started.
type Job interface {
	Run(ctx context.Context)
}

// Server defines an interface for managing the lifecycle of application servers,
// such as HTTP, gRPC, Thrift, or MQ servers. Servers must implement methods for
// starting and stopping gracefully.
type Server interface {
	Serve(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

/*********************************** bean ************************************/

// ConfigurationParam holds configuration parameters for bean setup.
type ConfigurationParam struct {
	Includes []string // List of methods to include
	Excludes []string // List of methods to exclude
}

// BeanRegistration provides methods to configure and register bean metadata.
type BeanRegistration interface {
	// Name returns the name of the bean.
	Name() string
	// Type returns the [reflect.Type] of the bean.
	Type() reflect.Type
	// Value returns the [reflect.Value] of the bean.
	Value() reflect.Value
	// SetName sets the name of the bean.
	SetName(name string)
	// SetInit sets the initialization function for the bean.
	SetInit(fn BeanInitFunc)
	// SetDestroy sets the destruction function for the bean.
	SetDestroy(fn BeanDestroyFunc)
	// SetInitMethod sets the initialization function for the bean by method name.
	SetInitMethod(method string)
	// SetDestroyMethod sets the destruction function for the bean by method name.
	SetDestroyMethod(method string)
	// SetCondition adds a condition for the bean.
	SetCondition(conditions ...Condition)
	// SetDependsOn sets the beans that this bean depends on.
	SetDependsOn(selectors ...BeanSelector)
	// SetExport defines the interfaces to be exported by the bean.
	SetExport(exports ...reflect.Type)
	// SetConfiguration applies the bean configuration.
	SetConfiguration(param ...ConfigurationParam)
	// SetRefreshable marks the bean as refreshable with the given tag.
	SetRefreshable(tag string)
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

// Condition adds a condition to validate the bean.
func (d *beanBuilder[T]) Condition(conditions ...Condition) *T {
	d.b.SetCondition(conditions...)
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
func (d *beanBuilder[T]) Configuration(param ...ConfigurationParam) *T {
	d.b.SetConfiguration(param...)
	return *(**T)(unsafe.Pointer(&d))
}

// Refreshable marks the bean as refreshable with the provided tag.
func (d *beanBuilder[T]) Refreshable(tag string) *T {
	d.b.SetRefreshable(tag)
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

// BeanDefinition represents a bean that has not yet been registered in the IoC container.
type BeanDefinition struct {
	beanBuilder[BeanDefinition]
}

// NewBeanDefinition creates a new BeanDefinition instance.
func NewBeanDefinition(d BeanRegistration) *BeanDefinition {
	return &BeanDefinition{
		beanBuilder: beanBuilder[BeanDefinition]{d},
	}
}
