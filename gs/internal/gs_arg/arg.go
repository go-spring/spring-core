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

// Package gs_arg provides a set of tools for working with function arguments.
package gs_arg

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/errutil"
)

// TagArg represents an argument that has a tag for binding or autowiring.
type TagArg struct {
	Tag string
}

// Tag creates a TagArg with the given tag.
func Tag(tag string) gs.Arg {
	return TagArg{Tag: tag}
}

// GetArgValue returns the value of the argument based on its type.
func (arg TagArg) GetArgValue(ctx gs.ArgContext, t reflect.Type) (reflect.Value, error) {

	// Binds property values based on the argument type.
	if util.IsPropBindingTarget(t) {
		v := reflect.New(t).Elem()
		if err := ctx.Bind(v, arg.Tag); err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	}

	// Wires dependent beans based on the argument type.
	if util.IsBeanInjectionTarget(t) {
		v := reflect.New(t).Elem()
		if err := ctx.Wire(v, arg.Tag); err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	}

	// If none of the conditions match, return an error.
	err := fmt.Errorf("error type %s", t.String())
	return reflect.Value{}, errutil.WrapError(err, "get arg error: %v", arg.Tag)
}

// IndexArg represents an argument that has an index.
type IndexArg struct {
	Idx int    // Index of the argument.
	Arg gs.Arg // The actual argument value.
}

// Index creates an IndexArg with the given index and argument.
func Index(n int, arg gs.Arg) gs.Arg {
	return IndexArg{Idx: n, Arg: arg}
}

// GetArgValue is not implemented for IndexArg, it panics if called.
func (arg IndexArg) GetArgValue(ctx gs.ArgContext, t reflect.Type) (reflect.Value, error) {
	panic(util.UnimplementedMethod)
}

// ValueArg represents an argument with a fixed value.
type ValueArg struct {
	v interface{} // The fixed value associated with this argument.
}

// Nil returns a ValueArg with a value of nil.
func Nil() gs.Arg {
	return ValueArg{v: nil}
}

// Value returns a ValueArg with the specified value.
func Value(v interface{}) gs.Arg {
	return ValueArg{v: v}
}

// GetArgValue returns the value of the fixed argument.
func (arg ValueArg) GetArgValue(ctx gs.ArgContext, t reflect.Type) (reflect.Value, error) {
	if arg.v == nil {
		return reflect.Zero(t), nil
	}
	return reflect.ValueOf(arg.v), nil
}

// ArgList represents a list of arguments for a function.
type ArgList struct {
	fnType reflect.Type // Type of the function to be invoked.
	args   []gs.Arg     // List of arguments for the function.
}

// NewArgList creates and validates an ArgList for the specified function.
func NewArgList(fnType reflect.Type, args []gs.Arg) (*ArgList, error) {

	// Calculates the number of fixed arguments in the function.
	fixedArgCount := fnType.NumIn()
	if fnType.IsVariadic() {
		fixedArgCount--
	}

	// Determines if the arguments use indexing.
	shouldIndex := func() bool {
		if len(args) == 0 {
			return false
		}
		_, ok := args[0].(IndexArg)
		return ok
	}()

	fnArgs := make([]gs.Arg, fixedArgCount)

	// Processes the first argument separately to determine its type.
	if len(args) > 0 {
		if args[0] == nil {
			err := errors.New("the first arg must not be nil")
			return nil, errutil.WrapError(err, "%v", args)
		}
		switch arg := args[0].(type) {
		case *OptionArg:
			fnArgs = append(fnArgs, arg)
		case IndexArg:
			if arg.Idx < 0 || arg.Idx >= fixedArgCount {
				err := fmt.Errorf("arg index %d exceeds max index %d", arg.Idx, fixedArgCount)
				return nil, errutil.WrapError(err, "%v", args)
			} else {
				fnArgs[arg.Idx] = arg.Arg
			}
		default:
			if fixedArgCount > 0 {
				fnArgs[0] = arg
			} else if fnType.IsVariadic() {
				fnArgs = append(fnArgs, arg)
			} else {
				err := fmt.Errorf("function has no args but given %d", len(args))
				return nil, errutil.WrapError(err, "%v", args)
			}
		}
	}

	// Processes the remaining arguments.
	for i := 1; i < len(args); i++ {
		switch arg := args[i].(type) {
		case *OptionArg:
			fnArgs = append(fnArgs, arg)
		case IndexArg:
			if !shouldIndex {
				err := fmt.Errorf("the Args must have or have no index")
				return nil, errutil.WrapError(err, "%v", args)
			}
			if arg.Idx < 0 || arg.Idx >= fixedArgCount {
				err := fmt.Errorf("arg index %d exceeds max index %d", arg.Idx, fixedArgCount)
				return nil, errutil.WrapError(err, "%v", args)
			} else if fnArgs[arg.Idx] != nil {
				err := fmt.Errorf("found same index %d", arg.Idx)
				return nil, errutil.WrapError(err, "%v", args)
			} else {
				fnArgs[arg.Idx] = arg.Arg
			}
		default:
			if shouldIndex {
				err := fmt.Errorf("the Args must have or have no index")
				return nil, errutil.WrapError(err, "%v", args)
			}
			if i < fixedArgCount {
				fnArgs[i] = arg
			} else if fnType.IsVariadic() {
				fnArgs = append(fnArgs, arg)
			} else {
				err := fmt.Errorf("the count %d of Args exceeds max index %d", len(args), fixedArgCount)
				return nil, errutil.WrapError(err, "%v", args)
			}
		}
	}

	// Fills any unassigned fixed arguments with default values.
	for i := 0; i < fixedArgCount; i++ {
		if fnArgs[i] == nil {
			fnArgs[i] = Tag("")
		}
	}

	return &ArgList{fnType: fnType, args: fnArgs}, nil
}

