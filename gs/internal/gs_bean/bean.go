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
	"github.com/go-spring/spring-core/util"
)

var refreshableType = reflect.TypeFor[gs.Refreshable]()

type BeanStatus int8

const (
	Deleted   = BeanStatus(-1)   // 已删除
	Default   = BeanStatus(iota) // 未处理
	Resolving                    // 正在决议
	Resolved                     // 已决议
	Creating                     // 正在创建
	Created                      // 已创建
	Wired                        // 注入完成
)

// GetStatusString 获取 bean 状态的字符串表示。
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
		return ""
	}
}

type BeanInit interface {
	OnInit(ctx gs.Context) error
}

type BeanDestroy interface {
	OnDestroy()
}

type BeanMetadata struct {
	f       gs.Callable       // 构造函数
	cond    []gs.Condition    // 判断条件
	init    interface{}       // 初始化函数
	destroy interface{}       // 销毁函数
	depends []gs.BeanSelector // 间接依赖项
	exports []reflect.Type    // 导出的接口
	file    string            // 注册点所在文件
	line    int               // 注册点所在行数
	status  BeanStatus        // 状态

	configuration gs.ConfigurationParam

	enableRefresh bool
	refreshParam  conf.BindParam
}

func (d *BeanMetadata) SetStatus(status BeanStatus) {
	d.status = status
}

func (d *BeanMetadata) Cond() []gs.Condition {
	return d.cond
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
func NewBean(t reflect.Type, v reflect.Value, f gs.Callable, name string, file string, line int) *BeanDefinition {
	return &BeanDefinition{
		BeanMetadata: &BeanMetadata{
			f:      f,
			file:   file,
			line:   line,
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
