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

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util/assert"
	"go.uber.org/mock/gomock"
)

func TestTagArg(t *testing.T) {

	t.Run("bind success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := NewMockArgContext(ctrl)
		c.EXPECT().Bind(gomock.Any(), gomock.Any()).DoAndReturn(
			func(v reflect.Value, s string) error {
				v.SetString("3")
				return nil
			})
		tag := Tag("${int:=3}")
		v, err := tag.GetArgValue(c, reflect.TypeFor[string]())
		assert.Nil(t, err)
		assert.Equal(t, v.String(), "3")
	})

	t.Run("bind error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := NewMockArgContext(ctrl)
		c.EXPECT().Bind(gomock.Any(), gomock.Any()).DoAndReturn(
			func(v reflect.Value, s string) error {
				return errors.New("bind error")
			})
		tag := Tag("${int:=3}")
		_, err := tag.GetArgValue(c, reflect.TypeFor[string]())
		assert.Error(t, err, "GetArgValue error << bind error")
	})

	t.Run("wire success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := NewMockArgContext(ctrl)
		c.EXPECT().Wire(gomock.Any(), gomock.Any()).DoAndReturn(
			func(v reflect.Value, s string) error {
				v.Set(reflect.ValueOf(&http.Server{Addr: ":9090"}))
				return nil
			})
		tag := Tag("http-server")
		v, err := tag.GetArgValue(c, reflect.TypeFor[*http.Server]())
		assert.Nil(t, err)
		assert.Equal(t, v.Interface().(*http.Server).Addr, ":9090")
	})

	t.Run("wire error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := NewMockArgContext(ctrl)
		c.EXPECT().Wire(gomock.Any(), gomock.Any()).DoAndReturn(
			func(v reflect.Value, s string) error {
				return errors.New("wire error")
			})
		tag := Tag("server")
		_, err := tag.GetArgValue(c, reflect.TypeFor[*bytes.Buffer]())
		assert.Error(t, err, "GetArgValue error << wire error")
	})

	t.Run("type error", func(t *testing.T) {
		tag := Tag("server")
		_, err := tag.GetArgValue(nil, reflect.TypeFor[*string]())
		assert.Error(t, err, "GetArgValue error << unsupported argument type: \\*string")
	})
}

func TestValueArg(t *testing.T) {

	t.Run("index", func(t *testing.T) {
		arg := Index(0, Value(1))
		assert.Equal(t, arg.(IndexArg).Idx, 0)
		assert.Panic(t, func() {
			_, _ = arg.GetArgValue(nil, reflect.TypeFor[int]())
		}, "unimplemented method")
	})

	t.Run("zero", func(t *testing.T) {
		tag := Value(nil)
		v, err := tag.GetArgValue(nil, reflect.TypeFor[*http.Server]())
		assert.Nil(t, err)
		assert.Nil(t, v.Interface())
	})

	t.Run("value", func(t *testing.T) {
		tag := Value(&http.Server{Addr: ":9090"})
		v, err := tag.GetArgValue(nil, reflect.TypeFor[*http.Server]())
		assert.Nil(t, err)
		assert.Equal(t, v.Interface().(*http.Server).Addr, ":9090")
	})

	t.Run("type error", func(t *testing.T) {
		tag := Value(new(int))
		_, err := tag.GetArgValue(nil, reflect.TypeFor[*http.Server]())
		assert.Error(t, err, "GetArgValue error << cannot assign type:\\*int to type:\\*http.Server")
	})
}

