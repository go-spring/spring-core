/*
 * Copyright 2025 The Go-Spring Authors.
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
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/go-spring/gs-assert/assert"
	"github.com/go-spring/gs-mock/gsmock"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
)

func TestTagArg(t *testing.T) {

	t.Run("bind success", func(t *testing.T) {
		m := gsmock.NewManager()
		c := gs.NewArgContextMockImpl(m)
		c.MockBind().Handle(func(v reflect.Value, s string) error {
			v.SetString("3")
			return nil
		})

		tag := Tag("${int:=3}")
		v, err := tag.GetArgValue(c, reflect.TypeFor[string]())
		assert.That(t, err).Nil()
		assert.That(t, v.String()).Equal("3")
	})

	t.Run("bind error", func(t *testing.T) {
		m := gsmock.NewManager()
		c := gs.NewArgContextMockImpl(m)
		c.MockBind().Handle(func(v reflect.Value, s string) error {
			return errors.New("bind error")
		})

		tag := Tag("${int:=3}")
		_, err := tag.GetArgValue(c, reflect.TypeFor[string]())
		assert.ThatError(t, err).Matches("GetArgValue error << bind error")
	})

	t.Run("wire success", func(t *testing.T) {
		m := gsmock.NewManager()
		c := gs.NewArgContextMockImpl(m)
		c.MockWire().Handle(func(v reflect.Value, s string) error {
			v.Set(reflect.ValueOf(&http.Server{Addr: ":9090"}))
			return nil
		})

		tag := Tag("http-server")
		v, err := tag.GetArgValue(c, reflect.TypeFor[*http.Server]())
		assert.That(t, err).Nil()
		assert.That(t, v.Interface().(*http.Server).Addr).Equal(":9090")
	})

	t.Run("wire error", func(t *testing.T) {
		m := gsmock.NewManager()
		c := gs.NewArgContextMockImpl(m)
		c.MockWire().Handle(func(v reflect.Value, s string) error {
			return errors.New("wire error")
		})

		tag := Tag("server")
		_, err := tag.GetArgValue(c, reflect.TypeFor[*bytes.Buffer]())
		assert.ThatError(t, err).Matches("GetArgValue error << wire error")
	})

	t.Run("type error", func(t *testing.T) {
		tag := Tag("server")
		_, err := tag.GetArgValue(nil, reflect.TypeFor[*string]())
		assert.ThatError(t, err).Matches("GetArgValue error << unsupported argument type: \\*string")
	})
}

func TestValueArg(t *testing.T) {

	t.Run("index", func(t *testing.T) {
		arg := Index(0, Value(1))
		assert.That(t, arg.(IndexArg).Idx).Equal(0)
		assert.Panic(t, func() {
			_, _ = arg.GetArgValue(nil, reflect.TypeFor[int]())
		}, "unimplemented method")
	})

	t.Run("zero", func(t *testing.T) {
		tag := Value(nil)
		v, err := tag.GetArgValue(nil, reflect.TypeFor[*http.Server]())
		assert.That(t, err).Nil()
		assert.That(t, v.Interface())
	})

	t.Run("value", func(t *testing.T) {
		tag := Value(&http.Server{Addr: ":9090"})
		v, err := tag.GetArgValue(nil, reflect.TypeFor[*http.Server]())
		assert.That(t, err).Nil()
		assert.That(t, v.Interface().(*http.Server).Addr).Equal(":9090")
	})

	t.Run("type error", func(t *testing.T) {
		tag := Value(new(int))
		_, err := tag.GetArgValue(nil, reflect.TypeFor[*http.Server]())
		assert.ThatError(t, err).Matches("GetArgValue error << cannot assign type:\\*int to type:\\*http.Server")
	})
}

func TestArgList_New(t *testing.T) {

	t.Run("invalid function type", func(t *testing.T) {
		fnType := reflect.TypeFor[int]()
		_, err := NewArgList(fnType, nil)
		assert.ThatError(t, err).Matches("NewArgList error << invalid function type:int")
	})

	t.Run("mixed index and non-index args", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Index(0, Value(1)),
			Value("test"),
		}
		_, err := NewArgList(fnType, args)
		assert.ThatError(t, err).Matches("NewArgList error << arguments must be all indexed or non-indexed")
	})

	t.Run("mixed non-index and index args", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Value(1),
			Index(1, Value("test")),
		}
		_, err := NewArgList(fnType, args)
		assert.ThatError(t, err).Matches("NewArgList error << arguments must be all indexed or non-indexed")
	})

	t.Run("invalid argument index - 1", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Index(-1, Value(1)),
		}
		_, err := NewArgList(fnType, args)
		assert.ThatError(t, err).Matches("NewArgList error << invalid argument index -1")
	})

	t.Run("invalid argument index - 2", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Index(2, Value(1)),
		}
		_, err := NewArgList(fnType, args)
		assert.ThatError(t, err).Matches("NewArgList error << invalid argument index 2")
	})

	t.Run("non-index args success", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		argList, err := NewArgList(fnType, args)
		assert.That(t, err).Nil()
		assert.That(t, argList).NotNil()
		assert.That(t, argList.args).Equal([]gs.Arg{
			Value(1),
			Value("test"),
		})
	})

	t.Run("index args success", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Index(0, Value(1)),
			Index(1, Value("test")),
		}
		argList, err := NewArgList(fnType, args)
		assert.That(t, err).Nil()
		assert.That(t, argList).NotNil()
		assert.That(t, argList.args).Equal([]gs.Arg{
			Value(1),
			Value("test"),
		})
	})

	t.Run("variadic function with non-index args success", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b ...string) {})
		args := []gs.Arg{
			Value(1),
			Value("test1"),
			Value("test2"),
		}
		argList, err := NewArgList(fnType, args)
		assert.That(t, err).Nil()
		assert.That(t, argList).NotNil()
		assert.That(t, argList.args).Equal([]gs.Arg{
			Value(1),
			Value("test1"),
			Value("test2"),
		})
	})

	t.Run("variadic function with index args success - 1", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b ...string) {})
		args := []gs.Arg{
			Index(0, Value(1)),
			Index(1, Value("test1")),
			Index(1, Value("test2")),
		}
		argList, err := NewArgList(fnType, args)
		assert.That(t, err).Nil()
		assert.That(t, argList).NotNil()
		assert.That(t, argList.args).Equal([]gs.Arg{
			Value(1),
			Value("test1"),
			Value("test2"),
		})
	})

	t.Run("variadic function with index args success - 2", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a error, b ...string) {})
		args := []gs.Arg{
			Index(1, Value("test1")),
			Index(1, Value("test2")),
		}
		argList, err := NewArgList(fnType, args)
		assert.That(t, err).Nil()
		assert.That(t, argList).NotNil()
		assert.That(t, argList.args).Equal([]gs.Arg{
			Tag(""),
			Value("test1"),
			Value("test2"),
		})
	})
}

func TestArgList_Get(t *testing.T) {

	t.Run("success with non-variadic function", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		argList, err := NewArgList(fnType, args)
		assert.That(t, err).Nil()

		ctx := gs.NewArgContextMockImpl(nil)
		values, err := argList.get(ctx)
		assert.That(t, err).Nil()
		assert.That(t, 2).Equal(len(values))
		assert.That(t, 1).Equal(values[0].Interface().(int))
		assert.That(t, "test").Equal(values[1].Interface().(string))
	})

	t.Run("success with variadic function", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b ...string) {})
		args := []gs.Arg{
			Value(1),
			Value("test1"),
			Value("test2"),
		}
		argList, err := NewArgList(fnType, args)
		assert.That(t, err).Nil()

		ctx := gs.NewArgContextMockImpl(nil)
		values, err := argList.get(ctx)
		assert.That(t, err).Nil()
		assert.That(t, 3).Equal(len(values))
		assert.That(t, 1).Equal(values[0].Interface().(int))
		assert.That(t, "test1").Equal(values[1].Interface().(string))
		assert.That(t, "test2").Equal(values[2].Interface().(string))
	})

	t.Run("error when getting arg value", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Value(1),
			Value(2),
		}
		argList, err := NewArgList(fnType, args)
		assert.That(t, err).Nil()

		ctx := gs.NewArgContextMockImpl(nil)
		_, err = argList.get(ctx)
		assert.ThatError(t, err).Matches("GetArgValue error << cannot assign type:int to type:string")
	})
}

func TestCallable_New(t *testing.T) {

	t.Run("invalid function type", func(t *testing.T) {
		fn := "not a function"
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		_, err := NewCallable(fn, args)
		assert.ThatError(t, err).Matches("NewArgList error << invalid function type:string")
	})

	t.Run("error in argument processing", func(t *testing.T) {
		fn := func(a int, b string) string {
			return fmt.Sprintf("%d-%s", a, b)
		}
		args := []gs.Arg{
			Value(1),
			Value(2),
		}
		callable, err := NewCallable(fn, args)
		assert.That(t, err).Nil()

		ctx := gs.NewArgContextMockImpl(nil)
		_, err = callable.Call(ctx)
		assert.ThatError(t, err).Matches("GetArgValue error << cannot assign type:int to type:string")
	})
}

func TestCallable_Call(t *testing.T) {

	t.Run("error in get arg value", func(t *testing.T) {
		fn := func(a int, b string) (string, error) {
			return "", nil
		}
		args := []gs.Arg{
			Value(1),
			Value(2),
		}
		callable, err := NewCallable(fn, args)
		assert.That(t, err).Nil()

		ctx := gs.NewArgContextMockImpl(nil)
		_, err = callable.Call(ctx)
		assert.ThatError(t, err).Matches("GetArgValue error << cannot assign type:int to type:string")
	})

	t.Run("function return none", func(t *testing.T) {
		fn := func(a int, b string) {}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		callable, err := NewCallable(fn, args)
		assert.That(t, err).Nil()

		ctx := gs.NewArgContextMockImpl(nil)
		v, err := callable.Call(ctx)
		assert.That(t, err).Nil()
		assert.That(t, len(v)).Equal(0)
	})

	t.Run("function return error", func(t *testing.T) {
		fn := func(a int, b string) (string, error) {
			return "", errors.New("execution error")
		}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		callable, err := NewCallable(fn, args)
		assert.That(t, err).Nil()

		ctx := gs.NewArgContextMockImpl(nil)
		v, err := callable.Call(ctx)
		assert.That(t, err).Nil()
		assert.That(t, 2).Equal(len(v))
		assert.That(t, "").Equal(v[0].Interface().(string))
		assert.That(t, "execution error").Equal(v[1].Interface().(error).Error())
	})

	t.Run("success - 1", func(t *testing.T) {
		fn := func(a int, b string) (string, error) {
			return fmt.Sprintf("%d-%s", a, b), nil
		}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		callable, err := NewCallable(fn, args)
		assert.That(t, err).Nil()

		ctx := gs.NewArgContextMockImpl(nil)
		v, err := callable.Call(ctx)
		assert.That(t, err).Nil()
		assert.That(t, len(v)).Equal(2)
		assert.That(t, "1-test").Equal(v[0].Interface().(string))
	})

	t.Run("success - 2", func(t *testing.T) {
		fn := func(a int, b string) string {
			return fmt.Sprintf("%d-%s", a, b)
		}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		callable, err := NewCallable(fn, args)
		assert.That(t, err).Nil()

		ctx := gs.NewArgContextMockImpl(nil)
		v, err := callable.Call(ctx)
		assert.That(t, err).Nil()
		assert.That(t, len(v)).Equal(1)
		assert.That(t, "1-test").Equal(v[0].Interface().(string))
	})

	t.Run("success with variadic function", func(t *testing.T) {
		fn := func(a int, b ...string) string {
			return fmt.Sprintf("%d-%s", a, strings.Join(b, ","))
		}
		args := []gs.Arg{
			Value(1),
			Value("test1"),
			Value("test2"),
		}
		callable, err := NewCallable(fn, args)
		assert.That(t, err).Nil()

		ctx := gs.NewArgContextMockImpl(nil)
		v, err := callable.Call(ctx)
		assert.That(t, err).Nil()
		assert.That(t, len(v)).Equal(1)
		assert.That(t, "1-test1,test2").Equal(v[0].Interface().(string))
	})
}

func TestBindArg_Bind(t *testing.T) {

	t.Run("invalid function type - 1", func(t *testing.T) {
		fn := "not a function"
		assert.Panic(t, func() {
			Bind(fn)
		}, "invalid function type")
	})

	t.Run("invalid function type - 2", func(t *testing.T) {
		fn := func(a int, b string) error {
			return nil
		}
		assert.Panic(t, func() {
			Bind(fn)
		}, "invalid function type")
	})

	t.Run("invalid function type - 3", func(t *testing.T) {
		fn := func(a int, b string) (string, bool) {
			return fmt.Sprintf("%d-%s", a, b), true
		}
		assert.Panic(t, func() {
			Bind(fn)
		}, "invalid function type")
	})

	t.Run("error in argument processing", func(t *testing.T) {
		fn := func(a int, b string) string {
			return fmt.Sprintf("%d-%s", a, b)
		}
		args := []gs.Arg{
			Value(1),
			Index(1, Value("test")),
		}
		assert.Panic(t, func() {
			Bind(fn, args...)
		}, "NewArgList error << arguments must be all indexed or non-indexed")
	})

	t.Run("success - 1", func(t *testing.T) {
		fn := func(a int, b string) string {
			return fmt.Sprintf("%d-%s", a, b)
		}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		arg := Bind(fn, args...)
		assert.ThatString(t, arg.fileline).Matches("gs/internal/gs_arg/arg_test.go:491")
	})

	t.Run("success - 2", func(t *testing.T) {
		fn := func(a int, b string) (string, error) {
			return fmt.Sprintf("%d-%s", a, b), nil
		}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		arg := Bind(fn, args...)
		assert.ThatString(t, arg.fileline).Matches("gs/internal/gs_arg/arg_test.go:503")
	})
}

func TestBindArg_GetArgValue(t *testing.T) {

	t.Run("error in get arg value", func(t *testing.T) {
		fn := func(a int, b string) string {
			return fmt.Sprintf("%d-%s", a, b)
		}
		args := []gs.Arg{
			Value(1),
			Value(2),
		}
		arg := Bind(fn, args...)
		ctx := gs.NewArgContextMockImpl(nil)
		_, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.ThatError(t, err).Matches("GetArgValue error << cannot assign type:int to type:string")
	})

	t.Run("success", func(t *testing.T) {
		fn := func(a int, b string) string {
			return fmt.Sprintf("%d-%s", a, b)
		}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		arg := Bind(fn, args...)
		ctx := gs.NewArgContextMockImpl(nil)
		v, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.That(t, err).Nil()
		assert.That(t, "1-test").Equal(v.Interface().(string))
	})

	t.Run("error in function execution", func(t *testing.T) {
		fn := func(a int, b string) (string, error) {
			return "", errors.New("execution error")
		}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		arg := Bind(fn, args...)
		ctx := gs.NewArgContextMockImpl(nil)
		_, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.ThatError(t, err).Matches("execution error")
	})

	t.Run("no error in function execution", func(t *testing.T) {
		fn := func(a int, b string) (string, error) {
			return fmt.Sprintf("%d-%s", a, b), nil
		}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		arg := Bind(fn, args...)
		ctx := gs.NewArgContextMockImpl(nil)
		v, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.That(t, err).Nil()
		assert.That(t, "1-test").Equal(v.Interface().(string))
	})

	t.Run("success with variadic function", func(t *testing.T) {
		fn := func(a int, b ...string) string {
			return fmt.Sprintf("%d-%s", a, strings.Join(b, ","))
		}
		args := []gs.Arg{
			Value(1),
			Value("test1"),
			Value("test2"),
		}
		arg := Bind(fn, args...)
		ctx := gs.NewArgContextMockImpl(nil)
		v, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.That(t, err).Nil()
		assert.That(t, "1-test1,test2").Equal(v.Interface().(string))
	})

	t.Run("error in condition", func(t *testing.T) {
		fn := func(a int, b string) string {
			return fmt.Sprintf("%d-%s", a, b)
		}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		arg := Bind(fn, args...)
		arg.Condition(gs_cond.OnFunc(func(ctx gs.CondContext) (bool, error) {
			return false, errors.New("condition error")
		}))

		m := gsmock.NewManager()
		c := gs.NewArgContextMockImpl(m)
		c.MockCheck().Handle(func(c gs.Condition) (bool, error) {
			ok, err := c.Matches(nil)
			return ok, err
		})

		_, err := arg.GetArgValue(c, reflect.TypeFor[string]())
		assert.ThatError(t, err).Matches("condition error")
	})

	t.Run("condition return false", func(t *testing.T) {
		fn := func(a int, b string) string {
			return fmt.Sprintf("%d-%s", a, b)
		}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		arg := Bind(fn, args...)
		arg.Condition(gs_cond.OnFunc(func(ctx gs.CondContext) (bool, error) {
			return false, nil
		}))

		m := gsmock.NewManager()
		c := gs.NewArgContextMockImpl(m)
		c.MockCheck().Handle(func(c gs.Condition) (bool, error) {
			ok, err := c.Matches(nil)
			return ok, err
		})

		v, err := arg.GetArgValue(c, reflect.TypeFor[string]())
		assert.That(t, err).Nil()
		assert.That(t, v.IsValid()).False()
	})

	t.Run("condition return true", func(t *testing.T) {
		fn := func(a int, b string) string {
			return fmt.Sprintf("%d-%s", a, b)
		}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		arg := Bind(fn, args...)
		arg.Condition(gs_cond.OnFunc(func(ctx gs.CondContext) (bool, error) {
			return true, nil
		}))

		m := gsmock.NewManager()
		c := gs.NewArgContextMockImpl(m)
		c.MockCheck().Handle(func(c gs.Condition) (bool, error) {
			ok, err := c.Matches(nil)
			return ok, err
		})

		v, err := arg.GetArgValue(c, reflect.TypeFor[string]())
		assert.That(t, err).Nil()
		assert.That(t, "1-test").Equal(v.Interface().(string))
	})
}
