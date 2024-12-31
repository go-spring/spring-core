package gs

import (
	"context"
	"reflect"

	"github.com/go-spring/spring-core/conf"
)

// A BeanSelector can be the ID of a bean, a `reflect.Type`, a pointer such as
// `(*error)(nil)`, or a BeanDefinition value.
type BeanSelector interface{}

// CondContext defines some methods of IoC container that conditions use.
type CondContext interface {
	// Has returns whether the IoC container has a property.
	Has(key string) bool
	// Prop returns the property's value when the IoC container has it, or
	// returns empty string when the IoC container doesn't have it.
	Prop(key string, opts ...conf.GetOption) string
	// Find returns bean definitions that matched with the bean selector.
	Find(selector BeanSelector) ([]*BeanDefinition, error)
}

// Condition is used when registering a bean to determine whether it's valid.
type Condition interface {
	Matches(ctx CondContext) (bool, error)
}

// Arg 用于为函数参数提供绑定值。可以是 bean.Selector 类型，表示注入 bean ；
// 可以是 ${X:=Y} 形式的字符串，表示属性绑定或者注入 bean ；可以是 ValueArg
// 类型，表示不从 IoC 容器获取而是用户传入的普通值；可以是 IndexArg 类型，表示
// 带有下标的参数绑定；可以是 *optionArg 类型，用于为 Option 方法提供参数绑定。
type Arg interface{}

// ArgContext defines some methods of IoC container that Callable use.
type ArgContext interface {
	// Matches returns true when the Condition returns true,
	// and returns false when the Condition returns false.
	Matches(c Condition) (bool, error)
	// Bind binds properties value by the "value" tag.
	Bind(v reflect.Value, tag string) error
	// Wire wires dependent beans by the "autowire" tag.
	Wire(v reflect.Value, tag string) error
}

type Callable interface {
	Arg(i int) (Arg, bool)
	In(i int) (reflect.Type, bool)
	Call(ctx ArgContext) ([]reflect.Value, error)
}

type GroupFunc func(p conf.ReadOnlyProperties) ([]*BeanDefinition, error)

// Context 提供了一些在 IoC 容器启动后基于反射获取和使用 property 与 bean 的接
// 口。因为很多人会担心在运行时大量使用反射会降低程序性能，所以命名为 Context，取
// 其诱人但危险的含义。事实上，这些在 IoC 容器启动后使用属性绑定和依赖注入的方案，
// 都可以转换为启动阶段的方案以提高程序的性能。
// 另一方面，为了统一 Container 和 App 两种启动方式下这些方法的使用方式，需要提取
// 出一个可共用的接口来，也就是说，无论程序是 Container 方式启动还是 App 方式启动，
// 都可以在需要使用这些方法的地方注入一个 Context 对象而不是 Container 对象或者
// App 对象，从而实现使用方式的统一。
type Context interface {
	Context() context.Context
	Keys() []string
	Has(key string) bool
	Prop(key string, opts ...conf.GetOption) string
	Resolve(s string) (string, error)
	Bind(i interface{}, opts ...conf.BindArg) error
	Get(i interface{}, selectors ...BeanSelector) error
	Wire(objOrCtor interface{}, ctorArgs ...Arg) (interface{}, error)
	Invoke(fn interface{}, args ...Arg) ([]interface{}, error)
	Go(fn func(ctx context.Context))
}
