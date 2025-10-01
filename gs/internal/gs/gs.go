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

//go:generate gs mock -o=gs_mock.go -i=ConditionContext,ArgContext,Server

package gs

import (
	"context"
	"reflect"
	"strings"
	"unsafe"
)

// anyType is the [reflect.Type] of the [any] type.
var anyType = reflect.TypeFor[any]()

// As returns the [reflect.Type] of the given generic interface type T.
// It ensures that T is an interface type; otherwise, it panics.
func As[T any]() reflect.Type {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Interface {
		panic("T must be interface")
	}
	return t
}

// BeanSelector is an abstraction that represents a way to select beans
// within the IoC container. It identifies a bean by its type and optionally its name.
type BeanSelector interface {
	// TypeAndName returns the [reflect.Type] and name that uniquely identify the bean.
	TypeAndName() (reflect.Type, string)
}

// BeanSelectorImpl is a concrete implementation of BeanSelector.
type BeanSelectorImpl struct {
	Type reflect.Type // The [reflect.Type] of the bean
	Name string       // The optional name of the bean
}

// BeanSelectorFor creates a BeanSelector for a specific type T.
// If a name is provided, it will be associated with the selector;
// otherwise, only the type is used to identify the bean.
func BeanSelectorFor[T any](name ...string) BeanSelector {
	if len(name) == 0 {
		return BeanSelectorImpl{Type: reflect.TypeFor[T]()}
	}
	return BeanSelectorImpl{Type: reflect.TypeFor[T](), Name: name[0]}
}

// TypeAndName returns the type and name of the bean selector.
func (s BeanSelectorImpl) TypeAndName() (reflect.Type, string) {
	return s.Type, s.Name
}

// String returns a human-readable string representation of the selector.
// Example: "{Type:*mypkg.MyBean,Name:myBeanInstance}"
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

// Condition defines a contract for conditional bean registration.
// A Condition can decide at runtime whether a particular bean should be registered.
type Condition interface {
	// Matches evaluates the condition against the given ConditionContext.
	// It returns true if the condition is satisfied.
	Matches(ctx ConditionContext) (bool, error)
}

// ConditionBean represents a bean in the IoC container that can be queried by conditions.
type ConditionBean interface {
	Name() string       // Returns the bean's name
	Type() reflect.Type // Returns the bean's type
}

// ConditionContext provides access to the IoC container for conditions.
// Conditions can query properties or find beans in the container.
type ConditionContext interface {
	// Has checks if a property with the given key exists in the IoC container.
	Has(key string) bool
	// Prop retrieves a property value from the IoC container with an optional default.
	Prop(key string, def ...string) string
	// Find searches for beans that match the given BeanSelector.
	Find(s BeanSelector) ([]ConditionBean, error)
}

/************************************* arg ***********************************/

// Arg defines an interface for resolving arguments used in dependency injection.
// It determines how to obtain values for function or method parameters.
type Arg interface {
	// GetArgValue retrieves the argument value for the given type
	// using the provided ArgContext.
	GetArgValue(ctx ArgContext, t reflect.Type) (reflect.Value, error)
}

// ArgContext provides the runtime context for resolving arguments.
// It allows checking conditions, binding properties, and wiring dependencies.
type ArgContext interface {
	// Check evaluates whether a given condition is satisfied.
	Check(c Condition) (bool, error)
	// Bind binds configuration or property values into the provided [reflect.Value].
	Bind(v reflect.Value, tag string) error
	// Wire injects dependencies (beans) into the provided [reflect.Value].
	Wire(v reflect.Value, tag string) error
}

/*********************************** app ************************************/

// Runner is an interface for components that need to run
// after all beans have been injected but before the application’s servers start.
type Runner interface {
	Run() error
}

// FuncRunner is a function type adapter that allows ordinary functions
// to be used as Runner components.
type FuncRunner func() error

func (f FuncRunner) Run() error {
	return f()
}

// Job is similar to Runner but allows passing a context to the task.
// It is typically used for background tasks or setup work that may be cancellable.
type Job interface {
	Run(ctx context.Context) error
}

// FuncJob is a function type adapter for the Job interface.
type FuncJob func(ctx context.Context) error

func (f FuncJob) Run(ctx context.Context) error {
	return f(ctx)
}

// ReadySignal represents a synchronization mechanism that signals
// when the application is ready to accept requests.
type ReadySignal interface {
	TriggerAndWait() <-chan struct{}
}

// Server defines the lifecycle of application servers (e.g., HTTP, gRPC).
// It provides methods for starting and gracefully shutting down the server.
type Server interface {
	ListenAndServe(sig ReadySignal) error
	Shutdown(ctx context.Context) error
}

/*********************************** bean ************************************/

// BeanMock represents a mocked bean instance that can replace a real bean
// for testing purposes.
type BeanMock struct {
	Object any          // The mock instance
	Target BeanSelector // The selector identifying the bean to replace
}

// BeanID uniquely identifies a bean in the IoC container by type and name.
type BeanID struct {
	Type reflect.Type // The bean's type
	Name string       // The bean's name
}

// BeanInitFunc defines the type for bean initialization functions.
// Example: `func(bean)` or `func(bean) error`.
type BeanInitFunc = any

// BeanDestroyFunc defines the type for bean destruction (cleanup) functions.
// Example: `func(bean)` or `func(bean) error`.
type BeanDestroyFunc = any

// Configuration specifies parameters for configuring beans during registration.
type Configuration struct {
	Includes []string // Methods to include
	Excludes []string // Methods to exclude
}

// BeanRegistration defines the API for configuring and registering a bean’s metadata
// in the IoC container.
type BeanRegistration interface {
	Name() string
	Type() reflect.Type
	Value() reflect.Value
	SetName(name string)
	SetInit(fn BeanInitFunc)
	SetDestroy(fn BeanDestroyFunc)
	SetInitMethod(method string)
	SetDestroyMethod(method string)
	SetCondition(conditions ...Condition)
	SetDependsOn(selectors ...BeanSelector)
	SetExport(exports ...reflect.Type)
	SetConfiguration(c ...Configuration)
	SetCaller(skip int)
	OnProfiles(profiles string)
}

// beanBuilder is a generic helper for configuring beans during their creation.
type beanBuilder[T any] struct {
	b BeanRegistration
}

// TypeAndName returns the bean’s type and name.
func (d *beanBuilder[T]) TypeAndName() (reflect.Type, string) {
	return d.b.Type(), d.b.Name()
}

// GetArgValue returns the bean’s value for argument injection.
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
func (d *beanBuilder[T]) Configuration(c ...Configuration) *T {
	d.b.SetConfiguration(c...)
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

// BeanDefinition represents a bean that is defined but not yet registered in the IoC container.
type BeanDefinition struct {
	beanBuilder[BeanDefinition]
}

// NewBeanDefinition creates a new BeanDefinition with the provided BeanRegistration.
func NewBeanDefinition(d BeanRegistration) *BeanDefinition {
	return &BeanDefinition{
		beanBuilder: beanBuilder[BeanDefinition]{d},
	}
}

// RegisteredBean represents a bean that has already been registered in the IoC container.
type RegisteredBean struct {
	beanBuilder[RegisteredBean]
}

// NewRegisteredBean creates a new RegisteredBean with the provided BeanRegistration.
func NewRegisteredBean(d BeanRegistration) *RegisteredBean {
	return &RegisteredBean{
		beanBuilder: beanBuilder[RegisteredBean]{d},
	}
}
