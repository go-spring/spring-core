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

package gs_core_test

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_core"
	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/assert"
)

// newBean 该方法是为了平衡调用栈的深度，一般情况下 gs.NewBean 不应该被直接使用。
func newBean(objOrCtor interface{}, ctorArgs ...gs.Arg) *gs.BeanDefinition {
	return gs_core.NewBean(objOrCtor, ctorArgs...)
}

// func TestParseSingletonTag(t *testing.T) {
//
//	data := map[string]SingletonTag{
//		"?":      {"", "", true},
//		"i":      {"", "i", false},
//		"i?":     {"", "i", true},
//		":i":     {"", "i", false},
//		":i?":    {"", "i", true},
//		"int:i":  {"int", "i", false},
//		"int:i?": {"int", "i", true},
//		"int:":   {"int", "", false},
//		"int:?":  {"int", "", true},
//	}
//
//	for k, v := range data {
//		tag := parseSingletonTag(k)
//		Equal(t, tag, v)
//	}
// }
//
// func TestParseBeanTag(t *testing.T) {
//
//	data := map[string]collectionTag{
//		"?":   {[]SingletonTag{}, true},
//	}
//
//	for k, v := range data {
//		tag := ParseCollectionTag(k)
//		Equal(t, tag, v)
//	}
// }

func TestIsFuncBeanType(t *testing.T) {

	type S struct{}
	type OptionFunc func(*S)

	data := map[reflect.Type]bool{
		reflect.TypeOf((func())(nil)):            false,
		reflect.TypeOf((func(int))(nil)):         false,
		reflect.TypeOf((func(int, int))(nil)):    false,
		reflect.TypeOf((func(int, ...int))(nil)): false,

		reflect.TypeOf((func() int)(nil)):          true,
		reflect.TypeOf((func() (int, int))(nil)):   false,
		reflect.TypeOf((func() (int, error))(nil)): true,

		reflect.TypeOf((func(int) int)(nil)):         true,
		reflect.TypeOf((func(int, int) int)(nil)):    true,
		reflect.TypeOf((func(int, ...int) int)(nil)): true,

		reflect.TypeOf((func(int) (int, error))(nil)):         true,
		reflect.TypeOf((func(int, int) (int, error))(nil)):    true,
		reflect.TypeOf((func(int, ...int) (int, error))(nil)): true,

		reflect.TypeOf((func() S)(nil)):          true,
		reflect.TypeOf((func() *S)(nil)):         true,
		reflect.TypeOf((func() (S, error))(nil)): true,

		reflect.TypeOf((func(OptionFunc) (*S, error))(nil)):    true,
		reflect.TypeOf((func(...OptionFunc) (*S, error))(nil)): true,
	}

	for k, v := range data {
		ok := util.IsConstructor(k)
		assert.Equal(t, ok, v)
	}
}

func TestObjectBean(t *testing.T) {

	// t.Run("bean must be ref type", func(t *testing.T) {
	//
	// 	data := []func(){
	// 		func() { newBean([...]int{0}) },
	// 		func() { newBean(false) },
	// 		func() { newBean(3) },
	// 		func() { newBean("3") },
	// 		func() { newBean(BeanZero{}) },
	// 		func() { newBean(pkg2.SamePkg{}) },
	// 	}
	//
	// 	for _, fn := range data {
	// 		assert.Panic(t, fn, "bean must be ref type")
	// 	}
	// })

	t.Run("valid bean", func(t *testing.T) {
		newBean(make(chan int))
		newBean(reflect.ValueOf(func() {}))
		newBean(&BeanZero{})
	})

	t.Run("check name && typename", func(t *testing.T) {
		data := map[*gs.BeanDefinition]struct {
			name string
		}{
			newBean(io.Writer(os.Stdout)):  {"File"},
			newBean(newHistoryTeacher("")): {"historyTeacher"},
		}
		for bd, v := range data {
			assert.Equal(t, bd.BeanRegistration().(*gs_bean.BeanDefinition).Name(), v.name)
		}
	})
}

