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

// IndexArg represents an argument that has an index.
type IndexArg struct {
	n   int
	arg gs.Arg
}

// Index creates an IndexArg with the given index and argument.
func Index(n int, arg gs.Arg) IndexArg {
	return IndexArg{n: n, arg: arg}
}

// ValueArg represents an argument with a fixed value.
type ValueArg struct {
	v interface{}
}

// Nil returns a ValueArg with a value of nil.
func Nil() ValueArg {
	return ValueArg{v: nil}
}

// Value returns a ValueArg with the specified value.
func Value(v interface{}) ValueArg {
	return ValueArg{v: v}
}

// ArgList represents a list of arguments for a function.
type ArgList struct {
	fnType reflect.Type
	args   []gs.Arg
}

// NewArgList creates and validates an ArgList for the specified function.
func NewArgList(fnType reflect.Type, args []gs.Arg) (*ArgList, error) {

	// calculates the number of fixed arguments in the function.
	fixedArgCount := fnType.NumIn()
	if fnType.IsVariadic() {
		fixedArgCount--
	}

	// determines if the arguments use indexing.
	shouldIndex := func() bool {
		if len(args) == 0 {
			return false
		}
		_, ok := args[0].(IndexArg)
		return ok
	}()

	fnArgs := make([]gs.Arg, fixedArgCount)

	// processes the first argument separately to determine its type.
	if len(args) > 0 {
		if args[0] == nil {
			return nil, util.Errorf(macro.FileLine(), "the first arg must not be nil")
		}
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

	// processes the remaining arguments.
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

	// fills any unassigned fixed arguments with default values.
	for i := 0; i < fixedArgCount; i++ {
		if fnArgs[i] == nil {
			fnArgs[i] = ""
		}
	}

	return &ArgList{fnType: fnType, args: fnArgs}, nil
}

// get returns the processed argument values for the function call.
func (r *ArgList) get(ctx gs.ArgContext, fileLine string) ([]reflect.Value, error) {

	fnType := r.fnType
	numIn := fnType.NumIn()
	variadic := fnType.IsVariadic()
	result := make([]reflect.Value, 0)

	// processes each argument and convert it to a [reflect.Value].
	for idx, arg := range r.args {

		var t reflect.Type
		if variadic && idx >= numIn-1 {
			t = fnType.In(numIn - 1).Elem()
		} else {
			t = fnType.In(idx)
		}

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

// getArg processes an individual argument and returns its [reflect.Value] representation.
func (r *ArgList) getArg(ctx gs.ArgContext, arg gs.Arg, t reflect.Type, fileLine string) (reflect.Value, error) {

	var (
		err error
		tag string
	)

	description := fmt.Sprintf("arg:\"%v\" %s", arg, fileLine)
	syslog.Debugf("get value %s", description)
	defer func() {
		if err == nil {
			syslog.Debugf("get value %s success", description)
		} else {
			syslog.Debugf("get value %s error:%s", err.Error(), description)
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
	case string:
		tag = g
	default:
		tag = util.TypeName(g) + ":"
	}

	// binds property values based on the argument type.
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

	// wires dependent beans based on the argument type.
	if util.IsBeanReceiver(t) {
		v := reflect.New(t).Elem()
		if err = ctx.Wire(v, tag); err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	}

	return reflect.Value{}, util.Errorf(macro.FileLine(), "error type %s", t.String())
}

// OptionArg represents a binding for an option function argument.
type OptionArg struct {
	r *Callable
	c gs.Condition
}

// Option creates a binding for an option function argument.
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

// Condition sets a condition for invoking the option function.
func (arg *OptionArg) Condition(c gs.Condition) *OptionArg {
	arg.c = c
	return arg
}

// call invokes the option function if its condition is met.
func (arg *OptionArg) call(ctx gs.ArgContext) (reflect.Value, error) {

	var (
		ok  bool
		err error
	)

	syslog.Debugf("call option func %s", arg.r.fileLine)
	defer func() {
		if err == nil {
			syslog.Debugf("call option func success %s", arg.r.fileLine)
		} else {
			syslog.Debugf("call option func error %s %s", err.Error(), arg.r.fileLine)
		}
	}()

	// checks if the condition is met.
	if arg.c != nil {
		ok, err = ctx.Matches(arg.c)
		if err != nil {
			return reflect.Value{}, err
		} else if !ok {
			return reflect.Value{}, nil
		}
	}

	// invokes the function and return its result.
	out, err := arg.r.Call(ctx)
	if err != nil {
		return reflect.Value{}, err
	}
	return out[0], nil
}

// Callable wraps a function and its binding arguments.
type Callable struct {
	fn       interface{}
	fnType   reflect.Type
	argList  *ArgList
	fileLine string
}

// MustBind binds arguments to a function and panics if an error occurs.
func MustBind(fn interface{}, args ...gs.Arg) *Callable {
	r, err := Bind(fn, args, 1)
	if err != nil {
		panic(err)
	}
	return r
}

// Bind creates a Callable by binding arguments to a function.
func Bind(fn interface{}, args []gs.Arg, skip int) (*Callable, error) {

	fnType := reflect.TypeOf(fn)
	argList, err := NewArgList(fnType, args)
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

// Arg retrieves the i-th binding argument.
func (r *Callable) Arg(i int) (gs.Arg, bool) {
	if i >= len(r.argList.args) {
		return nil, false
	}
	return r.argList.args[i], true
}

// In retrieves the i-th input parameter type of the function.
func (r *Callable) In(i int) (reflect.Type, bool) {
	if i >= r.fnType.NumIn() {
		return nil, false
	}
	return r.fnType.In(i), true
}

// Call invokes the function with its bound arguments processed in the IoC container.
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
