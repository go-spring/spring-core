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

package gs

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_app"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
	"github.com/go-spring/spring-core/gs/internal/gs_init"
	"github.com/go-spring/stdlib/flatten"
)

const (
	Version = "go-spring@v1.2.5"
	Website = "https://github.com/go-spring/"
)

// BeanID represents a selector for a bean.
type BeanID = gs.BeanID

// Dync is a generic alias for a dynamic configuration value.
// It represents a property that can change at runtime.
type Dync[T any] = gs_dync.Value[T]

// As returns the [reflect.Type] of an interface T.
// T is expected to be an interface type.
func As[T any]() reflect.Type {
	return gs.As[T]()
}

/************************************ arg ***********************************/

// Arg represents an argument used when binding constructor parameters.
type Arg = gs.Arg

// TagArg creates an argument that injects a property or bean
// identified by the specified struct-tag expression.
func TagArg(tag string) Arg {
	return gs_arg.Tag(tag)
}

// ValueArg creates an argument with a fixed value.
func ValueArg(v any) Arg {
	return gs_arg.Value(v)
}

// IndexArg targets a specific constructor parameter by index
// and provides the given Arg as its value.
func IndexArg(n int, arg Arg) Arg {
	return gs_arg.Index(n, arg)
}

// BindArg binds arguments dynamically to an option-style constructor.
func BindArg(fn any, args ...Arg) *gs_arg.BindArg {
	return gs_arg.Bind(fn, args...)
}

/************************************ cond ***********************************/

type (
	// Condition represents a logical predicate that decides whether
	// a bean or module should be activated.
	Condition = gs.Condition

	// ConditionContext provides the evaluation context for a Condition.
	ConditionContext = gs.ConditionContext

	// PropertyCondition is a convenience wrapper for property-based conditions.
	PropertyCondition = gs_cond.PropertyCondition
)

// OnOnce wraps the given conditions so they are evaluated only once.
// Subsequent calls return the same result. (Not concurrency-safe.)
func OnOnce(conditions ...Condition) Condition {
	var (
		done   bool
		result bool
	)
	return OnFunc(func(ctx ConditionContext) (_ bool, err error) {
		if done {
			return result, nil
		}
		done = true
		result, err = gs_cond.And(conditions...).Matches(ctx)
		return result, err
	})
}

// OnFunc creates a Condition backed by the given function.
func OnFunc(fn func(ctx ConditionContext) (bool, error)) Condition {
	return gs_cond.OnFunc(fn)
}

// OnProperty creates a property-based condition.
func OnProperty(name string) PropertyCondition {
	return gs_cond.OnProperty(name)
}

// OnBean requires that a bean of the given type (and optional name) exists.
func OnBean[T any](name ...string) Condition {
	return gs_cond.OnBean[T](name...)
}

// OnMissingBean requires that no bean of the given type (and optional name) exists.
func OnMissingBean[T any](name ...string) Condition {
	return gs_cond.OnMissingBean[T](name...)
}

// OnSingleBean requires that exactly one instance of the given bean type exists.
func OnSingleBean[T any](name ...string) Condition {
	return gs_cond.OnSingleBean[T](name...)
}

// RegisterExpressFunc registers a custom expression function
// that can be used inside conditional expressions.
// It should be called during application initialization.
func RegisterExpressFunc(name string, fn any) {
	gs_cond.RegisterExpressFunc(name, fn)
}

// OnExpression creates a condition from an expression.
func OnExpression(expression string) Condition {
	return gs_cond.OnExpression(expression)
}

// Not returns the logical negation of the given condition.
func Not(c Condition) Condition {
	return gs_cond.Not(c)
}

// Or combines multiple conditions using logical OR.
func Or(conditions ...Condition) Condition {
	return gs_cond.Or(conditions...)
}

// And combines multiple conditions using logical AND.
func And(conditions ...Condition) Condition {
	return gs_cond.And(conditions...)
}

// None returns a condition that is true if all provided conditions are false.
func None(conditions ...Condition) Condition {
	return gs_cond.None(conditions...)
}

/*********************************** app *************************************/

type (
	BeanProvider    = gs_init.BeanProvider
	Runner          = gs_app.Runner
	Server          = gs_app.Server
	ReadySignal     = gs_app.ReadySignal
	ContextAware    = gs_app.ContextAware
	ConfigRefresher = gs_app.ConfigRefresher
)

// Provide registers a global bean definition.
// It must be called during package initialization (init phase).
// Calling it after application configuration has started will panic.
// It accepts either an existing instance or a constructor function.
// The optional args are used to bind parameters for the constructor or to
// provide explicit injection values.
func Provide(objOrCtor any, args ...Arg) *gs_bean.BeanDefinition {
	if inited {
		panic("gs.Provide can only be called in init function")
	}
	b := gs_bean.NewBean(objOrCtor, args...)
	gs_init.AddBean(b)
	return b.Caller(2)
}

// ModuleFunc defines the signature of a module function.
type ModuleFunc = gs_init.ModuleFunc

// Module registers a configuration module that is conditionally activated
// based on property values.
func Module(c PropertyCondition, fn ModuleFunc) {
	if inited {
		panic("gs.Module can only be called in init function")
	}
	_, file, line, _ := runtime.Caller(1)
	gs_init.AddModule(c, fn, file, line)
}

// Group registers a set of beans based on a configuration property map.
// Each map entry spawns a bean constructed via fn and optionally destroyed via d.
// The bean name is derived from the map key.
func Group[T any, R any](tag string, fn func(c T) (R, error), d func(R) error) {
	if inited {
		panic("gs.Group can only be called in init function")
	}
	if !strings.HasPrefix(tag, "${") || !strings.HasSuffix(tag, "}") {
		panic("gs.Group tag must be in ${...} format")
	}
	_, file, line, _ := runtime.Caller(1)
	key := strings.TrimSuffix(strings.TrimPrefix(tag, "${"), "}")
	gs_init.AddModule(OnProperty(key), func(r BeanProvider, p flatten.Storage) error {
		var m map[string]T
		if err := conf.Bind(p, &m, "${"+key+"}"); err != nil {
			return err
		}
		for name, c := range m {
			b := r.Provide(fn, ValueArg(c)).Name(name)
			if d != nil {
				b.Destroy(d)
			}
			b.SetFileLine(file, line)
		}
		return nil
	}, file, line)
}
