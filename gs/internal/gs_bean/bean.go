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

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/stdlib/typeutil"
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

// Configuration specifies parameters for configuring beans during registration.
type Configuration struct {
	Includes []string // Methods to include
	Excludes []string // Methods to exclude
}

// BeanDefinition contains both metadata and runtime information of a bean.
type BeanDefinition struct {
	v             reflect.Value    // The value of the bean.
	t             reflect.Type     // The type of the bean.
	f             *gs_arg.Callable // Callable for constructor functions
	name          string           // The name of the bean.
	init          any              // Bean initialization function
	destroy       any              // Bean destruction function
	dependsOn     []gs.BeanID      // Explicit dependencies of the bean
	exports       []reflect.Type   // Interfaces exported by this bean
	conditions    []gs.Condition   // Conditions controlling bean creation
	status        BeanStatus       // Current lifecycle status
	fileLine      string           // File and line where bean is defined
	configuration *Configuration   // Configuration for sub/child beans
}

// Clone creates a copy of the BeanDefinition.
// For pointer beans, a new instance of the underlying type is created.
// For function beans, the value is shared (functions are immutable in Go).
// This ensures the cloned BeanDefinition has a separate reflect.Value when necessary.
func (d *BeanDefinition) Clone() *BeanDefinition {
	r := *d
	if d.f != nil { // Constructor
		r.v = reflect.New(d.t).Elem()
		return &r
	}
	if d.t.Kind() == reflect.Func { // Function
		return &r
	}
	r.v = reflect.New(d.t.Elem())
	return &r
}

// validLifeCycleFunc checks if the given function is a valid lifecycle function.
// Valid lifecycle functions must have the signature:
//
//	func(bean) or func(bean) error
func validLifeCycleFunc(fnType reflect.Type, beanType reflect.Type) bool {
	if !typeutil.IsFuncType(fnType) || fnType.NumIn() != 1 {
		return false
	}
	if t := fnType.In(0); t.Kind() == reflect.Interface {
		if !beanType.Implements(t) {
			return false
		}
	} else if t != beanType {
		return false
	}
	return typeutil.ReturnNothing(fnType) || typeutil.ReturnOnlyError(fnType)
}

// GetInit returns the bean's initialization function.
func (d *BeanDefinition) GetInit() any {
	return d.init
}

// GetDestroy returns the bean's destruction function.
func (d *BeanDefinition) GetDestroy() any {
	return d.destroy
}

// GetDependsOn returns the list of dependencies for the bean.
func (d *BeanDefinition) GetDependsOn() []gs.BeanID {
	return d.dependsOn
}

// DependsOn adds dependencies to the bean.
func (d *BeanDefinition) DependsOn(selectors ...gs.BeanID) *BeanDefinition {
	d.dependsOn = append(d.dependsOn, selectors...)
	return d
}

// Exports returns the interfaces exported by the bean.
func (d *BeanDefinition) Exports() []reflect.Type {
	return d.exports
}

// Conditions returns the list of conditions for the bean.
func (d *BeanDefinition) Conditions() []gs.Condition {
	return d.conditions
}

// Condition appends conditions for the bean.
func (d *BeanDefinition) Condition(conditions ...gs.Condition) *BeanDefinition {
	d.conditions = append(d.conditions, conditions...)
	return d
}

// GetConfiguration returns the configuration for the bean.
func (d *BeanDefinition) GetConfiguration() *Configuration {
	return d.configuration
}

// Configuration sets configuration (include/exclude) for the bean.
func (d *BeanDefinition) Configuration(c ...Configuration) *BeanDefinition {
	var cfg Configuration
	if len(c) > 0 {
		cfg = c[0]
	}
	d.configuration = &Configuration{
		Includes: cfg.Includes,
		Excludes: cfg.Excludes,
	}
	return d
}

// Caller records the source file and line number of the bean.
func (d *BeanDefinition) Caller(skip int) *BeanDefinition {
	_, file, line, _ := runtime.Caller(skip)
	d.SetFileLine(file, line)
	return d
}

// FileLine returns the source file and line number of the bean.
func (d *BeanDefinition) FileLine() string {
	return d.fileLine
}

// SetFileLine sets the source file and line number of the bean.
func (d *BeanDefinition) SetFileLine(file string, line int) {
	d.fileLine = fmt.Sprintf("%s:%d", file, line)
}

// BeanID returns the bean's identifier.
func (d *BeanDefinition) BeanID() gs.BeanID {
	return gs.BeanID{Name: d.GetName(), Type: d.GetType()}
}

// GetName returns the bean's name.
func (d *BeanDefinition) GetName() string {
	return d.name
}

// GetType returns the bean's type.
func (d *BeanDefinition) GetType() reflect.Type {
	return d.t
}

// GetValue returns the bean as reflect.Value.
func (d *BeanDefinition) GetValue() reflect.Value {
	return d.v
}

// Interface returns the underlying bean.
func (d *BeanDefinition) Interface() any {
	return d.v.Interface()
}

