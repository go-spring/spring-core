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
	"errors"
	"fmt"
	"reflect"
	"runtime"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util"
)

// refreshableType is the [reflect.Type] of the interface [gs.Refreshable].
var refreshableType = reflect.TypeFor[gs.Refreshable]()

// BeanStatus is the status of a bean.
type BeanStatus int8

const (
	StatusDeleted = BeanStatus(-1)
	StatusDefault = BeanStatus(iota)
	StatusResolving
	StatusResolved
	StatusCreating
	StatusCreated
	StatusWired
)

// GetStatusString returns the string of the given status.
func GetStatusString(status BeanStatus) string {
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
	f       gs.Callable
	init    interface{}
	destroy interface{}
	depend  []gs.BeanSelector
	export  []reflect.Type
	cond    []gs.Condition
	status  BeanStatus

	file string
	line int

	isConfiguration    bool
	configurationParam gs.ConfigurationParam

	refreshable bool
	refreshTag  string
}

// Init returns the bean initialization function.
func (d *BeanMetadata) Init() interface{} {
	return d.init
}

// Destroy returns the bean destruction function.
func (d *BeanMetadata) Destroy() interface{} {
	return d.destroy
}

// DependsOn returns the bean's dependencies.
func (d *BeanMetadata) DependsOn() []gs.BeanSelector {
	return d.depend
}

// Export returns the bean's exported types.
func (d *BeanMetadata) Export() []reflect.Type {
	return d.export
}

// Configuration returns whether the bean is a configuration bean.
func (d *BeanMetadata) Configuration() bool {
	return d.isConfiguration
}

func (d *BeanMetadata) ConfigurationParam() gs.ConfigurationParam {
	return d.configurationParam
}

func (d *BeanMetadata) Refreshable() bool {
	return d.refreshable
}

func (d *BeanMetadata) RefreshTag() string {
	return d.refreshTag
}

// Condition returns the combined conditions for the bean.
func (d *BeanMetadata) Condition() []gs.Condition {
	return d.cond
}

// File returns the bean's file.
func (d *BeanMetadata) File() string {
	return d.file
}

// Line returns the bean's line.
func (d *BeanMetadata) Line() int {
	return d.line
}

// Class returns the bean's class.
func (d *BeanMetadata) Class() string {
	if d.f == nil {
		return "object bean"
	}
	return "constructor bean"
}

// BeanRuntime represents the runtime information of a bean.
type BeanRuntime struct {
	v        reflect.Value
	t        reflect.Type
	name     string
	typeName string
}

// ID returns the bean's id.
func (d *BeanRuntime) ID() string {
	return d.typeName + ":" + d.name
}

// Name returns the bean's name.
func (d *BeanRuntime) Name() string {
	return d.name
}

// TypeName returns the bean's original type name.
func (d *BeanRuntime) TypeName() string {
	return d.typeName
}

// Type returns the bean's type.
func (d *BeanRuntime) Type() reflect.Type {
	return d.t
}

// Value returns the bean's value as reflect.Value.
func (d *BeanRuntime) Value() reflect.Value {
	return d.v
}

// Interface returns the bean's underlying value.
func (d *BeanRuntime) Interface() interface{} {
	return d.v.Interface()
}

// Status returns the bean's status.
func (d *BeanRuntime) Status() BeanStatus {
	return StatusWired
}

// Callable returns the bean's callable.
func (d *BeanRuntime) Callable() gs.Callable {
	return nil
}

// Match returns whether the bean matches the given typeName and beanName.
func (d *BeanRuntime) Match(typeName string, beanName string) bool {

	typeIsSame := false
	if typeName == "" || d.typeName == typeName {
		typeIsSame = true
	}

	nameIsSame := false
	if beanName == "" || d.name == beanName {
		nameIsSame = true
	}

	return typeIsSame && nameIsSame
}

// String returns the bean's string.
func (d *BeanRuntime) String() string {
	return d.name
}

// BeanDefinition bean 元数据。
type BeanDefinition struct {
	*BeanMetadata
	*BeanRuntime
}

// Callable returns the bean's callable.
func (d *BeanDefinition) Callable() gs.Callable {
	return d.f
}

// Status returns the bean's status.
func (d *BeanDefinition) Status() BeanStatus {
	return d.status
}

// SetStatus sets the bean's status.
func (d *BeanMetadata) SetStatus(status BeanStatus) {
	d.status = status
}

// SetCaller sets the bean's caller.
func (d *BeanDefinition) SetCaller(skip int) {
	_, d.file, d.line, _ = runtime.Caller(skip)
}

// SetName sets the bean's name.
func (d *BeanDefinition) SetName(name string) {
	d.name = name
}

func validLifeCycleFunc(fnType reflect.Type, beanValue reflect.Value) bool {
	if !util.IsFuncType(fnType) {
		return false
	}
	if fnType.NumIn() != 1 || !util.HasReceiver(fnType, beanValue) {
		return false
	}
	return util.ReturnNothing(fnType) || util.ReturnOnlyError(fnType)
}

// SetInit sets the bean's initialization function.
func (d *BeanDefinition) SetInit(fn interface{}) {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Value()) {
		d.init = fn
		return
	}
	panic(errors.New("init should be func(bean) or func(bean)error"))
}

// SetDestroy sets the bean's destruction function.
func (d *BeanDefinition) SetDestroy(fn interface{}) {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Value()) {
		d.destroy = fn
		return
	}
	panic(errors.New("destroy should be func(bean) or func(bean)error"))
}

func (d *BeanDefinition) AddCondition(cond gs.Condition) {
	if cond != nil {
		d.cond = append(d.cond, cond)
	}
}

// SetDependsOn sets the bean's dependency.
func (d *BeanDefinition) SetDependsOn(selectors ...gs.BeanSelector) {
	d.depend = append(d.depend, selectors...)
}

// SetExport sets the bean's exported interfaces.
func (d *BeanDefinition) SetExport(exports ...interface{}) {
	for _, o := range exports {
		t, ok := o.(reflect.Type)
		if !ok {
			t = reflect.TypeOf(o)
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
		}
		if t.Kind() != reflect.Interface {
			panic(errors.New("only interface type can be exported"))
		}
		exported := false
		for _, export := range d.export {
			if t == export {
				exported = true
				break
			}
		}
		if exported {
			continue
		}
		d.export = append(d.export, t)
	}
}

func (d *BeanDefinition) SetConfiguration(param ...gs.ConfigurationParam) {
	d.isConfiguration = true
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

func (d *BeanDefinition) SetRefreshable(tag string) {
	if !d.Type().Implements(refreshableType) {
		panic(errors.New("must implement dync.Refreshable interface"))
	}
	d.refreshable = true
	d.refreshTag = tag
}

func (d *BeanDefinition) String() string {
	return fmt.Sprintf("%s name:%q %s:%d", d.Class(), d.name, d.file, d.line)
}

// NewBean 普通函数注册时需要使用 reflect.ValueOf(fn) 形式以避免和构造函数发生冲突。
func NewBean(t reflect.Type, v reflect.Value, f gs.Callable, name string) *BeanDefinition {
	return &BeanDefinition{
		BeanMetadata: &BeanMetadata{
			f:      f,
			status: StatusDefault,
		},
		BeanRuntime: &BeanRuntime{
			t:        t,
			v:        v,
			name:     name,
			typeName: util.TypeName(t),
		},
	}
}
