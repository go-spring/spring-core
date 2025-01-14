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

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util"
)

// refreshableType is the [reflect.Type] of [gs.Refreshable].
var refreshableType = reflect.TypeFor[gs.Refreshable]()

// BeanStatus is the status of a bean.
type BeanStatus int8

const (
	Deleted = BeanStatus(-1)
	Default = BeanStatus(iota)
	Resolving
	Resolved
	Creating
	Created
	Wired
)

// GetStatusString returns the string of the given status.
func GetStatusString(status BeanStatus) string {
	switch status {
	case Deleted:
		return "Deleted"
	case Default:
		return "Default"
	case Resolving:
		return "Resolving"
	case Resolved:
		return "Resolved"
	case Creating:
		return "Creating"
	case Created:
		return "Created"
	case Wired:
		return "Wired"
	default:
		panic("unknown bean status")
	}
}

// BeanInit is used for init of a bean.
type BeanInit interface {
	OnInit(ctx gs.Context) error
}

// BeanDestroy is used for destroy of a bean.
type BeanDestroy interface {
	OnDestroy()
}

// BeanMetadata stores the metadata of a bean.
type BeanMetadata struct {
	f       gs.Callable
	cond    []gs.Condition
	init    interface{}
	destroy interface{}
	depends []gs.BeanSelector
	exports []reflect.Type
	file    string
	line    int
	status  BeanStatus

	configuration gs.ConfigurationParam

	enableRefresh bool
	refreshParam  conf.BindParam
}

// Condition returns the condition of a bean.
func (d *BeanMetadata) Condition() gs.Condition {
	if n := len(d.cond); n == 0 {
		return nil
	} else if n == 1 {
		return d.cond[0]
	} else {
		return gs_cond.And(d.cond...)
	}
}

func (d *BeanMetadata) Init() interface{} {
	return d.init
}

func (d *BeanMetadata) Destroy() interface{} {
	return d.destroy
}

func (d *BeanMetadata) Depends() []gs.BeanSelector {
	return d.depends
}

func (d *BeanMetadata) Exports() []reflect.Type {
	return d.exports
}

func (d *BeanMetadata) IsConfiguration() bool {
	return d.configuration.Enable
}

func (d *BeanMetadata) GetIncludeMethod() []string {
	return d.configuration.Include
}

func (d *BeanMetadata) GetExcludeMethod() []string {
	return d.configuration.Exclude
}

func (d *BeanMetadata) EnableRefresh() bool {
	return d.enableRefresh
}

func (d *BeanMetadata) RefreshParam() conf.BindParam {
	return d.refreshParam
}

func (d *BeanMetadata) File() string {
	return d.file
}

func (d *BeanMetadata) Line() int {
	return d.line
}

// FileLine 返回 bean 的注册点。
func (d *BeanMetadata) FileLine() string {
	return fmt.Sprintf("%s:%d", d.file, d.line)
}

// Class 返回 bean 的类型描述。
func (d *BeanMetadata) Class() string {
	if d.f == nil {
		return "object bean"
	}
	return "constructor bean"
}

type BeanRuntime struct {
	v        reflect.Value // 值
	t        reflect.Type  // 类型
	name     string        // 名称
	typeName string        // 原始类型的全限定名
	primary  bool          // 是否为主版本
}

// ID 返回 bean 的 ID 。
func (d *BeanRuntime) ID() string {
	return d.typeName + ":" + d.name
}

// Name 返回 bean 的名称。
func (d *BeanRuntime) Name() string {
	return d.name
}

// TypeName 返回 bean 的原始类型的全限定名。
func (d *BeanRuntime) TypeName() string {
	return d.typeName
}

func (d *BeanRuntime) Callable() gs.Callable {
	return nil
}

// Interface 返回 bean 的真实值。
func (d *BeanRuntime) Interface() interface{} {
	return d.v.Interface()
}

func (d *BeanRuntime) IsPrimary() bool {
	return d.primary
}

func (d *BeanRuntime) Type() reflect.Type {
	return d.t
}

func (d *BeanRuntime) Value() reflect.Value {
	return d.v
}

func (d *BeanRuntime) Status() BeanStatus {
	return Wired
}

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

func (d *BeanRuntime) String() string {
	return d.name
}

// BeanDefinition bean 元数据。
type BeanDefinition struct {
	*BeanMetadata
	*BeanRuntime
}

func (d *BeanDefinition) Callable() gs.Callable {
	return d.f
}

func (d *BeanDefinition) Status() BeanStatus {
	return d.status
}

func (d *BeanMetadata) SetStatus(status BeanStatus) {
	d.status = status
}

func (d *BeanDefinition) String() string {
	return fmt.Sprintf("%s name:%q %s", d.Class(), d.name, d.FileLine())
}

// SetName 设置 bean 的名称。
func (d *BeanDefinition) SetName(name string) {
	d.name = name
}

func (d *BeanDefinition) SetCaller(skip int) {
	_, d.file, d.line, _ = runtime.Caller(skip)
}

// SetCondition 设置 bean 的 Condition。
func (d *BeanDefinition) SetCondition(cond gs.Condition) {
	if cond != nil {
		d.cond = append(d.cond, cond)
	}
}

// SetDependsOn 设置 bean 的间接依赖项。
func (d *BeanDefinition) SetDependsOn(selectors ...gs.BeanSelector) {
	d.depends = append(d.depends, selectors...)
}

// SetPrimary 设置 bean 为主版本。
func (d *BeanDefinition) SetPrimary() {
	d.primary = true
}

// validLifeCycleFunc 判断是否是合法的用于 bean 生命周期控制的函数，生命周期函数
// 的要求：只能有一个入参并且必须是 bean 的类型，没有返回值或者只返回 error 类型值。
func validLifeCycleFunc(fnType reflect.Type, beanValue reflect.Value) bool {
	if !util.IsFuncType(fnType) {
		return false
	}
	if fnType.NumIn() != 1 || !util.HasReceiver(fnType, beanValue) {
		return false
	}
	return util.ReturnNothing(fnType) || util.ReturnOnlyError(fnType)
}

// SetInit 设置 bean 的初始化函数。
func (d *BeanDefinition) SetInit(fn interface{}) {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Value()) {
		d.init = fn
		return
	}
	panic(errors.New("init should be func(bean) or func(bean)error"))
}

// SetDestroy 设置 bean 的销毁函数。
func (d *BeanDefinition) SetDestroy(fn interface{}) {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Value()) {
		d.destroy = fn
		return
	}
	panic(errors.New("destroy should be func(bean) or func(bean)error"))
}

// SetExport 设置 bean 的导出接口。
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

func (d *BeanDefinition) SetConfiguration(param ...gs.ConfigurationParam) {
	if len(param) > 0 {
		d.configuration = param[0]
	}
	d.configuration.Enable = true
}

func (d *BeanDefinition) SetEnableRefresh(tag string) {
	if !d.Type().Implements(refreshableType) {
		panic(errors.New("must implement dync.Refreshable interface"))
	}
	d.enableRefresh = true
	err := d.refreshParam.BindTag(tag, "")
	if err != nil {
		panic(err)
	}
}

// NewBean 普通函数注册时需要使用 reflect.ValueOf(fn) 形式以避免和构造函数发生冲突。
func NewBean(t reflect.Type, v reflect.Value, f gs.Callable, name string) *BeanDefinition {
	return &BeanDefinition{
		BeanMetadata: &BeanMetadata{
			f:      f,
			status: Default,
		},
		BeanRuntime: &BeanRuntime{
			t:        t,
			v:        v,
			name:     name,
			typeName: util.TypeName(t),
		},
	}
}
