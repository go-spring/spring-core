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

package gs_dync

import (
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/util/assert"
)

type MockPanicRefreshable struct{}

func (m *MockPanicRefreshable) onRefresh(prop conf.Properties, param conf.BindParam) error {
	panic("mock panic")
}

type MockErrorRefreshable struct{}

func (m *MockErrorRefreshable) onRefresh(prop conf.Properties, param conf.BindParam) error {
	return errors.New("mock error")
}

func TestValue(t *testing.T) {
	var v Value[int]
	assert.Equal(t, v.Value(), 0)

	refresh := func(prop conf.Properties) error {
		return v.onRefresh(prop, conf.BindParam{Key: "key"})
	}

	err := refresh(conf.Map(map[string]interface{}{
		"key": "42",
	}))
	assert.Nil(t, err)
	assert.Equal(t, v.Value(), 42)

	err = refresh(conf.Map(map[string]interface{}{
		"key": map[string]interface{}{
			"value": "42",
		},
	}))
	assert.Error(t, err, "bind path= type=int error << property key isn't simple value")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-v.NewListener().C
			assert.Equal(t, v.Value(), 59)
		}()
	}

	time.Sleep(50 * time.Millisecond)
	err = refresh(conf.Map(map[string]interface{}{
		"key": 59,
	}))
	assert.Nil(t, err)

	wg.Wait()

	b, err := json.Marshal(map[string]interface{}{"key": &v})
	assert.Nil(t, err)
	assert.JsonEqual(t, string(b), `{"key":59}`)
}

func TestDync(t *testing.T) {

	t.Run("refresh field", func(t *testing.T) {
		p := New(conf.New())
		assert.Equal(t, p.ObjectsCount(), 0)

		prop := conf.Map(map[string]interface{}{
			"config.sub.value": "99",
		})
		err := p.Refresh(prop)
		assert.Nil(t, err)
		assert.Equal(t, p.Data(), prop)

		var v int
		err = p.RefreshField(reflect.ValueOf(&v), conf.BindParam{Key: "config.sub.value"}, true)
		assert.Nil(t, err)
		assert.Equal(t, v, 99)
		assert.Equal(t, p.ObjectsCount(), 0)

		var cfg struct {
			Sub struct {
				Value Value[int] `value:"${value}"`
			} `value:"${sub}"`
		}

		err = p.RefreshField(reflect.ValueOf(&cfg), conf.BindParam{Key: "config"}, false)
		assert.Nil(t, err)
		assert.Equal(t, cfg.Sub.Value.Value(), 99)
		assert.Equal(t, p.ObjectsCount(), 0)

		prop = conf.Map(map[string]interface{}{
			"config.sub.value": "abc",
		})
		err = p.Refresh(prop)
		assert.Nil(t, err)

		err = p.RefreshField(reflect.ValueOf(&cfg), conf.BindParam{Key: "config"}, true)
		assert.Error(t, err, "strconv.ParseInt: parsing \"abc\": invalid syntax")
		assert.Equal(t, p.ObjectsCount(), 1)

		prop = conf.Map(map[string]interface{}{
			"config.sub.value": "xyz",
		})
		err = p.Refresh(prop)
		assert.Error(t, err, "strconv.ParseInt: parsing \"xyz\": invalid syntax")

		mock := &MockErrorRefreshable{}
		err = p.RefreshField(reflect.ValueOf(mock), conf.BindParam{Key: "config"}, true)
		assert.Error(t, err, "mock error")
		assert.Equal(t, p.ObjectsCount(), 2)
	})

	t.Run("refresh panic", func(t *testing.T) {
		p := New(conf.New())

		assert.Panic(t, func() {
			mock := &MockPanicRefreshable{}
			_ = p.RefreshField(reflect.ValueOf(mock), conf.BindParam{Key: "error"}, true)
		}, "mock panic")
		assert.Equal(t, p.ObjectsCount(), 1)

		assert.Panic(t, func() {
			prop := conf.Map(map[string]interface{}{
				"error.key": "value",
			})
			_ = p.Refresh(prop)
		}, "mock panic")
	})

	t.Run("concurrent refresh", func(t *testing.T) {
		p := New(conf.New())
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				prop := conf.Map(map[string]interface{}{
					"key": time.Now().String(),
				})
				_ = p.Refresh(prop)
			}()
		}

		wg.Wait()
	})

	t.Run("key matching", func(t *testing.T) {
		p := New(conf.New())

		mock1 := &Value[struct {
			C struct {
				D string `value:"${d:=xyz}"`
			} `value:"${c}"`
		}]{}
		err := p.RefreshField(reflect.ValueOf(mock1), conf.BindParam{Key: "a.b"}, true)
		assert.Nil(t, err)
		assert.Equal(t, mock1.Value().C.D, "xyz")

		mock2 := &Value[struct {
			D string `value:"${d:=xyz}"`
		}]{}
		err = p.RefreshField(reflect.ValueOf(mock2), conf.BindParam{Key: "a.b.c"}, true)
		assert.Nil(t, err)
		assert.Equal(t, mock2.Value().D, "xyz")

		prop := conf.Map(map[string]interface{}{
			"a.b.c.d": "123",
		})
		err = p.Refresh(prop)
		assert.Nil(t, err)
		assert.Equal(t, mock1.Value().C.D, "123")
		assert.Equal(t, mock2.Value().D, "123")
	})

}
