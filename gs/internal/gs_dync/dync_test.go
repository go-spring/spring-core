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
	assert.ThatError(t, err).Matches("bind path= type=int error << property \"key\" isn't simple value")

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
	assert.ThatString(t, string(b)).JSONEqual(`{"key":59}`)
}

func TestDync(t *testing.T) {

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
		assert.ThatError(t, err).Matches("strconv.ParseInt: parsing \"xyz\": invalid syntax")
		assert.ThatError(t, err).Matches("strconv.ParseInt: parsing \"abc\": invalid syntax")

		s1 := &Value[string]{}
		err = p.RefreshField(reflect.ValueOf(s1), conf.BindParam{Key: "config.s3.value"})
		assert.That(t, err).Nil()
		assert.That(t, s1.Value()).Equal("xyz")
		assert.That(t, p.ObjectsCount()).Equal(3)

		s2 := &Value[int]{}
		err = p.RefreshField(reflect.ValueOf(s2), conf.BindParam{Key: "config.s3.value"})
		assert.ThatError(t, err).Matches("strconv.ParseInt: parsing \\\"xyz\\\": invalid syntax")
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
		assert.ThatError(t, err).Matches("strconv.ParseInt: parsing \"xyz\": invalid syntax")

		err = p.Refresh(conf.Map(map[string]any{
			"config.s1.value": "10",
		}))
		assert.That(t, err).Nil()
		assert.That(t, v.Value().S1.Value).Equal(10)
	})
}
