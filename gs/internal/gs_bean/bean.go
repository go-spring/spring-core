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

// Package gs_bean provides core bean management for Go-Spring framework.
package gs_bean

import (
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"

	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
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

// String returns a human-readable string for the bean status.
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

// BeanMetadata holds static (design-time) metadata about a bean,
// such as lifecycle functions, dependencies, conditions, and configuration.
type BeanMetadata struct {
	f             *gs_arg.Callable   // Callable for constructor functions
	init          gs.BeanInitFunc    // Bean initialization function
	destroy       gs.BeanDestroyFunc // Bean destruction function
	dependsOn     []gs.BeanSelector  // Explicit dependencies of the bean
	exports       []reflect.Type     // Interfaces exported by this bean
	conditions    []gs.Condition     // Conditions controlling bean creation
	status        BeanStatus         // Current lifecycle status
	mocked        bool               // Indicates if the bean is mocked
	fileLine      string             // File and line where bean is defined
	configuration *gs.Configuration  // Configuration for sub/child beans
}

// Mocked returns true if the bean is mocked.
func (d *BeanMetadata) Mocked() bool {
	return d.mocked
}

// validLifeCycleFunc checks if the given function is a valid lifecycle function.
// Valid lifecycle functions must have the signature:
//
//	func(bean) or func(bean) error
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

// Init returns the bean's initialization function.
func (d *BeanMetadata) Init() gs.BeanInitFunc {
	return d.init
}

// Destroy returns the bean's destruction function.
func (d *BeanMetadata) Destroy() gs.BeanDestroyFunc {
	return d.destroy
}

// DependsOn returns the list of dependencies for the bean.
func (d *BeanMetadata) DependsOn() []gs.BeanSelector {
	return d.dependsOn
}

// SetDependsOn adds dependencies to the bean.
func (d *BeanMetadata) SetDependsOn(selectors ...gs.BeanSelector) {
	d.dependsOn = append(d.dependsOn, selectors...)
}

// Exports returns the interfaces exported by the bean.
func (d *BeanMetadata) Exports() []reflect.Type {
	return d.exports
}

// Conditions returns the list of conditions for the bean.
func (d *BeanMetadata) Conditions() []gs.Condition {
	return d.conditions
}

// SetCondition appends conditions for the bean.
func (d *BeanMetadata) SetCondition(conditions ...gs.Condition) {
	d.conditions = append(d.conditions, conditions...)
}

// Configuration returns the configuration for the bean.
func (d *BeanMetadata) Configuration() *gs.Configuration {
	return d.configuration
}

// SetConfiguration sets configuration (include/exclude) for the bean.
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

// SetCaller records the source file and line number of the bean.
func (d *BeanMetadata) SetCaller(skip int) {
	_, file, line, _ := runtime.Caller(skip)
	d.SetFileLine(file, line)
}

// FileLine returns the source file and line number of the bean.
func (d *BeanMetadata) FileLine() string {
	return d.fileLine
}

// SetFileLine sets the source file and line number of the bean.
func (d *BeanMetadata) SetFileLine(file string, line int) {
	d.fileLine = fmt.Sprintf("%s:%d", file, line)
}

// BeanRuntime holds runtime information about the bean.
type BeanRuntime struct {
	v    reflect.Value // The value of the bean.
	t    reflect.Type  // The type of the bean.
	name string        // The name of the bean.
}

// Name returns the bean's name.
func (d *BeanRuntime) Name() string {
	return d.name
}

// Type returns the bean's type.
func (d *BeanRuntime) Type() reflect.Type {
	return d.t
}

// Value returns the bean as reflect.Value.
func (d *BeanRuntime) Value() reflect.Value {
	return d.v
}

// Interface returns the underlying bean.
func (d *BeanRuntime) Interface() any {
	return d.v.Interface()
}

// Callable returns the bean's callable constructor.
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

// makeBean creates a new BeanDefinition.
func makeBean(t reflect.Type, v reflect.Value, f *gs_arg.Callable, name string) *BeanDefinition {
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

// SetMock replaces the bean's runtime instance with a mock object.
func (d *BeanDefinition) SetMock(obj any) {
	*d = BeanDefinition{
		BeanMetadata: &BeanMetadata{
			exports: d.exports,
			mocked:  true,
		},
		BeanRuntime: &BeanRuntime{
			t:    reflect.TypeOf(obj),
			v:    reflect.ValueOf(obj),
			name: d.name,
		},
	}
}

// Callable returns the bean's callable constructor.
func (d *BeanDefinition) Callable() *gs_arg.Callable {
	return d.f
}

// SetName sets the bean's name.
func (d *BeanDefinition) SetName(name string) {
	d.name = name
}

// Status returns the bean's current lifecycle status.
func (d *BeanDefinition) Status() BeanStatus {
	return d.status
}

// SetStatus sets the bean's current lifecycle status.
func (d *BeanDefinition) SetStatus(status BeanStatus) {
	d.status = status
}

// SetInit sets the bean's initialization function.
func (d *BeanDefinition) SetInit(fn gs.BeanInitFunc) {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Type()) {
		d.init = fn
		return
	}
	panic("init should be func(bean) or func(bean)error")
}

// SetDestroy sets the bean's destruction function.
func (d *BeanDefinition) SetDestroy(fn gs.BeanDestroyFunc) {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Type()) {
		d.destroy = fn
		return
	}
	panic("destroy should be func(bean) or func(bean)error")
}

