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

//go:generate mockgen -build_flags="-mod=mod" -package=arg -source=arg.go -destination=arg_mock.go

// Package arg 用于实现函数参数绑定。
package gs_arg

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/syslog"
	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/macro"
)

// IndexArg is an Arg that has an index.
type IndexArg struct {
	n   int
	arg gs.Arg
}

// Index returns an IndexArg.
func Index(n int, arg gs.Arg) IndexArg {
	return IndexArg{n: n, arg: arg}
}

// R0 returns an IndexArg with index 0.
func R0(arg gs.Arg) IndexArg { return Index(0, arg) }

// R1 returns an IndexArg with index 1.
func R1(arg gs.Arg) IndexArg { return Index(1, arg) }

// R2 returns an IndexArg with index 2.
func R2(arg gs.Arg) IndexArg { return Index(2, arg) }

// R3 returns an IndexArg with index 3.
func R3(arg gs.Arg) IndexArg { return Index(3, arg) }

// R4 returns an IndexArg with index 4.
func R4(arg gs.Arg) IndexArg { return Index(4, arg) }

// R5 returns an IndexArg with index 5.
func R5(arg gs.Arg) IndexArg { return Index(5, arg) }

// R6 returns an IndexArg with index 6.
func R6(arg gs.Arg) IndexArg { return Index(6, arg) }

// ValueArg is an Arg that has a value.
type ValueArg struct {
	v interface{}
}

// Nil return a ValueArg with a value of nil.
func Nil() ValueArg {
	return ValueArg{v: nil}
}

// Value return a ValueArg with a value of v.
func Value(v interface{}) ValueArg {
	return ValueArg{v: v}
}

// argList stores the arguments of a function.
type argList struct {
	fnType reflect.Type
	args   []gs.Arg
}

// newArgList returns a new argList.
func newArgList(fnType reflect.Type, args []gs.Arg) (*argList, error) {

	fixedArgCount := fnType.NumIn()
	if fnType.IsVariadic() {
		fixedArgCount--
	}

	shouldIndex := func() bool {
		if len(args) == 0 {
			return false
		}
		_, ok := args[0].(IndexArg)
		return ok
	}()

	fnArgs := make([]gs.Arg, fixedArgCount)

	if len(args) > 0 {
		switch arg := args[0].(type) {
		case *OptionArg:
			fnArgs = append(fnArgs, arg)
		case IndexArg:
			if arg.n < 0 || arg.n >= fixedArgCount {
				return nil, util.Errorf(macro.FileLine(), "arg index %d exceeds max index %d", arg.n, fixedArgCount)
			} else {
				fnArgs[arg.n] = arg.arg
			}
		default:
			if fixedArgCount > 0 {
				fnArgs[0] = arg
			} else if fnType.IsVariadic() {
				fnArgs = append(fnArgs, arg)
			} else {
				return nil, util.Errorf(macro.FileLine(), "function has no args but given %d", len(args))
			}
		}
	}

	for i := 1; i < len(args); i++ {
		switch arg := args[i].(type) {
		case *OptionArg:
			fnArgs = append(fnArgs, arg)
		case IndexArg:
			if !shouldIndex {
				return nil, util.Errorf(macro.FileLine(), "the Args must have or have no index")
			}
			if arg.n < 0 || arg.n >= fixedArgCount {
				return nil, util.Errorf(macro.FileLine(), "arg index %d exceeds max index %d", arg.n, fixedArgCount)
			} else if fnArgs[arg.n] != nil {
				return nil, util.Errorf(macro.FileLine(), "found same index %d", arg.n)
			} else {
				fnArgs[arg.n] = arg.arg
			}
		default:
			if shouldIndex {
				return nil, util.Errorf(macro.FileLine(), "the Args must have or have no index")
			}
			if i < fixedArgCount {
				fnArgs[i] = arg
			} else if fnType.IsVariadic() {
				fnArgs = append(fnArgs, arg)
			} else {
				return nil, util.Errorf(macro.FileLine(), "the count %d of Args exceeds max index %d", len(args), fixedArgCount)
			}
		}
	}

	for i := 0; i < fixedArgCount; i++ {
		if fnArgs[i] == nil {
			fnArgs[i] = ""
		}
	}

	return &argList{fnType: fnType, args: fnArgs}, nil
}

// get returns all processed Args value. fileLine is the binding position of Callable.
func (r *argList) get(ctx gs.ArgContext, fileLine string) ([]reflect.Value, error) {

	fnType := r.fnType
	numIn := fnType.NumIn()
	variadic := fnType.IsVariadic()
	result := make([]reflect.Value, 0)

	for idx, arg := range r.args {

		var t reflect.Type
		if variadic && idx >= numIn-1 {
			t = fnType.In(numIn - 1).Elem()
		} else {
			t = fnType.In(idx)
		}

		// option arg may not return a value when the condition is not met.
		v, err := r.getArg(ctx, arg, t, fileLine)
		if err != nil {
			return nil, util.Wrapf(err, macro.FileLine(), "returns error when getting %d arg", idx)
		}
		if v.IsValid() {
			result = append(result, v)
		}
	}

	return result, nil
}

