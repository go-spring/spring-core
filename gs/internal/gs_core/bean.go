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

package gs_core

import (
	"errors"
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
func NewBean(objOrCtor interface{}, ctorArgs ...gs.Arg) *gs.UnregisteredBean {

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
		panic(errors.New("bean must be ref type"))
	}

	if !v.IsValid() || v.IsNil() {
		panic(errors.New("bean can't be nil"))
	}

	const skip = 2
	var f gs.Callable
	_, file, line, _ := runtime.Caller(skip)

	// 以 reflect.ValueOf(fn) 形式注册的函数被视为函数对象 bean 。
	if !fromValue && t.Kind() == reflect.Func {

		if !util.IsConstructor(t) {
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
		if util.IsBeanType(out0) {
			v = v.Elem()
		}

		t = v.Type()
		if !util.IsBeanType(t) {
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
		method := strings.LastIndexByte(fnInfo.Name(), ')') > 0
		if method {
			selector, ok := f.Arg(0)
			if !ok || selector == "" {
				selector, _ = f.In(0)
			}
			cond = gs_cond.OnBean(selector)
		}
	}

	if t.Kind() == reflect.Ptr && !util.IsValueType(t.Elem()) {
		panic(errors.New("bean should be *val but not *ref"))
	}

	// Type.String() 一般返回 *pkg.Type 形式的字符串，
	// 我们只取最后的类型名，如有需要请自定义 bean 名称。
	if name == "" {
		s := strings.Split(t.String(), ".")
		name = strings.TrimPrefix(s[len(s)-1], "*")
	}

	d := gs_bean.NewBean(t, v, f, name, file, line)
	return gs.NewUnregisteredBean(d).On(cond)
}
