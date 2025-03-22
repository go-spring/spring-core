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

package gs_arg_test

import (
	"bytes"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-spring/spring-core/gs/gsmock"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/util/assert"
)

func TestTagArg(t *testing.T) {

	t.Run("bind success", func(t *testing.T) {
		r, _ := gsmock.Init(t.Context())
		c := gs.NewMockArgContext(r)
		c.MockBind().Handle(func(v reflect.Value, s string) (error, bool) {
			v.SetString("3")
			return nil, true
		})
		tag := gs_arg.Tag("${int:=3}")
		v, err := tag.GetArgValue(c, reflect.TypeFor[string]())
		assert.Nil(t, err)
		assert.Equal(t, v.String(), "3")
	})

	t.Run("bind error", func(t *testing.T) {
		r, _ := gsmock.Init(t.Context())
		c := gs.NewMockArgContext(r)
		c.MockBind().Handle(func(v reflect.Value, s string) (error, bool) {
			return errors.New("bind error"), true
		})
		tag := gs_arg.Tag("${int:=3}")
		_, err := tag.GetArgValue(c, reflect.TypeFor[string]())
		assert.Error(t, err, "bind error")
	})

	t.Run("wire success", func(t *testing.T) {
		r, _ := gsmock.Init(t.Context())
		c := gs.NewMockArgContext(r)
		c.MockWire().Handle(func(v reflect.Value, s string) (error, bool) {
			v.Set(reflect.ValueOf(&http.Server{Addr: ":9090"}))
			return nil, true
		})
		tag := gs_arg.Tag("http-server")
		v, err := tag.GetArgValue(c, reflect.TypeFor[*http.Server]())
		assert.Nil(t, err)
		assert.Equal(t, v.Interface().(*http.Server).Addr, ":9090")
	})

	t.Run("wire error", func(t *testing.T) {
		r, _ := gsmock.Init(t.Context())
		c := gs.NewMockArgContext(r)
		c.MockWire().Handle(func(v reflect.Value, s string) (error, bool) {
			return errors.New("wire error"), true
		})
		tag := gs_arg.Tag("server")
		_, err := tag.GetArgValue(c, reflect.TypeFor[*bytes.Buffer]())
		assert.Error(t, err, "wire error")
	})

	t.Run("type error", func(t *testing.T) {
		tag := gs_arg.Tag("server")
		_, err := tag.GetArgValue(nil, reflect.TypeFor[*string]())
		assert.Error(t, err, "GetArgValue error << unsupported argument type: \\*string")
	})
}

func TestValueArg(t *testing.T) {

	t.Run("zero", func(t *testing.T) {
		tag := gs_arg.Zero()
		v, err := tag.GetArgValue(nil, reflect.TypeFor[*http.Server]())
		assert.Nil(t, err)
		assert.Nil(t, v.Interface())
	})

	t.Run("value", func(t *testing.T) {
		tag := gs_arg.Value(&http.Server{Addr: ":9090"})
		v, err := tag.GetArgValue(nil, reflect.TypeFor[*http.Server]())
		assert.Nil(t, err)
		assert.Equal(t, v.Interface().(*http.Server).Addr, ":9090")
	})

	t.Run("type error", func(t *testing.T) {
		tag := gs_arg.Value(new(int))
		_, err := tag.GetArgValue(nil, reflect.TypeFor[*http.Server]())
		assert.Error(t, err, "GetArgValue error << cannot assign type:\\*int to type:\\*http.Server")
	})
}

func TestArgList(t *testing.T) {

	t.Run("invalid function type", func(t *testing.T) {
		_, err := gs_arg.NewArgList(reflect.TypeFor[int](), nil)
		assert.Error(t, err, "NewArgList error << invalid function type:int")
	})

	t.Run("mixed index and non-index args", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{gs_arg.Index(0, gs_arg.Value(1)), gs_arg.Value("test")}
		_, err := gs_arg.NewArgList(fnType, args)
		assert.Error(t, err, "NewArgList error << all arguments must either have indexes or not have indexes")
	})

	t.Run("mixed non-index and index args", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{gs_arg.Value(1), gs_arg.Index(1, gs_arg.Value("test"))}
		_, err := gs_arg.NewArgList(fnType, args)
		assert.Error(t, err, "NewArgList error << all arguments must either have indexes or not have indexes")
	})

	t.Run("invalid argument index -1", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{gs_arg.Index(-1, gs_arg.Value(1))}
		_, err := gs_arg.NewArgList(fnType, args)
		assert.Error(t, err, "NewArgList error << invalid argument index -1")
	})

	t.Run("invalid argument index 2", func(t *testing.T) {
		fnType := reflect.TypeOf(func(a int, b string) {})
		args := []gs.Arg{gs_arg.Index(2, gs_arg.Value(1))}
		_, err := gs_arg.NewArgList(fnType, args)
		assert.Error(t, err, "NewArgList error << invalid argument index 2")
	})
}