// get returns the processed argument values for the function call.
func (r *ArgList) get(ctx gs.ArgContext) ([]reflect.Value, error) {

	fnType := r.fnType
	numIn := fnType.NumIn()
	variadic := fnType.IsVariadic()
	result := make([]reflect.Value, 0)

	// Processes each argument and converts it to a [reflect.Value].
	for idx, arg := range r.args {

		var t reflect.Type
		if variadic && idx >= numIn-1 {
			t = fnType.In(numIn - 1).Elem()
		} else {
			t = fnType.In(idx)
		}

		v, err := arg.GetArgValue(ctx, t)
		if err != nil {
			err = errutil.WrapError(err, "returns error when getting %d arg", idx)
			return nil, errutil.WrapError(err, "get arg list error: %v", arg)
		}
		if v.IsValid() {
			result = append(result, v)
		}
	}
	return result, nil
}

// CallableFunc is a function that can be called.
type CallableFunc = interface{}

// OptionArg represents a binding for an option function argument.
type OptionArg struct {
	r *Callable
	c []gs.Condition
}

// Option creates a binding for an option function argument.
func Option(fn CallableFunc, args ...gs.Arg) *OptionArg {

	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func || t.NumOut() != 1 {
		panic(errors.New("invalid option func"))
	}

	_, file, line, _ := runtime.Caller(1)
	r := MustBind(fn, args...)
	return &OptionArg{r: r.SetFileLine(file, line)}
}

// Condition sets a condition for invoking the option function.
func (arg *OptionArg) Condition(conditions ...gs.Condition) *OptionArg {
	arg.c = append(arg.c, conditions...)
	return arg
}

func (arg *OptionArg) GetArgValue(ctx gs.ArgContext, t reflect.Type) (reflect.Value, error) {

	// Checks if the condition is met.
	for _, c := range arg.c {
		ok, err := ctx.Matches(c)
		if err != nil {
			return reflect.Value{}, err
		} else if !ok {
			return reflect.Value{}, nil
		}
	}

	// Calls the function and returns its result.
	out, err := arg.r.Call(ctx)
	if err != nil {
		return reflect.Value{}, err
	}
	return out[0], nil
}

// Callable wraps a function and its binding arguments.
type Callable struct {
	fn       CallableFunc // The function to be called.
	fnType   reflect.Type // The type of the function.
	argList  *ArgList     // The argument list for the function.
	fileLine string       // File and line number where the function is defined.
}

// MustBind binds arguments to a function and panics if an error occurs.
func MustBind(fn CallableFunc, args ...gs.Arg) *Callable {
	r, err := Bind(fn, args)
	if err != nil {
		panic(err)
	}
	_, file, line, _ := runtime.Caller(1)
	return r.SetFileLine(file, line)
}

// Bind creates a Callable by binding arguments to a function.
func Bind(fn CallableFunc, args []gs.Arg) (*Callable, error) {
	fnType := reflect.TypeOf(fn)
	argList, err := NewArgList(fnType, args)
	if err != nil {
		return nil, err
	}
	return &Callable{fn: fn, fnType: fnType, argList: argList}, nil
}

// SetFileLine sets the file and line number of the function call.
func (r *Callable) SetFileLine(file string, line int) *Callable {
	r.fileLine = fmt.Sprintf("%s:%d", file, line)
	return r
}

// Call invokes the function with its bound arguments processed in the IoC container.
func (r *Callable) Call(ctx gs.ArgContext) ([]reflect.Value, error) {

	in, err := r.argList.get(ctx)
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

func (r *Callable) GetArgValue(ctx gs.ArgContext, t reflect.Type) (reflect.Value, error) {
	if results, err := r.Call(ctx); err != nil {
		return reflect.Value{}, err
	} else if len(results) < 1 {
		return reflect.Value{}, errors.New("xxx")
	} else {
		return results[0], nil
	}
}
