package gs_bean

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
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
	V reflect.Value    // 值
	T reflect.Type     // 类型
	F *gs_arg.Callable // 构造函数

	file string // 注册点所在文件
	line int    // 注册点所在行数

	name     string                 // 名称
	typeName string                 // 原始类型的全限定名
	status   BeanStatus             // 状态
	primary  bool                   // 是否为主版本
	method   bool                   // 是否为成员方法
	cond     gs_cond.Condition      // 判断条件
	order    float32                // 收集时的顺序
	init     interface{}            // 初始化函数
	destroy  interface{}            // 销毁函数
	depends  []gs_util.BeanSelector // 间接依赖项
	exports  []reflect.Type         // 导出的接口
}

func (d *BeanDefinition) GetName() string                    { return d.name }
func (d *BeanDefinition) GetTypeName() string                { return d.typeName }
func (d *BeanDefinition) GetStatus() BeanStatus              { return d.status }
func (d *BeanDefinition) SetStatus(status BeanStatus)        { d.status = status }
func (d *BeanDefinition) IsPrimary() bool                    { return d.primary }
func (d *BeanDefinition) IsMethod() bool                     { return d.method }
func (d *BeanDefinition) GetCond() gs_cond.Condition         { return d.cond }
func (d *BeanDefinition) GetOrder() float32                  { return d.order }
func (d *BeanDefinition) GetInit() interface{}               { return d.init }
func (d *BeanDefinition) GetDestroy() interface{}            { return d.destroy }
func (d *BeanDefinition) GetDepends() []gs_util.BeanSelector { return d.depends }
func (d *BeanDefinition) GetExports() []reflect.Type         { return d.exports }

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
func (d *BeanDefinition) On(cond gs_cond.Condition) *BeanDefinition {
	d.cond = cond
	return d
}

// Order 设置 bean 的排序序号，值越小顺序越靠前(优先级越高)。
func (d *BeanDefinition) Order(order float32) *BeanDefinition {
	d.order = order
	return d
}

// DependsOn 设置 bean 的间接依赖项。
func (d *BeanDefinition) DependsOn(selectors ...gs_util.BeanSelector) *BeanDefinition {
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

// NewBean 普通函数注册时需要使用 reflect.ValueOf(fn) 形式以避免和构造函数发生冲突。
func NewBean(objOrCtor interface{}, ctorArgs ...gs_arg.Arg) *BeanDefinition {

	var v reflect.Value
	var fromValue bool
	var method bool
	var name string

	switch i := objOrCtor.(type) {
	case reflect.Value:
		fromValue = true
		v = i
	default:
		v = reflect.ValueOf(i)
	}

	t := v.Type()
	if !gs_util.IsBeanType(t) {
		panic(errors.New("bean must be ref type"))
	}

	if !v.IsValid() || v.IsNil() {
		panic(errors.New("bean can't be nil"))
	}

	const skip = 2
	var f *gs_arg.Callable
	_, file, line, _ := runtime.Caller(skip)

	// 以 reflect.ValueOf(fn) 形式注册的函数被视为函数对象 bean 。
	if !fromValue && t.Kind() == reflect.Func {

		if !gs_util.IsConstructor(t) {
			t1 := "func(...)bean"
			t2 := "func(...)(bean, error)"
			panic(fmt.Errorf("constructor should be %s or %s", t1, t2))
		}

		var err error
		f, err = gs_arg.Bind(objOrCtor, ctorArgs, skip)
		if err != nil {
			panic(err)
		}

		out0 := t.Out(0)
		v = reflect.New(out0)
		if gs_util.IsBeanType(out0) {
			v = v.Elem()
		}

		t = v.Type()
		if !gs_util.IsBeanType(t) {
			panic(errors.New("bean must be ref type"))
		}

		// 成员方法一般是 xxx/gs_test.(*Server).Consumer 形式命名
		fnPtr := reflect.ValueOf(objOrCtor).Pointer()
		fnInfo := runtime.FuncForPC(fnPtr)
		funcName := fnInfo.Name()
		name = funcName[strings.LastIndex(funcName, "/")+1:]
		name = name[strings.Index(name, ".")+1:]
		if name[0] == '(' {
			name = name[strings.Index(name, ".")+1:]
		}
		method = strings.LastIndexByte(fnInfo.Name(), ')') > 0
	}

	if t.Kind() == reflect.Ptr && !gs_util.IsValueType(t.Elem()) {
		panic(errors.New("bean should be *val but not *ref"))
	}

	// Type.String() 一般返回 *pkg.Type 形式的字符串，
	// 我们只取最后的类型名，如有需要请自定义 bean 名称。
	if name == "" {
		s := strings.Split(t.String(), ".")
		name = strings.TrimPrefix(s[len(s)-1], "*")
	}

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
