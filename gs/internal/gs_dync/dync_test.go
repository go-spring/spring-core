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
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/go-spring/spring-base/testing/assert"
	"github.com/go-spring/spring-core/conf"
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
	assert.That(t, v.Value()).Equal(0)

	refresh := func(prop conf.Properties) error {
		return v.onRefresh(prop, conf.BindParam{Key: "key"})
	}

	err := refresh(conf.Map(map[string]any{
		"key": "42",
	}))
	assert.That(t, err).Nil()
	assert.That(t, v.Value()).Equal(42)

	err = refresh(conf.Map(map[string]any{
		"key": map[string]any{
			"value": "42",
		},
	}))
	assert.Error(t, err).Matches("bind path= type=int error: property \"key\" isn't simple value")

	var wg sync.WaitGroup
	for i := range 5 {
		n := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			l := v.NewListener()
			if n == 4 {
				return
			}
			<-l.C
			assert.That(t, v.Value()).Equal(59)
		}()
	}

	time.Sleep(50 * time.Millisecond)
	err = refresh(conf.Map(map[string]any{
		"key": 59,
	}))
	assert.That(t, err).Nil()

	wg.Wait()

	b, err := json.Marshal(map[string]any{"key": &v})
	assert.That(t, err).Nil()
	assert.String(t, string(b)).JSONEqual(`{"key":59}`)
}

func TestValue_DifferentTypes(t *testing.T) {

	t.Run("string", func(t *testing.T) {
		var v Value[string]
		err := v.onRefresh(
			conf.Map(map[string]any{"key": "hello"}),
			conf.BindParam{Key: "key"},
		)
		assert.That(t, err).Nil()
		assert.That(t, v.Value()).Equal("hello")
	})

	t.Run("bool", func(t *testing.T) {
		var v Value[bool]
		err := v.onRefresh(
			conf.Map(map[string]any{"key": "true"}),
			conf.BindParam{Key: "key"},
		)
		assert.That(t, err).Nil()
		assert.That(t, v.Value()).Equal(true)
	})

	t.Run("float64", func(t *testing.T) {
		var v Value[float64]
		err := v.onRefresh(
			conf.Map(map[string]any{"key": "3.14"}),
			conf.BindParam{Key: "key"},
		)
		assert.That(t, err).Nil()
		assert.That(t, v.Value()).Equal(3.14)
	})

	t.Run("slice", func(t *testing.T) {
		var v Value[[]int]
		err := v.onRefresh(
			conf.Map(map[string]any{"key": []any{1, 2, 3}}),
			conf.BindParam{Key: "key"},
		)
		assert.That(t, err).Nil()
		assert.That(t, v.Value()).Equal([]int{1, 2, 3})
	})
}

func TestValue_ConcurrentAccess(t *testing.T) {
	var v Value[int]

	err := v.onRefresh(
		conf.Map(map[string]any{"key": "100"}),
		conf.BindParam{Key: "key"},
	)
	assert.That(t, err).Nil()
	assert.That(t, v.Value()).Equal(100)

	var wg sync.WaitGroup
	const goroutines = 100

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val := v.Value()
			assert.Number(t, val).Between(0, 100)
		}()
	}

	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			err := v.onRefresh(
				conf.Map(map[string]any{"key": fmt.Sprintf("%d", idx)}),
				conf.BindParam{Key: "key"},
			)
			assert.That(t, err).Nil()
		}(i)
	}

	wg.Wait()
}

func TestValue_Listener(t *testing.T) {
	var v Value[int]

	listeners := make([]*Listener, 5)
	for i := range listeners {
		listeners[i] = v.NewListener()
	}

	go func() {
		err := v.onRefresh(
			conf.Map(map[string]any{"key": "42"}),
			conf.BindParam{Key: "key"},
		)
		assert.That(t, err).Nil()
	}()

	var wg sync.WaitGroup
	for _, l := range listeners {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-l.C:
				assert.That(t, v.Value()).Equal(42)
			case <-time.After(time.Second):
				t.Errorf("timeout")
			}
		}()
	}
	wg.Wait()
}

