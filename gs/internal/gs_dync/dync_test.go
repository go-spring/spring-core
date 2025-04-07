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
		n := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			l := v.NewListener()
			if n == 4 {
				return
			}
			<-l.C
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

	t.Run("refresh panic", func(t *testing.T) {
		p := New(conf.New())

		assert.Panic(t, func() {
			mock := &MockPanicRefreshable{}
			_ = p.RefreshField(reflect.ValueOf(mock), conf.BindParam{Key: "error"}, true)
		}, "mock panic")

		assert.Panic(t, func() {
			prop := conf.Map(map[string]interface{}{
				"error.key": "value",
			})
			_ = p.Refresh(prop)
		}, "mock panic")

		assert.Equal(t, p.ObjectsCount(), 1)
	})

	t.Run("refresh field", func(t *testing.T) {
		p := New(conf.New())
		assert.Equal(t, p.ObjectsCount(), 0)

		prop := conf.Map(map[string]interface{}{
			"config.s1.value": "99",
		})
		err := p.Refresh(prop)
		assert.Nil(t, err)
		assert.Equal(t, p.Data(), prop)

		var v int
		err = p.RefreshField(reflect.ValueOf(&v), conf.BindParam{Key: "config.s1.value"}, true)
		assert.Nil(t, err)
		assert.Equal(t, v, 99)
		assert.Equal(t, p.ObjectsCount(), 0)

		var cfg struct {
			S1 struct {
				Value Value[int] `value:"${value}"`
			} `value:"${s1}"`
			S2 struct {
				Value Value[int] `value:"${value:=123}"`
			} `value:"${s2}"`
		}

		err = p.RefreshField(reflect.ValueOf(&cfg), conf.BindParam{Key: "config"}, false)
		assert.Nil(t, err)
		assert.Equal(t, p.ObjectsCount(), 0)
		assert.Equal(t, cfg.S1.Value.Value(), 99)
		assert.Equal(t, cfg.S2.Value.Value(), 123)

		err = p.RefreshField(reflect.ValueOf(&cfg), conf.BindParam{Key: "config"}, true)
		assert.Nil(t, err)
		assert.Equal(t, p.ObjectsCount(), 2)
		assert.Equal(t, cfg.S1.Value.Value(), 99)
		assert.Equal(t, cfg.S2.Value.Value(), 123)

		prop = conf.Map(map[string]interface{}{
			"config.s1.value": "99",
			"config.s2.value": "456",
			"config.s4.value": "123",
		})
		err = p.Refresh(prop)
		assert.Equal(t, p.ObjectsCount(), 2)
		assert.Equal(t, cfg.S1.Value.Value(), 99)
		assert.Equal(t, cfg.S2.Value.Value(), 456)

		prop = conf.Map(map[string]interface{}{
			"config.s1.value": "99",
			"config.s2.value": "456",
			"config.s3.value": "xyz",
		})
		err = p.Refresh(prop)
		assert.Equal(t, p.ObjectsCount(), 2)
		assert.Equal(t, cfg.S1.Value.Value(), 99)
		assert.Equal(t, cfg.S2.Value.Value(), 456)

		prop = conf.Map(map[string]interface{}{
			"config.s1.value": "xyz",
			"config.s2.value": "abc",
			"config.s3.value": "xyz",
		})
		err = p.Refresh(prop)
		assert.Error(t, err, "strconv.ParseInt: parsing \"xyz\": invalid syntax")
		assert.Error(t, err, "strconv.ParseInt: parsing \"abc\": invalid syntax")

		s1 := &Value[string]{}
		err = p.RefreshField(reflect.ValueOf(s1), conf.BindParam{Key: "config.s3.value"}, false)
		assert.Nil(t, err)
		assert.Equal(t, s1.Value(), "xyz")

		s2 := &Value[int]{}
		err = p.RefreshField(reflect.ValueOf(s2), conf.BindParam{Key: "config.s3.value"}, false)
		assert.Error(t, err, "strconv.ParseInt: parsing \\\"xyz\\\": invalid syntax")
		assert.Equal(t, p.ObjectsCount(), 2)
	})

}
