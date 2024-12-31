package gs

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-spring/spring-core/gs/internal/gs_util"
)

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

// BeanDefinition bean 元数据。
type BeanDefinition struct {
	V reflect.Value // 值
	T reflect.Type  // 类型
	F Callable      // 构造函数

	file string // 注册点所在文件
	line int    // 注册点所在行数

	name     string         // 名称
	typeName string         // 原始类型的全限定名
	status   BeanStatus     // 状态
	primary  bool           // 是否为主版本
	method   bool           // 是否为成员方法
	cond     Condition      // 判断条件
	order    float32        // 收集时的顺序
	init     interface{}    // 初始化函数
	destroy  interface{}    // 销毁函数
	depends  []BeanSelector // 间接依赖项
	exports  []reflect.Type // 导出的接口

	configuration bool     // 是否扫描成员方法
	includeMethod []string // 包含哪些成员方法
	excludeMethod []string // 排除那些成员方法
}

func (d *BeanDefinition) GetName() string             { return d.name }
func (d *BeanDefinition) GetTypeName() string         { return d.typeName }
func (d *BeanDefinition) GetStatus() BeanStatus       { return d.status }
func (d *BeanDefinition) SetStatus(status BeanStatus) { d.status = status }
func (d *BeanDefinition) IsPrimary() bool             { return d.primary }
func (d *BeanDefinition) IsMethod() bool              { return d.method }
func (d *BeanDefinition) GetCond() Condition          { return d.cond }
func (d *BeanDefinition) GetOrder() float32           { return d.order }
func (d *BeanDefinition) GetInit() interface{}        { return d.init }
func (d *BeanDefinition) GetDestroy() interface{}     { return d.destroy }
func (d *BeanDefinition) GetDepends() []BeanSelector  { return d.depends }
func (d *BeanDefinition) GetExports() []reflect.Type  { return d.exports }
func (d *BeanDefinition) IsConfiguration() bool       { return d.configuration }
func (d *BeanDefinition) GetIncludeMethod() []string  { return d.includeMethod }
func (d *BeanDefinition) GetExcludeMethod() []string  { return d.excludeMethod }

// Type 返回 bean 的类型。
func (d *BeanDefinition) Type() reflect.Type {
	return d.T
}

// Value 返回 bean 的值。
func (d *BeanDefinition) Value() reflect.Value {
	return d.V
}

// Interface 返回 bean 的真实值。
func (d *BeanDefinition) Interface() interface{} {
	return d.V.Interface()
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
	if d.F == nil {
		return "object bean"
	}
	return "constructor bean"
}

func (d *BeanDefinition) String() string {
	return fmt.Sprintf("%s name:%q %s", d.GetClass(), d.name, d.FileLine())
}

// Match 测试 bean 的类型全限定名和 bean 的名称是否都匹配。
func (d *BeanDefinition) Match(typeName string, beanName string) bool {

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

// Name 设置 bean 的名称。
func (d *BeanDefinition) Name(name string) *BeanDefinition {
	d.name = name
	return d
}

// On 设置 bean 的 Condition。
func (d *BeanDefinition) On(cond Condition) *BeanDefinition {
	d.cond = cond
	return d
}

// Order 设置 bean 的排序序号，值越小顺序越靠前(优先级越高)。
func (d *BeanDefinition) Order(order float32) *BeanDefinition {
	d.order = order
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
	if !gs_util.IsFuncType(fnType) {
		return false
	}
	if fnType.NumIn() != 1 || !gs_util.HasReceiver(fnType, beanValue) {
		return false
	}
	return gs_util.ReturnNothing(fnType) || gs_util.ReturnOnlyError(fnType)
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
	err := d.export(exports...)
	if err != nil {
		panic(err)
	}
	return d
}

func (d *BeanDefinition) export(exports ...interface{}) error {
	for _, o := range exports {
		var typ reflect.Type
		if t, ok := o.(reflect.Type); ok {
			typ = t
		} else { // 处理 (*error)(nil) 这种导出形式
			typ = gs_util.Indirect(reflect.TypeOf(o))
		}
		if typ.Kind() != reflect.Interface {
			return errors.New("only interface type can be exported")
		}
		exported := false
		for _, export := range d.exports {
			if typ == export {
				exported = true
				break
			}
		}
		if exported {
			continue
		}
		d.exports = append(d.exports, typ)
	}
	return nil
}

// Configuration 设置 bean 为配置类。
func (d *BeanDefinition) Configuration(includes []string, excludes []string) *BeanDefinition {
	d.configuration = true
	d.includeMethod = includes
	d.excludeMethod = excludes
	return d
}

// NewBean 普通函数注册时需要使用 reflect.ValueOf(fn) 形式以避免和构造函数发生冲突。
func NewBean(t reflect.Type, v reflect.Value, f Callable, name string, method bool, file string, line int) *BeanDefinition {
	return &BeanDefinition{
		T:        t,
		V:        v,
		F:        f,
		name:     name,
		typeName: gs_util.TypeName(t),
		status:   Default,
		method:   method,
		file:     file,
		line:     line,
	}
}