func TestDync(t *testing.T) {

	t.Run("invalid property format", func(t *testing.T) {
		p := New(conf.New())

		var v Value[int]
		err := p.RefreshField(
			reflect.ValueOf(&v),
			conf.BindParam{Key: "${invalid..key}"},
		)
		assert.That(t, err).NotNil()
	})

	t.Run("missing required property", func(t *testing.T) {
		p := New(conf.New())

		var cfg struct {
			Value Value[int] `value:"${required.property}"`
		}

		err := p.RefreshField(
			reflect.ValueOf(&cfg),
			conf.BindParam{Key: "config"},
		)
		assert.That(t, err).NotNil()
	})

	t.Run("type mismatch error", func(t *testing.T) {
		p := New(conf.Map(map[string]any{
			"config.value": "not_a_number",
		}))

		var v Value[int]
		err := p.RefreshField(
			reflect.ValueOf(&v),
			conf.BindParam{Key: "config.value"},
		)
		assert.Error(t, err).Matches("strconv.ParseInt: parsing.*invalid syntax")
	})

	t.Run("refresh panic", func(t *testing.T) {
		p := New(conf.New())

		assert.Panic(t, func() {
			mock := &MockPanicRefreshable{}
			_ = p.RefreshField(reflect.ValueOf(mock), conf.BindParam{Key: "error"})
		}, "mock panic")

		assert.Panic(t, func() {
			prop := conf.Map(map[string]any{
				"error.key": "value",
			})
			_ = p.Refresh(prop)
		}, "mock panic")

		assert.That(t, p.ObjectsCount()).Equal(1)
	})

	t.Run("refresh field", func(t *testing.T) {
		p := New(conf.New())
		assert.That(t, p.ObjectsCount()).Equal(0)

		prop := conf.Map(map[string]any{
			"config.s1.value": "99",
		})
		err := p.Refresh(prop)
		assert.That(t, err).Nil()
		assert.That(t, p.Data()).Equal(prop)

		var v int
		err = p.RefreshField(reflect.ValueOf(&v), conf.BindParam{Key: "config.s1.value"})
		assert.That(t, err).Nil()
		assert.That(t, v).Equal(99)
		assert.That(t, p.ObjectsCount()).Equal(0)

		var cfg struct {
			S1 struct {
				Value Value[int] `value:"${value}"`
			} `value:"${s1}"`
			S2 struct {
				Value Value[int] `value:"${value:=123}"`
			} `value:"${s2}"`
		}

		err = p.RefreshField(reflect.ValueOf(&cfg), conf.BindParam{Key: "config"})
		assert.That(t, err).Nil()
		assert.That(t, p.ObjectsCount()).Equal(2)
		assert.That(t, cfg.S1.Value.Value()).Equal(99)
		assert.That(t, cfg.S2.Value.Value()).Equal(123)

		prop = conf.Map(map[string]any{
			"config.s1.value": "99",
			"config.s2.value": "456",
			"config.s4.value": "123",
		})
		err = p.Refresh(prop)
		assert.That(t, err).Nil()
		assert.That(t, p.ObjectsCount()).Equal(2)
		assert.That(t, cfg.S1.Value.Value()).Equal(99)
		assert.That(t, cfg.S2.Value.Value()).Equal(456)

		prop = conf.Map(map[string]any{
			"config.s1.value": "99",
			"config.s2.value": "456",
			"config.s3.value": "xyz",
		})
		err = p.Refresh(prop)
		assert.That(t, err).Nil()
		assert.That(t, p.ObjectsCount()).Equal(2)
		assert.That(t, cfg.S1.Value.Value()).Equal(99)
		assert.That(t, cfg.S2.Value.Value()).Equal(456)

		prop = conf.Map(map[string]any{
			"config.s1.value": "xyz",
			"config.s2.value": "abc",
			"config.s3.value": "xyz",
		})
		err = p.Refresh(prop)
		assert.Error(t, err).Matches("strconv.ParseInt: parsing \"xyz\": invalid syntax")
		assert.Error(t, err).Matches("strconv.ParseInt: parsing \"abc\": invalid syntax")

		s1 := &Value[string]{}
		err = p.RefreshField(reflect.ValueOf(s1), conf.BindParam{Key: "config.s3.value"})
		assert.That(t, err).Nil()
		assert.That(t, s1.Value()).Equal("xyz")
		assert.That(t, p.ObjectsCount()).Equal(3)

		s2 := &Value[int]{}
		err = p.RefreshField(reflect.ValueOf(s2), conf.BindParam{Key: "config.s3.value"})
		assert.Error(t, err).Matches("strconv.ParseInt: parsing \\\"xyz\\\": invalid syntax")
		assert.That(t, p.ObjectsCount()).Equal(4)
	})

	t.Run("refresh struct", func(t *testing.T) {
		p := New(conf.Map(map[string]any{
			"config.s1.value": "99",
		}))

		v := &Value[struct {
			S1 struct {
				Value int `value:"${value}"`
			} `value:"${s1}"`
		}]{}

		var param conf.BindParam
		err := param.BindTag("${config}", "")
		assert.That(t, err).Nil()

		err = p.RefreshField(reflect.ValueOf(v), param)
		assert.That(t, err).Nil()
		assert.That(t, v.Value().S1.Value).Equal(99)

		err = p.Refresh(conf.Map(map[string]any{
			"config.s1.value": "xyz",
		}))
		assert.Error(t, err).Matches("strconv.ParseInt: parsing \"xyz\": invalid syntax")

		err = p.Refresh(conf.Map(map[string]any{
			"config.s1.value": "10",
		}))
		assert.That(t, err).Nil()
		assert.That(t, v.Value().S1.Value).Equal(10)
	})

	t.Run("with default value", func(t *testing.T) {
		p := New(conf.New())

		var cfg struct {
			Value Value[int] `value:"${property:=42}"`
		}

		err := p.RefreshField(reflect.ValueOf(&cfg), conf.BindParam{Key: "config"})
		assert.That(t, err).Nil()
		assert.That(t, cfg.Value.Value()).Equal(42)
	})

	t.Run("override default value", func(t *testing.T) {
		p := New(conf.Map(map[string]any{
			"config.property": "100",
		}))

		var cfg struct {
			Value Value[int] `value:"${property:=42}"`
		}

		err := p.RefreshField(reflect.ValueOf(&cfg), conf.BindParam{Key: "config"})
		assert.That(t, err).Nil()
		assert.That(t, cfg.Value.Value()).Equal(100)
	})
}
