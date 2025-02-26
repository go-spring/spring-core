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
	"reflect"
	"strings"
	"unsafe"

	"github.com/go-spring/spring-core/conf"
)

// BeanInitFunc defines the prototype for initialization functions.
// For example: `func(bean)` or `func(bean) error`.
type BeanInitFunc = interface{}

// BeanDestroyFunc defines the prototype for destroy functions.
// For example: `func(bean)` or `func(bean) error`.
type BeanDestroyFunc = interface{}

// BeanInitInterface defines an interface for bean initialization.
type BeanInitInterface interface {
	OnBeanInit(ctx Context) error
}

// BeanDestroyInterface defines an interface for bean destruction.
type BeanDestroyInterface interface {
	OnBeanDestroy()
}

// BeanSelectorInterface is an interface for selecting beans.
type BeanSelectorInterface interface {
	TypeAndName() (reflect.Type, string)
}

// BeanSelector is an identifier for a bean.
type BeanSelector struct {
	Type reflect.Type // Type of the bean
	Name string       // Name of the bean
}

// BeanSelectorForType returns a BeanSelector for the given type.
func BeanSelectorForType[T any]() BeanSelector {
	return BeanSelector{Type: reflect.TypeFor[T]()}
}

// TypeAndName returns the type and name of the bean.
func (s BeanSelector) TypeAndName() (reflect.Type, string) {
	return s.Type, s.Name
}

func (s BeanSelector) String() string {
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
	Find(s BeanSelectorInterface) ([]CondBean, error)
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
	// SetCondition adds a condition for the bean.
	SetCondition(conditions ...Condition)
	// SetDependsOn sets the beans that this bean depends on.
	SetDependsOn(selectors ...BeanSelectorInterface)
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
	r := d.BeanRegistration()
	return r.Type(), r.Name()
}

// GetArgValue returns the value of the bean.
func (d *beanBuilder[T]) GetArgValue(ctx ArgContext, t reflect.Type) (reflect.Value, error) {
	return d.BeanRegistration().Value(), nil
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

// Condition adds a condition to validate the bean.
func (d *beanBuilder[T]) Condition(conditions ...Condition) *T {
	d.b.SetCondition(conditions...)
	return *(**T)(unsafe.Pointer(&d))
}

// DependsOn sets the beans that this bean depends on.
func (d *beanBuilder[T]) DependsOn(selectors ...BeanSelectorInterface) *T {
	d.b.SetDependsOn(selectors...)
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

/************************************ ioc ************************************/

// Container represents the modifiable aspects of an IoC (Inversion of Control) container.
// It provides methods for registering, refreshing, and managing beans within the container.
type Container interface {
	// Object registers a bean using the provided object instance.
	Object(i interface{}) *RegisteredBean
	// Provide registers a bean using the provided constructor function and optional arguments.
	Provide(ctor interface{}, args ...Arg) *RegisteredBean
	// Register registers a bean using the provided bean definition.
	Register(b *BeanDefinition) *RegisteredBean
	// GroupRegister registers multiple beans by executing the given function that returns [*BeanDefinition]s.
	GroupRegister(fn func(p Properties) ([]*BeanDefinition, error))
	// RefreshProperties updates the properties of the container.
	RefreshProperties(p Properties) error
	// Refresh initializes and wires all beans in the container.
	Refresh() error
	// ReleaseUnusedMemory cleans up unused resources and releases memory.
	ReleaseUnusedMemory()
	// Close shuts down the container and cleans up all resources.
	Close()
}

// Context represents the unmodifiable (or runtime) aspects of an IoC container.
// It provides methods for accessing properties, resolving values, and retrieving beans.
type Context interface {
	// Keys returns all the keys present in the container's properties.
	Keys() []string
	// Has checks if the specified key exists in the container's properties.
	Has(key string) bool
	// SubKeys returns the sub-keys under a specific key in the container's properties.
	SubKeys(key string) ([]string, error)
	// Prop retrieves the value of the specified key from the container's properties.
	Prop(key string, def ...string) string
	// Resolve resolves placeholders or references (e.g., ${KEY}) in the given string to actual values.
	Resolve(s string) (string, error)
	// Bind binds the value of the specified key to the provided struct or variable.
	Bind(i interface{}, tag ...string) error
	// Get retrieves a bean of the specified type using the provided tag.
	Get(i interface{}, tag ...string) error
	// Wire creates and returns a bean by wiring it with the provided constructor or object.
	Wire(objOrCtor interface{}, ctorArgs ...Arg) (interface{}, error)
	// Invoke calls the provided function with the specified arguments and returns the result.
	Invoke(fn interface{}, args ...Arg) ([]interface{}, error)
}

// ContextAware is used to inject the container's Context into a bean.
type ContextAware struct {
	GSContext Context `autowire:""`
}