func TestArgList_New(t *testing.T) {

	t.Run("invalid function type", func(t *testing.T) {
		fnType := reflect.TypeFor[int]()
		_, err := NewArgList(fnType, nil)
		assert.Error(t, err, "NewArgList error << invalid function type:int")
	})

	t.Run("mixed index and non-index args", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Index(0, Value(1)),
			Value("test"),
		}
		_, err := NewArgList(fnType, args)
		assert.Error(t, err, "NewArgList error << arguments must be all indexed or non-indexed")
	})

	t.Run("mixed non-index and index args", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Value(1),
			Index(1, Value("test")),
		}
		_, err := NewArgList(fnType, args)
		assert.Error(t, err, "NewArgList error << arguments must be all indexed or non-indexed")
	})

	t.Run("invalid argument index - 1", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Index(-1, Value(1)),
		}
		_, err := NewArgList(fnType, args)
		assert.Error(t, err, "NewArgList error << invalid argument index -1")
	})

	t.Run("invalid argument index - 2", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Index(2, Value(1)),
		}
		_, err := NewArgList(fnType, args)
		assert.Error(t, err, "NewArgList error << invalid argument index 2")
	})

	t.Run("non-index args success", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		argList, err := NewArgList(fnType, args)
		assert.Nil(t, err)
		assert.NotNil(t, argList)
		assert.Equal(t, argList.args, []gs.Arg{
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
		assert.Nil(t, err)
		assert.NotNil(t, argList)
		assert.Equal(t, argList.args, []gs.Arg{
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
		assert.Nil(t, err)
		assert.NotNil(t, argList)
		assert.Equal(t, argList.args, []gs.Arg{
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
		assert.Nil(t, err)
		assert.NotNil(t, argList)
		assert.Equal(t, argList.args, []gs.Arg{
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
		assert.Nil(t, err)
		assert.NotNil(t, argList)
		assert.Equal(t, argList.args, []gs.Arg{
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
		assert.Nil(t, err)

		ctx := NewMockArgContext(nil)
		values, err := argList.get(ctx)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(values))
		assert.Equal(t, 1, values[0].Interface().(int))
		assert.Equal(t, "test", values[1].Interface().(string))
	})

	t.Run("success with variadic function", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b ...string) {})
		args := []gs.Arg{
			Value(1),
			Value("test1"),
			Value("test2"),
		}
		argList, err := NewArgList(fnType, args)
		assert.Nil(t, err)

		ctx := NewMockArgContext(nil)
		values, err := argList.get(ctx)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(values))
		assert.Equal(t, 1, values[0].Interface().(int))
		assert.Equal(t, "test1", values[1].Interface().(string))
		assert.Equal(t, "test2", values[2].Interface().(string))
	})

	t.Run("error when getting arg value", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{
			Value(1),
			Value(2),
		}
		argList, err := NewArgList(fnType, args)
		assert.Nil(t, err)

		ctx := NewMockArgContext(nil)
		_, err = argList.get(ctx)
		assert.Error(t, err, "GetArgValue error << cannot assign type:int to type:string")
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
		assert.Error(t, err, "NewArgList error << invalid function type:string")
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
		assert.Nil(t, err)

		ctx := NewMockArgContext(nil)
		_, err = callable.Call(ctx)
		assert.Error(t, err, "GetArgValue error << cannot assign type:int to type:string")
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
		assert.Nil(t, err)

		ctx := NewMockArgContext(nil)
		_, err = callable.Call(ctx)
		assert.Error(t, err, "GetArgValue error << cannot assign type:int to type:string")
	})

	t.Run("function return none", func(t *testing.T) {
		fn := func(a int, b string) {}
		args := []gs.Arg{
			Value(1),
			Value("test"),
		}
		callable, err := NewCallable(fn, args)
		assert.Nil(t, err)

		ctx := NewMockArgContext(nil)
		v, err := callable.Call(ctx)
		assert.Nil(t, err)
		assert.Equal(t, len(v), 0)
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
		assert.Nil(t, err)

		ctx := NewMockArgContext(nil)
		v, err := callable.Call(ctx)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(v))
		assert.Equal(t, "", v[0].Interface().(string))
		assert.Equal(t, "execution error", v[1].Interface().(error).Error())
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
		assert.Nil(t, err)

		ctx := NewMockArgContext(nil)
		v, err := callable.Call(ctx)
		assert.Nil(t, err)
		assert.Equal(t, len(v), 2)
		assert.Equal(t, "1-test", v[0].Interface().(string))
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
		assert.Nil(t, err)

		ctx := NewMockArgContext(nil)
		v, err := callable.Call(ctx)
		assert.Nil(t, err)
		assert.Equal(t, len(v), 1)
		assert.Equal(t, "1-test", v[0].Interface().(string))
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
		assert.Nil(t, err)

		ctx := NewMockArgContext(nil)
		v, err := callable.Call(ctx)
		assert.Nil(t, err)
		assert.Equal(t, len(v), 1)
		assert.Equal(t, "1-test1,test2", v[0].Interface().(string))
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
		assert.Matches(t, arg.fileline, "gs/internal/gs_arg/arg_test.go:495")
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
		assert.Matches(t, arg.fileline, "gs/internal/gs_arg/arg_test.go:507")
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
		ctx := NewMockArgContext(nil)
		_, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.Error(t, err, "GetArgValue error << cannot assign type:int to type:string")
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
		ctx := NewMockArgContext(nil)
		v, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.Nil(t, err)
		assert.Equal(t, "1-test", v.Interface().(string))
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
		ctx := NewMockArgContext(nil)
		_, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.Error(t, err, "execution error")
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
		ctx := NewMockArgContext(nil)
		v, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.Nil(t, err)
		assert.Equal(t, "1-test", v.Interface().(string))
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
		ctx := NewMockArgContext(nil)
		v, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.Nil(t, err)
		assert.Equal(t, "1-test1,test2", v.Interface().(string))
	})

	t.Run("error in condition", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
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
		ctx := NewMockArgContext(ctrl)
		ctx.EXPECT().Check(gomock.Any()).DoAndReturn(
			func(c gs.Condition) (bool, error) {
				ok, err := c.Matches(nil)
				return ok, err
			})
		_, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.Error(t, err, "condition error")
	})

	t.Run("condition return false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
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
		ctx := NewMockArgContext(ctrl)
		ctx.EXPECT().Check(gomock.Any()).DoAndReturn(
			func(c gs.Condition) (bool, error) {
				ok, err := c.Matches(nil)
				return ok, err
			})
		v, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.Nil(t, err)
		assert.False(t, v.IsValid())
	})

	t.Run("condition return true", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
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
		ctx := NewMockArgContext(ctrl)
		ctx.EXPECT().Check(gomock.Any()).DoAndReturn(
			func(c gs.Condition) (bool, error) {
				ok, err := c.Matches(nil)
				return ok, err
			})
		v, err := arg.GetArgValue(ctx, reflect.TypeFor[string]())
		assert.Nil(t, err)
		assert.Equal(t, "1-test", v.Interface().(string))
	})
}
