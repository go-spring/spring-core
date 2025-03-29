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

// TagArg represents an argument resolved using a tag for property binding or dependency injection.
type TagArg struct {
	Tag string
}

// Tag creates a TagArg with the given tag.
func Tag(tag string) gs.Arg {
	return TagArg{Tag: tag}
}

// GetArgValue resolves the tag to a value based on the target type.
// For primitive types (int, string), it binds from configuration.
// For structs/interfaces, it wires dependencies from the container.
// It returns an error if the type is neither bindable nor injectable.
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

	err := fmt.Errorf("unsupported argument type: %s", t.String())
	return reflect.Value{}, errutil.WrapError(err, "GetArgValue error")
}

// IndexArg represents an argument with an explicit positional index in the function signature.
type IndexArg struct {
	Idx int    //The positional index (0-based).
	Arg gs.Arg //The wrapped argument value.
}

// Index creates an IndexArg with the given index and argument.
func Index(n int, arg gs.Arg) gs.Arg {
	return IndexArg{Idx: n, Arg: arg}
}

// GetArgValue panics if called directly. IndexArg must be processed by ArgList.
func (arg IndexArg) GetArgValue(ctx gs.ArgContext, t reflect.Type) (reflect.Value, error) {
	panic(util.UnimplementedMethod)
}

// ValueArg represents a fixed-value argument.
type ValueArg struct {
	v interface{}
}

// Value creates a fixed-value argument.
func Value(v interface{}) gs.Arg {
	return ValueArg{v: v}
}

// GetArgValue returns the fixed value and validates type compatibility.
// It returns an error if the value type is incompatible with the target type.
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

// ArgList manages arguments for a function, supporting both fixed and variadic parameters.
type ArgList struct {
	fnType reflect.Type // The reflected type of the target function.
	args   []gs.Arg     // The argument list (indexed or non-indexed).
}

// NewArgList validates and creates an ArgList for a function. It returns errors
// for invalid function types, mixed indexed/non-indexed args, or out-of-bounds indexes.
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
	for i := range fnArgs {
		fnArgs[i] = Tag("")
	}

	var (
		useIdx bool
		notIdx bool
	)

	for i := range args {
		switch arg := args[i].(type) {
		case IndexArg:
			useIdx = true
			if notIdx {
				err := errors.New("arguments must be all indexed or non-indexed")
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
				err := errors.New("arguments must be all indexed or non-indexed")
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

// get resolves all arguments and returns their reflected values.
func (r *ArgList) get(ctx gs.ArgContext) ([]reflect.Value, error) {

	fnType := r.fnType
	numIn := fnType.NumIn()
	variadic := fnType.IsVariadic()
	result := make([]reflect.Value, 0, len(r.args))

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

// CallableFunc is a function that can be called.
type CallableFunc = interface{}

// Callable wraps a function and its bound arguments for invocation.
type Callable struct {
	fn      CallableFunc
	argList *ArgList
}

// NewCallable binds arguments to a function and creates a Callable. It
// returns errors for invalid function types or argument validation failures.
func NewCallable(fn CallableFunc, args []gs.Arg) (*Callable, error) {
	fnType := reflect.TypeOf(fn)
	argList, err := NewArgList(fnType, args)
	if err != nil {
		return nil, err
	}
	return &Callable{fn: fn, argList: argList}, nil
}

// Call invokes the function with resolved arguments.
func (r *Callable) Call(ctx gs.ArgContext) ([]reflect.Value, error) {
	ret, err := r.argList.get(ctx)
	if err != nil {
		return nil, err
	}
	return reflect.ValueOf(r.fn).Call(ret), nil
}

// BindArg represents a bound function with conditions for conditional execution.
type BindArg struct {
	r          *Callable      // The wrapped Callable.
	fileline   string         // Source location of the Bind call (for debugging).
	conditions []gs.Condition // Conditions that must be met to execute the function.
}

// validBindFunc validates if a function is a valid binding target.
// Valid signatures:
//   - func(...) error
//   - func(...) (T, error)
func validBindFunc(fn CallableFunc) error {
	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func {
		return errors.New("invalid function type")
	}
	if numOut := t.NumOut(); numOut == 1 {
		if o := t.Out(0); !util.IsErrorType(o) {
			return nil
		}
	} else if numOut == 2 {
		if o := t.Out(t.NumOut() - 1); util.IsErrorType(o) {
			return nil
		}
	}
	return errors.New("invalid function type")
}

// Bind creates a binding for an option function. It panics on validation errors.
// `fn` is The target function (must return error or (T, error)). `args` is the
// bound arguments (indexed or non-indexed).
func Bind(fn CallableFunc, args ...gs.Arg) *BindArg {
	if err := validBindFunc(fn); err != nil {
		panic(err)
	}
	r, err := NewCallable(fn, args)
	if err != nil {
		panic(err)
	}
	arg := &BindArg{r: r}
	_, file, line, _ := runtime.Caller(1)
	arg.SetFileLine(file, line)
	return arg
}

// SetFileLine sets the source location of the Bind call.
func (arg *BindArg) SetFileLine(file string, line int) {
	arg.fileline = fmt.Sprintf("%s:%d", file, line)
}

// Condition adds pre-execution conditions to the binding.
func (arg *BindArg) Condition(c ...gs.Condition) *BindArg {
	arg.conditions = append(arg.conditions, c...)
	return arg
}

// GetArgValue executes the function if all conditions are met and returns the result.
// It returns an invalid [reflect.Value] if conditions are not met. It also propagates
// errors from the function or condition checks.
func (arg *BindArg) GetArgValue(ctx gs.ArgContext, t reflect.Type) (reflect.Value, error) {

	// Checks if the condition is met.
	for _, c := range arg.conditions {
		ok, err := ctx.Check(c)
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
	if len(out) == 1 {
		return out[0], nil
	}
	err, _ = out[1].Interface().(error)
	return out[0], err
}
