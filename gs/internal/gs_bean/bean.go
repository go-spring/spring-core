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

package gs_bean

import (
	"fmt"
	"reflect"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util"
)

// refreshableType is the [reflect.Type] of the interface [gs.Refreshable].
var refreshableType = reflect.TypeFor[gs.Refreshable]()

// BeanStatus represents the different lifecycle statuses of a bean.
type BeanStatus int8

const (
	StatusDeleted   = BeanStatus(-1)   // Bean has been deleted.
	StatusDefault   = BeanStatus(iota) // Default status of the bean.
	StatusResolving                    // Bean is being resolved.
	StatusResolved                     // Bean has been resolved.
	StatusCreating                     // Bean is being created.
	StatusCreated                      // Bean has been created.
	StatusWired                        // Bean has been wired.
)

// String returns a human-readable string of the bean status.
func (status BeanStatus) String() string {
	switch status {
	case StatusDeleted:
		return "deleted"
	case StatusDefault:
		return "default"
	case StatusResolving:
		return "resolving"
	case StatusResolved:
		return "resolved"
	case StatusCreating:
		return "creating"
	case StatusCreated:
		return "created"
	case StatusWired:
		return "wired"
	default:
		return "unknown"
	}
}

// BeanInit defines an interface for bean initialization.
type BeanInit interface {
	OnBeanInit(ctx gs.Context) error
}

// BeanDestroy defines an interface for bean destruction.
type BeanDestroy interface {
	OnBeanDestroy()
}

// BeanMetadata holds the metadata information of a bean.
type BeanMetadata struct {
	f          gs.Callable       // Callable for the bean (ctor or method).
	init       interface{}       // Initialization function for the bean.
	destroy    interface{}       // Destruction function for the bean.
	dependsOn  []gs.BeanSelector // List of dependencies for the bean.
	exports    []reflect.Type    // List of exported types for the bean.
	conditions []gs.Condition    // List of conditions for the bean.
	status     BeanStatus        // Current status of the bean.

	file string // The file where the bean is defined.
	line int    // The line number in the file where the bean is defined.

	configurationBean  bool                  // Whether the bean is a configuration bean.
	configurationParam gs.ConfigurationParam // Configuration parameters for the bean.

	refreshable bool   // Whether the bean can be refreshed.
	refreshTag  string // Refresh tag for the bean.
}

// Init returns the initialization function of the bean.
func (d *BeanMetadata) Init() interface{} {
	return d.init
}

// Destroy returns the destruction function of the bean.
func (d *BeanMetadata) Destroy() interface{} {
	return d.destroy
}

// DependsOn returns the list of dependencies for the bean.
func (d *BeanMetadata) DependsOn() []gs.BeanSelector {
	return d.dependsOn
}

// Exports returns the list of exported types for the bean.
func (d *BeanMetadata) Exports() []reflect.Type {
	return d.exports
}

// ConfigurationBean returns whether the bean is a configuration bean.
func (d *BeanMetadata) ConfigurationBean() bool {
	return d.configurationBean
}

// ConfigurationParam returns the configuration parameters for the bean.
func (d *BeanMetadata) ConfigurationParam() gs.ConfigurationParam {
	return d.configurationParam
}

// Refreshable returns whether the bean is refreshable.
func (d *BeanMetadata) Refreshable() bool {
	return d.refreshable
}

// RefreshTag returns the refresh tag of the bean.
func (d *BeanMetadata) RefreshTag() string {
	return d.refreshTag
}

// Conditions returns the list of conditions for the bean.
func (d *BeanMetadata) Conditions() []gs.Condition {
	return d.conditions
}

// File returns the file where the bean is defined.
func (d *BeanMetadata) File() string {
	return d.file
}

// Line returns the line number where the bean is defined.
func (d *BeanMetadata) Line() int {
	return d.line
}

// Class returns the class type of the bean.
func (d *BeanMetadata) Class() string {
	if d.f == nil {
		return "object bean"
	}
	return "constructor bean"
}

// BeanRuntime holds runtime information about the bean.
type BeanRuntime struct {
	v    reflect.Value // The value of the bean.
	t    reflect.Type  // The type of the bean.
	name string        // The name of the bean.
}

// Name returns the name of the bean.
func (d *BeanRuntime) Name() string {
	return d.name
}

// Type returns the type of the bean.
func (d *BeanRuntime) Type() reflect.Type {
	return d.t
}

// Value returns the value of the bean as [reflect.Value].
func (d *BeanRuntime) Value() reflect.Value {
	return d.v
}