// SetInitMethod sets the bean's initialization method by name.
func (d *BeanDefinition) SetInitMethod(method string) {
	m, ok := d.t.MethodByName(method)
	if !ok {
		panic(fmt.Sprintf("method %s not found on type %s", method, d.t))
	}
	d.SetInit(m.Func.Interface())
}

// SetDestroyMethod sets the bean's destruction method by name.
func (d *BeanDefinition) SetDestroyMethod(method string) {
	m, ok := d.t.MethodByName(method)
	if !ok {
		panic(fmt.Sprintf("method %s not found on type %s", method, d.t))
	}
	d.SetDestroy(m.Func.Interface())
}

// SetExport registers interfaces exported by the bean.
func (d *BeanDefinition) SetExport(exports ...reflect.Type) {
	for _, t := range exports {
		if t.Kind() != reflect.Interface {
			panic("only interface type can be exported")
		}
		if !d.Type().Implements(t) {
			panic(fmt.Sprintf("doesn't implement interface %s", t))
		}
		if slices.Contains(d.exports, t) {
			continue
		}
		d.exports = append(d.exports, t)
	}
}

// OnProfiles adds conditions based on active profiles.
func (d *BeanDefinition) OnProfiles(profiles string) {
	d.SetCondition(gs_cond.OnFunc(func(ctx gs.ConditionContext) (bool, error) {
		val := strings.TrimSpace(ctx.Prop("spring.profiles.active"))
		if val == "" {
			return false, nil
		}
		ss := strings.Split(strings.TrimSpace(profiles), ",")
		for s := range strings.SplitSeq(val, ",") {
			if slices.Contains(ss, s) {
				return true, nil
			}
		}
		return false, nil
	}))
}

// TypeAndName returns the bean's type and name.
func (d *BeanDefinition) TypeAndName() (reflect.Type, string) {
	return d.Type(), d.Name()
}

// String returns a human-readable description of the bean.
func (d *BeanDefinition) String() string {
	return fmt.Sprintf("name=%s %s", d.name, d.fileLine)
}

// NewBean creates a new BeanDefinition.
// If objOrCtor is a constructor function, it binds its arguments and infers bean name.
// Otherwise, it wraps an existing instance as a bean.
func NewBean(objOrCtor any, ctorArgs ...gs.Arg) *gs.BeanDefinition {

	var f *gs_arg.Callable
	var v reflect.Value
	var fromValue bool
	var name string
	var cond gs.Condition

	switch i := objOrCtor.(type) {
	case reflect.Value:
		fromValue = true
		v = i
	default:
		v = reflect.ValueOf(i)
	}

	t := v.Type()
	if !util.IsBeanType(t) {
		panic("bean must be ref type")
	}

	// Ensure the bean instance is valid and not nil
	if !v.IsValid() || v.IsNil() {
		panic("bean can't be nil")
	}

	// Handle constructor functions
	if !fromValue && t.Kind() == reflect.Func {

		if !util.IsConstructor(t) {
			t1 := "func(...)bean"
			t2 := "func(...)(bean, error)"
			panic(fmt.Sprintf("constructor should be %s or %s", t1, t2))
		}

		// Bind constructor arguments
		var err error
		f, err = gs_arg.NewCallable(objOrCtor, ctorArgs)
		if err != nil {
			panic(err)
		}

		var in0 reflect.Type
		if t.NumIn() > 0 {
			in0 = t.In(0)
		}

		// Prepare the return type
		out0 := t.Out(0)
		v = reflect.New(out0)
		if util.IsBeanType(out0) {
			v = v.Elem()
		}

		t = v.Type()
		if !util.IsBeanType(t) {
			panic("bean must be ref type")
		}

		// Derive bean name from constructor function name
		fnPtr := reflect.ValueOf(objOrCtor).Pointer()
		fnInfo := runtime.FuncForPC(fnPtr)
		funcName := fnInfo.Name()
		name = funcName[strings.LastIndex(funcName, "/")+1:]
		name = name[strings.Index(name, ".")+1:]
		if name[0] == '(' {
			name = name[strings.Index(name, ".")+1:]
		}

		// If the constructor is a method, set a condition for its owner bean
		method := strings.LastIndexByte(fnInfo.Name(), ')') > 0
		if method {
			var s gs.BeanSelector = gs.BeanSelectorImpl{Type: in0}
			if len(ctorArgs) > 0 {
				switch a := ctorArgs[0].(type) {
				case *gs.RegisteredBean:
					s = a
				case *gs.BeanDefinition:
					s = a
				case gs_arg.IndexArg:
					if a.Idx == 0 {
						switch x := a.Arg.(type) {
						case *gs.RegisteredBean:
							s = x
						case *gs.BeanDefinition:
							s = x
						default:
							panic("the arg of IndexArg[0] should be *RegisteredBean or *BeanDefinition")
						}
					}
				default:
					panic("ctorArgs[0] should be *RegisteredBean or *BeanDefinition or IndexArg[0]")
				}
			}
			cond = gs_cond.OnBeanSelector(s)
		}
	}

	// Fallback: derive name from the type
	if name == "" {
		s := strings.Split(t.String(), ".")
		name = strings.TrimPrefix(s[len(s)-1], "*")
	}

	d := makeBean(t, v, f, name)
	if cond != nil {
		d.SetCondition(cond)
	}
	return gs.NewBeanDefinition(d)
}
