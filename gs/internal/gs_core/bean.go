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

package gs_core

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util"
)

// NewBean 普通函数注册时需要使用 reflect.ValueOf(fn) 形式以避免和构造函数发生冲突。
func NewBean(objOrCtor interface{}, ctorArgs ...gs.Arg) *gs.BeanDefinition {

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

	// Ensure the value is valid and not nil
	if !v.IsValid() || v.IsNil() {
		panic("bean can't be nil")
	}

	var f *gs_arg.Callable
	_, file, line, _ := runtime.Caller(1)

	// If objOrCtor is a function (not from reflect.Value),
	// process it as a constructor
	if !fromValue && t.Kind() == reflect.Func {

		if !util.IsConstructor(t) {
			t1 := "func(...)bean"
			t2 := "func(...)(bean, error)"
			panic(fmt.Sprintf("constructor should be %s or %s", t1, t2))
		}

		// Bind the constructor arguments
		var err error
		f, err = gs_arg.NewCallable(objOrCtor, ctorArgs)
		if err != nil {
			panic(err)
		}

		var in0 reflect.Type
		if t.NumIn() > 0 {
			in0 = t.In(0)
		}

		// Obtain the return type of the constructor
		out0 := t.Out(0)
		v = reflect.New(out0)
		if util.IsBeanType(out0) {
			v = v.Elem()
		}

		t = v.Type()
		if !util.IsBeanType(t) {
			panic("bean must be ref type")
		}

		// Extract function name for naming the bean
		fnPtr := reflect.ValueOf(objOrCtor).Pointer()
		fnInfo := runtime.FuncForPC(fnPtr)
		funcName := fnInfo.Name()
		name = funcName[strings.LastIndex(funcName, "/")+1:]
		name = name[strings.Index(name, ".")+1:]
		if name[0] == '(' {
			name = name[strings.Index(name, ".")+1:]
		}

		// Check if the function is a method and set a condition if needed
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

	if t.Kind() == reflect.Ptr && !util.IsPropBindingTarget(t.Elem()) {
		panic("bean should be *val but not *ref")
	}

	// Type.String() 一般返回 *pkg.Type 形式的字符串，
	// 我们只取最后的类型名，如有需要请自定义 bean 名称。
	if name == "" {
		s := strings.Split(t.String(), ".")
		name = strings.TrimPrefix(s[len(s)-1], "*")
	}

	d := gs_bean.NewBean(t, v, f, name)
	d.SetFileLine(file, line)

	bd := gs.NewBeanDefinition(d)
	if cond != nil {
		bd.Condition(cond)
	}
	return bd
}