// GetArgValue returns the bean’s value for argument injection.
func (d *BeanDefinition) GetArgValue(ctx gs.ArgContext, t reflect.Type) (reflect.Value, error) {
	return d.GetValue(), nil
}

// makeBean creates a new BeanDefinition.
func makeBean(t reflect.Type, v reflect.Value, f *gs_arg.Callable, name string) *BeanDefinition {
	return &BeanDefinition{
		f:      f,
		t:      t,
		v:      v,
		name:   name,
		status: StatusDefault,
	}
}

// Callable returns the bean's callable constructor.
func (d *BeanDefinition) Callable() *gs_arg.Callable {
	return d.f
}

// Name sets the bean's name.
func (d *BeanDefinition) Name(name string) *BeanDefinition {
	d.name = name
	return d
}

// Status returns the bean's current lifecycle status.
func (d *BeanDefinition) Status() BeanStatus {
	return d.status
}

// SetStatus sets the bean's current lifecycle status.
func (d *BeanDefinition) SetStatus(status BeanStatus) {
	d.status = status
}

// Init sets the bean's initialization function.
func (d *BeanDefinition) Init(fn any) *BeanDefinition {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.GetType()) {
		d.init = fn
		return d
	}
	panic("init should be func(bean) or func(bean)error")
}

// Destroy sets the bean's destruction function.
func (d *BeanDefinition) Destroy(fn any) *BeanDefinition {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.GetType()) {
		d.destroy = fn
		return d
	}
	panic("destroy should be func(bean) or func(bean)error")
}

// InitMethod sets the bean's initialization method by name.
func (d *BeanDefinition) InitMethod(method string) *BeanDefinition {
	m, ok := d.t.MethodByName(method)
	if !ok {
		panic(fmt.Sprintf("method %s not found on type %s", method, d.t))
	}
	return d.Init(m.Func.Interface())
}

// DestroyMethod sets the bean's destruction method by name.
func (d *BeanDefinition) DestroyMethod(method string) *BeanDefinition {
	m, ok := d.t.MethodByName(method)
	if !ok {
		panic(fmt.Sprintf("method %s not found on type %s", method, d.t))
	}
	return d.Destroy(m.Func.Interface())
}

// Export registers interfaces exported by the bean.
func (d *BeanDefinition) Export(exports ...reflect.Type) *BeanDefinition {
	for _, t := range exports {
		if t.Kind() != reflect.Interface {
			panic("only interface type can be exported")
		}
		if !d.GetType().Implements(t) {
			panic(fmt.Sprintf("doesn't implement interface %s", t))
		}
		if slices.Contains(d.exports, t) {
			continue
		}
		d.exports = append(d.exports, t)
	}
	return d
}

// OnProfiles adds a creation condition based on active profiles.
// The bean will only be created if the application's "spring.profiles.active"
// property contains at least one of the specified profiles.
// Multiple profiles can be provided as a comma-separated string.
//
// Example:
//
//	d.OnProfiles("dev,test")  // bean created if active profile is "dev" or "test"
func (d *BeanDefinition) OnProfiles(profiles string) *BeanDefinition {
	d.Condition(gs_cond.OnFunc(func(ctx gs.ConditionContext) (bool, error) {
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
	return d
}

// String returns a human-readable description of the bean.
func (d *BeanDefinition) String() string {
	return fmt.Sprintf("name=%s %s", d.name, d.fileLine)
}

// NewBean creates a new BeanDefinition.
// If objOrCtor is a constructor function, it binds its arguments and infers bean name.
// Otherwise, it wraps an existing instance as a bean.
func NewBean(objOrCtor any, ctorArgs ...gs.Arg) *BeanDefinition {

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
	if !typeutil.IsBeanType(t) {
		panic("bean must be ref type")
	}

	// Ensure the bean instance is valid and not nil
	if !v.IsValid() || v.IsNil() {
		panic("bean can't be nil")
	}

	// Handle constructor functions
	if !fromValue && t.Kind() == reflect.Func {

		if !typeutil.IsConstructor(t) {
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
		if typeutil.IsBeanType(out0) {
			v = v.Elem()
		}

		t = v.Type()
		if !typeutil.IsBeanType(t) {
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
			var s = gs.BeanID{Type: in0}
			if len(ctorArgs) > 0 {
				switch a := ctorArgs[0].(type) {
				case *BeanDefinition:
					s = gs.BeanID{Type: a.t, Name: a.name}
				case gs_arg.IndexArg:
					if a.Idx == 0 {
						switch x := a.Arg.(type) {
						case *BeanDefinition:
							s = gs.BeanID{Type: x.t, Name: x.name}
						default:
							panic("the arg of IndexArg[0] should be *BeanDefinition")
						}
					}
				default:
					panic("ctorArgs[0] should be *BeanDefinition or IndexArg[0]")
				}
			}
			cond = gs_cond.OnBeanID(s)
		}
	}

	// Fallback: derive name from the type
	if name == "" {
		s := strings.Split(t.String(), ".")
		name = strings.TrimPrefix(s[len(s)-1], "*")
	}

	d := makeBean(t, v, f, name)
	if cond != nil {
		d.Condition(cond)
	}
	return d
}