func TestConstructorBean(t *testing.T) {

	bd := newBean(NewStudent)
	assert.Equal(t, bd.BeanRegistration().Type().String(), "*gs_core_test.Student")

	bd = newBean(NewPtrStudent)
	assert.Equal(t, bd.BeanRegistration().Type().String(), "*gs_core_test.Student")

	// mapFn := func() map[int]string { return make(map[int]string) }
	// bd = newBean(mapFn)
	// assert.Equal(t, bd.Type().String(), "*map[int]string")

	// sliceFn := func() []int { return make([]int, 1) }
	// bd = newBean(sliceFn)
	// assert.Equal(t, bd.Type().String(), "*[]int")

	funcFn := func() func(int) { return nil }
	bd = newBean(funcFn)
	assert.Equal(t, bd.BeanRegistration().Type().String(), "func(int)")

	interfaceFn := func(name string) Teacher { return newHistoryTeacher(name) }
	bd = newBean(interfaceFn)
	assert.Equal(t, bd.BeanRegistration().Type().String(), "gs_core_test.Teacher")

	// assert.Panic(t, func() {
	// 	_ = newBean(func() (*int, *int) { return nil, nil })
	// }, "constructor should be func\\(...\\)bean or func\\(...\\)\\(bean, error\\)")
}

type Runner interface {
	Run()
}

type RunStringer struct {
}

func NewRunStringer() fmt.Stringer {
	return &RunStringer{}
}

func (rs *RunStringer) String() string {
	return "RunStringer"
}

func (rs *RunStringer) Run() {
	fmt.Println("RunStringer")
}

func TestInterface(t *testing.T) {

	t.Run("interface type", func(t *testing.T) {
		fnValue := reflect.ValueOf(NewRunStringer)
		fmt.Println(fnValue.Type())
		retValue := fnValue.Call([]reflect.Value{})[0]
		fmt.Println(retValue.Type(), retValue.Elem().Type())
		r := new(Runner)
		fmt.Println(reflect.TypeOf(r).Elem())
		ok := retValue.Elem().Type().AssignableTo(reflect.TypeOf(r).Elem())
		fmt.Println(ok)
	})

	fn := func() io.Reader {
		return os.Stdout
	}

	fnType := reflect.TypeOf(fn)
	// func() io.Reader
	fmt.Println(fnType)

	outType := fnType.Out(0)
	// io.Reader
	fmt.Println(outType)

	fnValue := reflect.ValueOf(fn)
	out := fnValue.Call([]reflect.Value{})

	outValue := out[0]
	// 0xc000010010 io.Reader
	fmt.Println(outValue, outValue.Type())
	// &{0xc0000a4000} *os.File
	fmt.Println(outValue.Elem(), outValue.Elem().Type())
}

type callable interface {
	Call() int
}

type caller struct {
	i int
}

func (c *caller) Call() int {
	return c.i
}

func TestInterfaceMethod(t *testing.T) {
	c := callable(&caller{3})
	fmt.Println(c.Call())
}

func TestVariadicFunction(t *testing.T) {

	fn := func(a string, i ...int) {
		fmt.Println(a, i)
	}

	typ := reflect.TypeOf(fn)
	fmt.Println(typ, typ.IsVariadic())

	for i := 0; i < typ.NumIn(); i++ {
		in := typ.In(i)
		fmt.Println(in)
	}

	fnValue := reflect.ValueOf(fn)
	fnValue.Call([]reflect.Value{
		reflect.ValueOf("string"),
		reflect.ValueOf(3),
		reflect.ValueOf(4),
	})

	c := caller{6}
	fmt.Println((*caller).Call(&c))

	typ = reflect.TypeOf((*caller).Call)
	fmt.Println(typ)

	var arr []int
	fmt.Println(len(arr))
}

type reCaller caller

func TestNumMethod(t *testing.T) {

	typ0 := reflect.TypeOf(new(caller))
	assert.Equal(t, typ0.NumMethod(), 1)

	typ1 := reflect.TypeOf(new(reCaller))
	assert.Equal(t, typ1.NumMethod(), 0)

	typ2 := reflect.TypeOf((*reCaller)(nil))
	assert.True(t, typ1 == typ2)
}