func (r *argList) getArg(ctx gs.ArgContext, arg gs.Arg, t reflect.Type, fileLine string) (reflect.Value, error) {

	var (
		err error
		tag string
	)

	description := fmt.Sprintf("arg:\"%v\" %s", arg, fileLine)
	syslog.Debug("get value %s", description)
	defer func() {
		if err == nil {
			syslog.Debug("get value success %s", description)
		} else {
			syslog.Debug("get value error %s %s", err.Error(), description)
		}
	}()

	switch g := arg.(type) {
	case *Callable:
		if results, err := g.Call(ctx); err != nil {
			return reflect.Value{}, util.Wrapf(err, macro.FileLine(), "")
		} else if len(results) < 1 {
			return reflect.Value{}, errors.New("")
		} else {
			return results[0], nil
		}
	case ValueArg:
		if g.v == nil {
			return reflect.Zero(t), nil
		}
		return reflect.ValueOf(g.v), nil
	case *OptionArg:
		return g.call(ctx)
	case *gs.BeanDefinition:
		tag = g.ID()
	case string:
		tag = g
	default:
		tag = util.TypeName(g) + ":"
	}

	// binds properties value by the "value" tag.
	if util.IsValueType(t) {
		if tag == "" {
			tag = "${}"
		}
		v := reflect.New(t).Elem()
		if err = ctx.Bind(v, tag); err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	}

	// wires dependent beans by the "autowire" tag.
	if util.IsBeanReceiver(t) {
		v := reflect.New(t).Elem()
		if err = ctx.Wire(v, tag); err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	}

	return reflect.Value{}, util.Errorf(macro.FileLine(), "error type %s", t.String())
}

// OptionArg Option 函数的参数绑定。
type OptionArg struct {
	r *Callable
	c gs.Condition
}

// Provide 为 Option 方法绑定运行时参数。
func Provide(fn interface{}, args ...gs.Arg) *Callable {
	r, err := Bind(fn, args, 1)
	if err != nil {
		panic(err)
	}
	return r
}

// Option 返回 Option 函数的参数绑定。
func Option(fn interface{}, args ...gs.Arg) *OptionArg {

	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func || t.NumOut() != 1 {
		panic(errors.New("invalid option func"))
	}

	r, err := Bind(fn, args, 1)
	if err != nil {
		panic(err)
	}
	return &OptionArg{r: r}
}

// On 设置一个 gs_cond.Condition 对象。
func (arg *OptionArg) On(c gs.Condition) *OptionArg {
	arg.c = c
	return arg
}

func (arg *OptionArg) call(ctx gs.ArgContext) (reflect.Value, error) {

	var (
		ok  bool
		err error
	)

	syslog.Debug("call option func %s", arg.r.fileLine)
	defer func() {
		if err == nil {
			syslog.Debug("call option func success %s", arg.r.fileLine)
		} else {
			syslog.Debug("call option func error %s %s", err.Error(), arg.r.fileLine)
		}
	}()

	if arg.c != nil {
		ok, err = ctx.Matches(arg.c)
		if err != nil {
			return reflect.Value{}, err
		} else if !ok {
			return reflect.Value{}, nil
		}
	}

	out, err := arg.r.Call(ctx)
	if err != nil {
		return reflect.Value{}, err
	}
	return out[0], nil
}

// Callable wrappers a function and its binding arguments, then you can invoke
// the Call method of Callable to get the function's result.
type Callable struct {
	fn       interface{}
	fnType   reflect.Type
	argList  *argList
	fileLine string
}

// Bind returns a Callable that wrappers a function and its binding arguments.
// The argument skip is the number of frames to skip over.
func Bind(fn interface{}, args []gs.Arg, skip int) (*Callable, error) {

	fnType := reflect.TypeOf(fn)
	argList, err := newArgList(fnType, args)
	if err != nil {
		return nil, err
	}

	_, file, line, _ := runtime.Caller(skip + 1)
	r := &Callable{
		fn:       fn,
		fnType:   fnType,
		argList:  argList,
		fileLine: fmt.Sprintf("%s:%d", file, line),
	}
	return r, nil
}

// Arg returns the ith binding argument.
func (r *Callable) Arg(i int) (gs.Arg, bool) {
	if i >= len(r.argList.args) {
		return nil, false
	}
	return r.argList.args[i], true
}

func (r *Callable) In(i int) (reflect.Type, bool) {
	if i >= r.fnType.NumIn() {
		return nil, false
	}
	return r.fnType.In(i), true
}

// Call invokes the function with its binding arguments processed in the IoC
// container. If the function returns an error, then the Call returns it.
func (r *Callable) Call(ctx gs.ArgContext) ([]reflect.Value, error) {

	in, err := r.argList.get(ctx, r.fileLine)
	if err != nil {
		return nil, err
	}

	out := reflect.ValueOf(r.fn).Call(in)
	n := len(out)
	if n == 0 {
		return out, nil
	}

	o := out[n-1]
	if util.IsErrorType(o.Type()) {
		if i := o.Interface(); i != nil {
			return out[:n-1], i.(error)
		}
		return out[:n-1], nil
	}
	return out, nil
}
