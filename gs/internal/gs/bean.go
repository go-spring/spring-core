package gs

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/util"
)

var refreshableType = reflect.TypeFor[Refreshable]()

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
	OnInit(ctx Context) error
}

type BeanDestroy interface {
	OnDestroy()
}

// beanRegistration bean 注册数据。
type beanRegistration struct {
	f       Callable       // 构造函数
	method  bool           // 是否为成员方法
	cond    Condition      // 判断条件
	init    interface{}    // 初始化函数
	destroy interface{}    // 销毁函数
	depends []BeanSelector // 间接依赖项
	exports []reflect.Type // 导出的接口
	file    string         // 注册点所在文件
	line    int            // 注册点所在行数
	status  BeanStatus     // 状态

	configuration bool     // 是否扫描成员方法
	includeMethod []string // 包含哪些成员方法
	excludeMethod []string // 排除那些成员方法

	enableRefresh bool
	refreshParam  conf.BindParam
}

type BeanRuntimeMeta struct {
	v        reflect.Value // 值
	t        reflect.Type  // 类型
	name     string        // 名称
	typeName string        // 原始类型的全限定名
	primary  bool          // 是否为主版本
}

func (d *BeanRuntimeMeta) Callable() Callable {
	return nil
}

// Interface 返回 bean 的真实值。
func (d *BeanRuntimeMeta) Interface() interface{} {
	return d.v.Interface()
}

func (d *BeanRuntimeMeta) GetName() string {
	return d.name
}

func (d *BeanRuntimeMeta) IsPrimary() bool {
	return d.primary
}

func (d *BeanRuntimeMeta) Type() reflect.Type {
	return d.t
}

func (d *BeanRuntimeMeta) Value() reflect.Value {
	return d.v
}

func (d *BeanRuntimeMeta) GetStatus() BeanStatus {
	return Wired
}

func (d *BeanRuntimeMeta) Match(typeName string, beanName string) bool {

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

func (d *BeanRuntimeMeta) String() string {
	return d.name
}

// BeanDefinition bean 元数据。
type BeanDefinition struct {
	*beanRegistration
	*BeanRuntimeMeta
}

func (d *BeanDefinition) Callable() Callable {
	return d.f
}

func (d *BeanDefinition) GetTypeName() string {
	return d.typeName
}

func (d *BeanDefinition) GetStatus() BeanStatus {
	return d.status
}

func (d *BeanDefinition) SetStatus(status BeanStatus) {
	d.status = status
}

func (d *BeanDefinition) IsMethod() bool {
	return d.method
}

func (d *BeanDefinition) GetCond() Condition {
	return d.cond
}

func (d *BeanDefinition) GetInit() interface{} {
	return d.init
}

func (d *BeanDefinition) GetDestroy() interface{} {
	return d.destroy
}

func (d *BeanDefinition) GetDepends() []BeanSelector {
	return d.depends
}

func (d *BeanDefinition) GetExports() []reflect.Type {
	return d.exports
}

func (d *BeanDefinition) IsConfiguration() bool {
	return d.configuration
}

func (d *BeanDefinition) GetIncludeMethod() []string {
	return d.includeMethod
}

func (d *BeanDefinition) GetExcludeMethod() []string {
	return d.excludeMethod
}

func (d *BeanDefinition) IsRefreshEnable() bool {
	return d.enableRefresh
}

func (d *BeanDefinition) GetRefreshParam() conf.BindParam {
	return d.refreshParam
}

// ID 返回 bean 的 ID 。
func (d *BeanDefinition) ID() string {
	return d.typeName + ":" + d.name
}

// BeanName 返回 bean 的名称。
func (d *BeanDefinition) BeanName() string {
	return d.name
}

// TypeName 返回 bean 的原始类型的全限定名。
func (d *BeanDefinition) TypeName() string {
	return d.typeName
}

// Created 返回是否已创建。
func (d *BeanDefinition) Created() bool {
	return d.status >= Created
}

// Wired 返回 bean 是否已经注入。
func (d *BeanDefinition) Wired() bool {
	return d.status == Wired
}

func (d *BeanDefinition) File() string {
	return d.file
}

func (d *BeanDefinition) Line() int {
	return d.line
}

// FileLine 返回 bean 的注册点。
func (d *BeanDefinition) FileLine() string {
	return fmt.Sprintf("%s:%d", d.file, d.line)
}

// GetClass 返回 bean 的类型描述。
func (d *BeanDefinition) GetClass() string {
	if d.f == nil {
		return "object bean"
	}
	return "constructor bean"
}

func (d *BeanDefinition) String() string {
	return fmt.Sprintf("%s name:%q %s", d.GetClass(), d.name, d.FileLine())
}

// Name 设置 bean 的名称。
func (d *BeanDefinition) Name(name string) *BeanDefinition {
	d.name = name
	return d
}

func (d *BeanDefinition) Caller(skip int) *BeanDefinition {
	_, d.file, d.line, _ = runtime.Caller(skip)
	return d
}

// On 设置 bean 的 Condition。
func (d *BeanDefinition) On(cond Condition) *BeanDefinition {
	d.cond = cond
	return d
}

// DependsOn 设置 bean 的间接依赖项。
func (d *BeanDefinition) DependsOn(selectors ...BeanSelector) *BeanDefinition {
	d.depends = append(d.depends, selectors...)
	return d
}

// Primary 设置 bean 为主版本。
func (d *BeanDefinition) Primary() *BeanDefinition {
	d.primary = true
	return d
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

// Init 设置 bean 的初始化函数。
func (d *BeanDefinition) Init(fn interface{}) *BeanDefinition {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Value()) {
		d.init = fn
		return d
	}
	panic(errors.New("init should be func(bean) or func(bean)error"))
}

// Destroy 设置 bean 的销毁函数。
func (d *BeanDefinition) Destroy(fn interface{}) *BeanDefinition {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Value()) {
		d.destroy = fn
		return d
	}
	panic(errors.New("destroy should be func(bean) or func(bean)error"))
}

// Export 设置 bean 的导出接口。
func (d *BeanDefinition) Export(exports ...interface{}) *BeanDefinition {
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
	return d
}

// Configuration 设置 bean 为配置类。
func (d *BeanDefinition) Configuration(includes []string, excludes []string) *BeanDefinition {
	d.configuration = true
	d.includeMethod = includes
	d.excludeMethod = excludes
	return d
}

// EnableRefresh 设置 bean 为可刷新的。
func (d *BeanDefinition) EnableRefresh(tag string) *BeanDefinition {
	if !d.Type().Implements(refreshableType) {
		panic(errors.New("must implement dync.Refreshable interface"))
	}
	d.enableRefresh = true
	err := d.refreshParam.BindTag(tag, "")
	if err != nil {
		panic(err)
	}
	return d
}

// NewBean 普通函数注册时需要使用 reflect.ValueOf(fn) 形式以避免和构造函数发生冲突。
func NewBean(t reflect.Type, v reflect.Value, f Callable, name string, method bool, file string, line int) *BeanDefinition {
	return &BeanDefinition{
		beanRegistration: &beanRegistration{
			f:      f,
			method: method,
			file:   file,
			line:   line,
			status: Default,
		},
		BeanRuntimeMeta: &BeanRuntimeMeta{
			t:        t,
			v:        v,
			name:     name,
			typeName: util.TypeName(t),
		},
	}
}
