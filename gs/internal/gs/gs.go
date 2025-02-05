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
	"reflect"
	"unsafe"

	"github.com/go-spring/spring-core/conf"
)

// BeanSelector is an identifier for a bean.
type BeanSelector struct {
	Type reflect.Type // Type of the bean
	Tag  string       // Tag of the bean
}

func TagBeanSelector(tag string) BeanSelector {
	return BeanSelector{Tag: tag}
}

func TypeBeanSelector[T any]() BeanSelector {
	return BeanSelector{Type: reflect.TypeFor[T]()}
}

func (s BeanSelector) String() string {
	if s.Type == nil {
		return s.Tag
	}
	return "(" + s.Type.String() + ")" + s.Tag
}

/********************************** condition ********************************/

// CondBean represents a bean that has an ID, Name, TypeName, and Type.
type CondBean interface {
	ID() string
	Name() string
	TypeName() string
	Type() reflect.Type
}

// CondContext defines methods for the IoC container used by conditions.
type CondContext interface {
	// Has checks if the IoC container has a property with the given key.
	Has(key string) bool
	// Prop returns the value of a property from the IoC container,
	// or an empty string if it doesn't exist.
	Prop(key string) string
	// MustProp returns the value of a property from the IoC container,
	MustProp(key string, def string) string
	// Find searches for bean definitions matching the provided BeanSelector.
	Find(s BeanSelector) ([]CondBean, error)
}

// CondFunc defines a function that determines whether a condition is met.
type CondFunc func(ctx CondContext) (bool, error)

// Condition is a conditional logic interface used when registering beans.
type Condition interface {
	Matches(ctx CondContext) (bool, error)
}

/************************************* arg ***********************************/

// Arg is used to provide binding values for function parameters. It can be:
// - A BeanSelector type (for injecting beans),
// - A string in the form of ${X:=Y} (for property binding or bean injection),
// - A ValueArg type (for normal user-provided values),
// - An IndexArg type (for indexed parameter binding),
// - An *OptionArg type (for binding arguments in Option methods).
type Arg interface {
	Value() reflect.Value
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

type Properties = conf.ReadOnlyProperties

// Refreshable represents an object that can be dynamically refreshed.
type Refreshable interface {
	// OnRefresh is called to refresh the properties when they change.
	OnRefresh(prop Properties, param conf.BindParam) error
}

/*********************************** bean ************************************/

// ConfigurationParam holds configuration parameters for bean setup.
type ConfigurationParam struct {
	Includes []string // Methods to include
	Excludes []string // Methods to exclude
}

// BeanRegistration provides methods to configure and register bean metadata.
type BeanRegistration interface {
	// ID returns the unique identifier of the bean.
	ID() string
	// Type returns the [reflect.Type] of the bean.
	Type() reflect.Type
	// SetName sets the name of the bean.
	SetName(name string)
	// SetInit sets the initialization function for the bean.
	SetInit(fn interface{})
	// SetDestroy sets the destruction function for the bean.
	SetDestroy(fn interface{})
	// SetCondition adds a condition for the bean.
	SetCondition(cond Condition)
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

// Init sets the initialization function for the bean.
func (d *beanBuilder[T]) Init(fn interface{}) *T {
	d.b.SetInit(fn)
	return *(**T)(unsafe.Pointer(&d))
}

// Destroy sets the destruction function for the bean.
func (d *beanBuilder[T]) Destroy(fn interface{}) *T {
	d.b.SetDestroy(fn)
	return *(**T)(unsafe.Pointer(&d))
}

// Condition adds a condition to validate the bean.
func (d *beanBuilder[T]) Condition(cond Condition) *T {
	d.b.SetCondition(cond)
	return *(**T)(unsafe.Pointer(&d))
}

// DependsOn sets the beans this bean depends on.
func (d *beanBuilder[T]) DependsOn(selectors ...BeanSelector) *T {
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
// It provides methods for registering, refreshing, and managing beans in the container.
type Container interface {
	// Object registers a bean using the provided object instance.
	Object(i interface{}) *RegisteredBean

	// Provide registers a bean using the provided constructor function and optional arguments.
	Provide(ctor interface{}, args ...Arg) *RegisteredBean

	// Register registers a bean using the provided bean definition.
	Register(b *BeanDefinition) *RegisteredBean

	// GroupRegister registers multiple beans by executing the given function to return BeanDefinitions.
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
	Prop(key string) string

	// MustProp retrieves the value of the specified key from the container's properties.
	MustProp(key string, def string) string

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
