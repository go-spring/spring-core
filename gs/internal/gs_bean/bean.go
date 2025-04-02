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

package gs_bean

import (
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util"
)

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

// refreshableType is the [reflect.Type] of the [gs.Refreshable] interface.
var refreshableType = reflect.TypeFor[gs.Refreshable]()

// BeanMetadata holds the metadata information of a bean.
type BeanMetadata struct {
	f             *gs_arg.Callable
	init          gs.BeanInitFunc
	destroy       gs.BeanDestroyFunc
	dependsOn     []gs.BeanSelector
	exports       []reflect.Type
	conditions    []gs.Condition
	status        BeanStatus
	fileLine      string
	refreshTag    string
	configuration *gs.Configuration
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

// Init returns the initialization function of the bean.
func (d *BeanMetadata) Init() gs.BeanInitFunc {
	return d.init
}

// Destroy returns the destruction function of the bean.
func (d *BeanMetadata) Destroy() gs.BeanDestroyFunc {
	return d.destroy
}

// DependsOn returns the list of dependencies for the bean.
func (d *BeanMetadata) DependsOn() []gs.BeanSelector {
	return d.dependsOn
}

// SetDependsOn sets the list of dependencies for the bean.
func (d *BeanMetadata) SetDependsOn(selectors ...gs.BeanSelector) {
	d.dependsOn = append(d.dependsOn, selectors...)
}

// Exports returns the list of exported types for the bean.
func (d *BeanMetadata) Exports() []reflect.Type {
	return d.exports
}

// Conditions returns the list of conditions for the bean.
func (d *BeanMetadata) Conditions() []gs.Condition {
	return d.conditions
}

// SetCondition adds a condition to the list of conditions for the bean.
func (d *BeanMetadata) SetCondition(c ...gs.Condition) {
	d.conditions = append(d.conditions, c...)
}

// Configuration returns the configuration parameters for the bean.
func (d *BeanMetadata) Configuration() *gs.Configuration {
	return d.configuration
}

// SetConfiguration sets the configuration flag and parameters for the bean.
func (d *BeanDefinition) SetConfiguration(c ...gs.Configuration) {
	var cfg gs.Configuration
	if len(c) > 0 {
		cfg = c[0]
	}
	d.configuration = &gs.Configuration{
		Includes: cfg.Includes,
		Excludes: cfg.Excludes,
	}
}

// RefreshTag returns the refresh tag of the bean.
func (d *BeanMetadata) RefreshTag() string {
	return d.refreshTag
}

// SetCaller sets the caller for the bean.
func (d *BeanMetadata) SetCaller(skip int) {
	_, file, line, _ := runtime.Caller(skip)
	d.SetFileLine(file, line)
}

// FileLine returns the file and line number for the bean.
func (d *BeanMetadata) FileLine() string {
	return d.fileLine
}

// SetFileLine sets the file and line number for the bean.
func (d *BeanMetadata) SetFileLine(file string, line int) {
	d.fileLine = fmt.Sprintf("%s:%d", file, line)
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

// Callable returns the callable for the bean.
func (d *BeanRuntime) Callable() *gs_arg.Callable {
	return nil
}

// Status returns the current status of the bean.
func (d *BeanRuntime) Status() BeanStatus {
	return StatusWired
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

// NewBean creates a new bean definition.
func NewBean(t reflect.Type, v reflect.Value, f *gs_arg.Callable, name string) *BeanDefinition {
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

// SetMock sets the mock object for the bean, replacing its runtime information.
func (d *BeanDefinition) SetMock(obj interface{}) {
	v := reflect.ValueOf(obj)
	t := v.Type()
	for _, r := range d.exports {
		if !t.Implements(r) {
			panic(fmt.Sprintf("mock object %s does not implement %s", t, r))
		}
	}
	*d = BeanDefinition{
		BeanMetadata: &BeanMetadata{
			exports: d.exports,
		},
		BeanRuntime: &BeanRuntime{
			t:    t,
			v:    v,
			name: d.name,
		},
	}
}

// Callable returns the callable for the bean.
func (d *BeanDefinition) Callable() *gs_arg.Callable {
	return d.f
}

// SetName sets the name of the bean.
func (d *BeanDefinition) SetName(name string) {
	d.name = name
}

// Status returns the current status of the bean.
func (d *BeanDefinition) Status() BeanStatus {
	return d.status
}

// SetStatus sets the current status of the bean.
func (d *BeanDefinition) SetStatus(status BeanStatus) {
	d.status = status
}

// SetInit sets the initialization function for the bean.
func (d *BeanDefinition) SetInit(fn gs.BeanInitFunc) {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Type()) {
		d.init = fn
		return
	}
	panic("init should be func(bean) or func(bean)error")
}

// SetDestroy sets the destruction function for the bean.
func (d *BeanDefinition) SetDestroy(fn gs.BeanDestroyFunc) {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Type()) {
		d.destroy = fn
		return
	}
	panic("destroy should be func(bean) or func(bean)error")
}

// SetInitMethod sets the initialization function for the bean by method name.
func (d *BeanDefinition) SetInitMethod(method string) {
	m, ok := d.t.MethodByName(method)
	if !ok {
		panic(fmt.Sprintf("method %s not found on type %s", method, d.t))
	}
	d.SetInit(m.Func.Interface())
}

// SetDestroyMethod sets the destruction function for the bean by method name.
func (d *BeanDefinition) SetDestroyMethod(method string) {
	m, ok := d.t.MethodByName(method)
	if !ok {
		panic(fmt.Sprintf("method %s not found on type %s", method, d.t))
	}
	d.SetDestroy(m.Func.Interface())
}

// SetExport sets the exported interfaces for the bean.
func (d *BeanDefinition) SetExport(exports ...reflect.Type) {
	for _, t := range exports {
		if t.Kind() != reflect.Interface {
			panic("only interface type can be exported")
		}
		if !d.Type().Implements(t) {
			panic(fmt.Sprintf("doesn't implement interface %s", t))
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

// SetRefreshable sets the refreshable flag and tag for the bean.
func (d *BeanDefinition) SetRefreshable(tag string) {
	if !d.Type().Implements(refreshableType) {
		panic("must implement gs.Refreshable interface")
	}
	d.refreshTag = tag
}

// OnProfiles sets the conditions for the bean based on the active profiles.
func (d *BeanDefinition) OnProfiles(profiles string) {
	d.SetCondition(gs_cond.OnFunc(func(ctx gs.CondContext) (bool, error) {
		val := strings.TrimSpace(ctx.Prop("spring.profiles.active"))
		if val == "" {
			return false, nil
		}
		ss := strings.Split(strings.TrimSpace(profiles), ",")
		for s := range slices.Values(strings.Split(val, ",")) {
			for _, x := range ss {
				if s == x {
					return true, nil
				}
			}
		}
		return false, nil
	}))
}

// TypeAndName returns the type and name of the bean.
func (d *BeanDefinition) TypeAndName() (reflect.Type, string) {
	return d.Type(), d.Name()
}

// String returns a string representation of the bean.
func (d *BeanDefinition) String() string {
	return fmt.Sprintf("name=%s %s", d.name, d.fileLine)
}
