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

func TestValue(t *testing.T) {

	var v Value[int]
	assert.Equal(t, v.Value(), 0)

	prop := conf.Map(map[string]interface{}{
		"key": "42",
	})
	err := v.OnRefresh(prop, conf.BindParam{Key: "key"})
	assert.Nil(t, err)
	assert.Equal(t, v.Value(), 42)

	b, err := json.Marshal(map[string]interface{}{"key": &v})
	assert.Nil(t, err)
	assert.JsonEqual(t, string(b), `{"key":42}`)

	prop = conf.Map(map[string]interface{}{
		"key": map[string]interface{}{
			"value": "42",
		},
	})
	err = v.OnRefresh(prop, conf.BindParam{Key: "key"})
	assert.Error(t, err, "bind path= type=int error << property key isn't simple value")

}

func TestDync(t *testing.T) {

	t.Run("refresh bean", func(t *testing.T) {
		p := New()
		assert.Equal(t, p.ObjectsCount(), 0)

		prop := conf.Map(map[string]interface{}{
			"key": "value",
		})
		err := p.Refresh(prop)
		assert.Nil(t, err)
		assert.Equal(t, p.Data(), prop)

		bean := &MockRefreshable{}
		err = p.RefreshBean(bean, conf.BindParam{Key: "test"}, false)
		assert.Nil(t, err)
		assert.True(t, bean.called)
		assert.Equal(t, p.ObjectsCount(), 0)

		bean = &MockRefreshable{}
		err = p.RefreshBean(bean, conf.BindParam{Key: "test"}, true)
		assert.Nil(t, err)
		assert.True(t, bean.called)
		assert.Equal(t, p.ObjectsCount(), 1)

		bean.called = false
		prop = conf.Map(map[string]interface{}{
			"test.value": "new",
		})
		err = p.Refresh(prop)
		assert.Nil(t, err)
		assert.True(t, bean.called)

		bean.called = false
		prop = conf.Map(map[string]interface{}{
			"test.value": "new",
		})
		err = p.Refresh(prop)
		assert.Nil(t, err)
		assert.False(t, bean.called)

		mock := &MockErrorRefreshable{}
		err = p.RefreshBean(mock, conf.BindParam{Key: "error"}, true)
		assert.Error(t, err, "mock error")
		assert.Equal(t, p.ObjectsCount(), 2)

		mock = &MockErrorRefreshable{}
		err = p.RefreshBean(mock, conf.BindParam{Key: "error"}, true)
		assert.Error(t, err, "mock error")
		assert.Equal(t, p.ObjectsCount(), 3)

		prop = conf.Map(map[string]interface{}{
			"error.key": "value",
		})
		err = p.Refresh(prop)
		assert.Error(t, err, "mock error; mock error")
	})

	t.Run("refresh field", func(t *testing.T) {
		p := New()
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

		bean := &MockRefreshable{}
		err = p.RefreshField(reflect.ValueOf(bean), conf.BindParam{Key: "config"}, true)
		assert.Nil(t, err)
		assert.Equal(t, p.ObjectsCount(), 3)
	})

	t.Run("refresh panic", func(t *testing.T) {
		p := New()
		mock := &MockPanicRefreshable{}

		assert.Panic(t, func() {
			_ = p.RefreshBean(mock, conf.BindParam{Key: "error"}, true)
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
		p := New()
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
		p := New()
		mock1 := &MockRefreshable{}
		_ = p.RefreshBean(mock1, conf.BindParam{Key: "a.b"}, true)
		mock2 := &MockRefreshable{}
		_ = p.RefreshBean(mock2, conf.BindParam{Key: "a.b.c"}, true)
		prop := conf.Map(map[string]interface{}{
			"a.b.c.d": "value",
		})
		err := p.Refresh(prop)
		assert.Nil(t, err)
		assert.True(t, mock1.called)
		assert.True(t, mock2.called)
	})

}

type MockRefreshable struct {
	called bool
}

func (m *MockRefreshable) OnRefresh(prop conf.Properties, param conf.BindParam) error {
	m.called = true
	return nil
}

type MockErrorRefreshable struct{}

func (m *MockErrorRefreshable) OnRefresh(prop conf.Properties, param conf.BindParam) error {
	return errors.New("mock error")
}

type MockPanicRefreshable struct{}

func (m *MockPanicRefreshable) OnRefresh(prop conf.Properties, param conf.BindParam) error {
	panic("mock panic")
}
