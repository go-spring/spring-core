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
			return reflect.Value{}, errutil.WrapError(err, "GetArgValue error")
		}
		return v, nil
	}

	// Wires dependent beans based on the argument type.
	if util.IsBeanInjectionTarget(t) {
		v := reflect.New(t).Elem()
		if err := ctx.Wire(v, arg.Tag); err != nil {
			return reflect.Value{}, errutil.WrapError(err, "GetArgValue error")
		}
		return v, nil
	}

	// The arg type must be either a property binding target or a bean injection target.
	err := fmt.Errorf("unsupported argument type: %s", t.String())
	return reflect.Value{}, errutil.WrapError(err, "GetArgValue error")
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

// Value returns a ValueArg with the specified value.
func Value(v interface{}) gs.Arg {
	return ValueArg{v: v}
}

// GetArgValue returns the value of the fixed argument.
func (arg ValueArg) GetArgValue(ctx gs.ArgContext, t reflect.Type) (reflect.Value, error) {
	if arg.v == nil {
		return reflect.Zero(t), nil
	}
	v := reflect.ValueOf(arg.v)
	if !v.Type().AssignableTo(t) {
		err := fmt.Errorf("cannot assign type:%T to type:%s", arg.v, t.String())
		return reflect.Value{}, errutil.WrapError(err, "GetArgValue error")
	}
	return v, nil
}

// ArgList represents a list of arguments for a function.
type ArgList struct {
	fnType reflect.Type // Type of the function to be invoked.
	args   []gs.Arg     // List of arguments for the function.
}

// NewArgList creates and validates an ArgList for the specified function.
func NewArgList(fnType reflect.Type, args []gs.Arg) (*ArgList, error) {
	if fnType.Kind() != reflect.Func {
		err := fmt.Errorf("invalid function type:%s", fnType.String())
		return nil, errutil.WrapError(err, "NewArgList error")
	}

	// Calculates the number of fixed arguments in the function.
	fixedArgCount := fnType.NumIn()
	if fnType.IsVariadic() {
		fixedArgCount--
	}

	fnArgs := make([]gs.Arg, fixedArgCount)
	for i := 0; i < len(fnArgs); i++ {
		fnArgs[i] = Tag("")
	}

	var (
		useIdx bool
		notIdx bool
	)

	for i := 0; i < len(args); i++ {
		switch arg := args[i].(type) {
		case IndexArg:
			useIdx = true
			if notIdx {
				err := fmt.Errorf("all arguments must either have indexes or not have indexes")
				return nil, errutil.WrapError(err, "NewArgList error")
			}
			if arg.Idx < 0 || arg.Idx >= fnType.NumIn() {
				err := fmt.Errorf("invalid argument index %d", arg.Idx)
				return nil, errutil.WrapError(err, "NewArgList error")
			}
			if arg.Idx < fixedArgCount {
				fnArgs[arg.Idx] = arg.Arg
			} else {
				fnArgs = append(fnArgs, arg.Arg)
			}
		default:
			notIdx = true
			if useIdx {
				err := fmt.Errorf("all arguments must either have indexes or not have indexes")
				return nil, errutil.WrapError(err, "NewArgList error")
			}
			if i < fixedArgCount {
				fnArgs[i] = arg
			} else {
				fnArgs = append(fnArgs, arg)
			}
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
			return nil, err
		}
		if v.IsValid() {
			result = append(result, v)
		}
	}
	return result, nil
}

// BindArg represents a binding for an option function argument.
type BindArg struct {
	r        *Callable
	fileLine string
	c        []gs.Condition
}

// Bind creates a binding for an option function argument.
func Bind(fn CallableFunc, args ...gs.Arg) *BindArg {
	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func || t.NumOut() != 1 {
		panic(errors.New("invalid option func"))
	}
	_, file, line, _ := runtime.Caller(1)
	r, err := NewCallable(fn, args)
	if err != nil {
		panic(err)
	}
	arg := &BindArg{r: r}
	arg.SetFileLine(file, line)
	return arg
}

// Condition sets a condition for invoking the option function.
func (arg *BindArg) Condition(conditions ...gs.Condition) *BindArg {
	arg.c = append(arg.c, conditions...)
	return arg
}

// SetFileLine sets the file and line number of the function call.
func (arg *BindArg) SetFileLine(file string, line int) {
	arg.fileLine = fmt.Sprintf("%s:%d", file, line)
}

// GetArgValue retrieves the function's return value if conditions are met.
func (arg *BindArg) GetArgValue(ctx gs.ArgContext, t reflect.Type) (reflect.Value, error) {

	// Checks if the condition is met.
	for _, c := range arg.c {
		ok, err := ctx.Check(c)
		if err != nil {
			return reflect.Value{}, err
		} else if !ok {
			return reflect.Value{}, nil
		}
	}

	// Calls the function and returns its result.
	return arg.r.Call(ctx)
}

// CallableFunc is a function that can be called.
type CallableFunc = interface{}

// Callable wraps a function and its binding arguments.
type Callable struct {
	fn      CallableFunc
	argList *ArgList
}

// NewCallable creates a Callable by binding arguments to a function.
func NewCallable(fn CallableFunc, args []gs.Arg) (*Callable, error) {
	fnType := reflect.TypeOf(fn)
	argList, err := NewArgList(fnType, args)
	if err != nil {
		return nil, err
	}
	return &Callable{fn: fn, argList: argList}, nil
}

// Call invokes the function with its bound arguments processed in the IoC container.
func (r *Callable) Call(ctx gs.ArgContext) (reflect.Value, error) {
	in, err := r.argList.get(ctx)
	if err != nil {
		return reflect.Value{}, err
	}

	out := reflect.ValueOf(r.fn).Call(in)
	n := len(out)
	if n == 0 {
		return reflect.Value{}, nil
	}

	o := out[n-1]
	if util.IsErrorType(o.Type()) {
		if i := o.Interface(); i != nil {
			return out[0], i.(error)
		}
		return out[0], nil
	}
	return out[0], nil
}
