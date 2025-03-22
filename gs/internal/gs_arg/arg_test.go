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
}