// Interface returns the underlying value of the bean.
func (d *BeanRuntime) Interface() interface{} {
	return d.v.Interface()
}

// Status returns the current status of the bean.
func (d *BeanRuntime) Status() BeanStatus {
	return StatusWired
}

// Callable returns the callable for the bean.
func (d *BeanRuntime) Callable() gs.Callable {
	return nil
}

// Match checks if the bean matches the given typeName and beanName.
func (d *BeanRuntime) Match(beanName string) bool {
	nameIsSame := false
	if beanName == "" || d.name == beanName {
		nameIsSame = true
	}
	return nameIsSame
}

// String returns a string representation of the bean.
func (d *BeanRuntime) String() string {
	return d.name
}

// BeanDefinition contains both metadata and runtime information of a bean.
type BeanDefinition struct {
	*BeanMetadata
	*BeanRuntime
}

// Callable returns the callable for the bean.
func (d *BeanDefinition) Callable() gs.Callable {
	return d.f
}

// Status returns the current status of the bean.
func (d *BeanDefinition) Status() BeanStatus {
	return d.status
}

// SetStatus sets the current status of the bean.
func (d *BeanMetadata) SetStatus(status BeanStatus) {
	d.status = status
}

// SetFileLine sets the file and line number for the bean.
func (d *BeanDefinition) SetFileLine(file string, line int) {
	d.file, d.line = file, line
}

// SetName sets the name of the bean.
func (d *BeanDefinition) SetName(name string) {
	d.name = name
}

// validLifeCycleFunc checks whether the provided function is a valid lifecycle function.
func validLifeCycleFunc(fnType reflect.Type, beanType reflect.Type) bool {
	if !util.IsFuncType(fnType) || fnType.NumIn() != 1 {
		return false
	}
	if t := fnType.In(0); t.Kind() == reflect.Interface {
		if !beanType.Implements(t) {
			return false
		}
	} else if t != beanType {
		return false
	}
	return util.ReturnNothing(fnType) || util.ReturnOnlyError(fnType)
}

// SetInit sets the initialization function for the bean.
func (d *BeanDefinition) SetInit(fn interface{}) {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Type()) {
		d.init = fn
		return
	}
	panic("init should be func(bean) or func(bean)error")
}

// SetDestroy sets the destruction function for the bean.
func (d *BeanDefinition) SetDestroy(fn interface{}) {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Type()) {
		d.destroy = fn
		return
	}
	panic("destroy should be func(bean) or func(bean)error")
}

// SetCondition adds a condition to the list of conditions for the bean.
func (d *BeanDefinition) SetCondition(cond ...gs.Condition) {
	d.conditions = append(d.conditions, cond...)
}

// SetDependsOn sets the list of dependencies for the bean.
func (d *BeanDefinition) SetDependsOn(selectors ...gs.BeanSelector) {
	d.dependsOn = append(d.dependsOn, selectors...)
}

// SetExport sets the exported interfaces for the bean.
func (d *BeanDefinition) SetExport(exports ...reflect.Type) {
	for _, t := range exports {
		if t.Kind() != reflect.Interface {
			panic("only interface type can be exported")
		}
		exported := false
		for _, export := range d.exports {
			if t == export {
				exported = true
				break
			}
		}
		if exported {
			continue
		}
		d.exports = append(d.exports, t)
	}
}

// SetConfiguration sets the configuration flag and parameters for the bean.
func (d *BeanDefinition) SetConfiguration(param ...gs.ConfigurationParam) {
	d.configurationBean = true
	if len(param) == 0 {
		return
	}
	x := param[0]
	if len(x.Includes) > 0 {
		d.configurationParam.Includes = x.Includes
	}
	if len(x.Excludes) > 0 {
		d.configurationParam.Excludes = x.Excludes
	}
}

// SetRefreshable sets the refreshable flag and tag for the bean.
func (d *BeanDefinition) SetRefreshable(tag string) {
	if !d.Type().Implements(refreshableType) {
		panic("must implement gs.Refreshable interface")
	}
	d.refreshable = true
	d.refreshTag = tag
}

// String returns a string representation of the bean.
func (d *BeanDefinition) String() string {
	return fmt.Sprintf("%s name:%q %s:%d", d.Class(), d.name, d.file, d.line)
}

// NewBean creates a new bean definition.
func NewBean(t reflect.Type, v reflect.Value, f gs.Callable, name string) *BeanDefinition {
	return &BeanDefinition{
		BeanMetadata: &BeanMetadata{
			f:      f,
			status: StatusDefault,
		},
		BeanRuntime: &BeanRuntime{
			t:    t,
			v:    v,
			name: name,
		},
	}
}
