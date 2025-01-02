/*
 * Copyright 2012-2024 the original author or authors.
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

package dync_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/dync"
	"github.com/stretchr/testify/assert"
)

// todo 自定义类型通过类型转换器实现刷新机制

type StructValue struct {
	Str string `value:"${string:=abc}" expr:"len($)<6"`
	Int int    `value:"${int:=3}" expr:"$<6"`
}

type Config struct {
	Int    dync.Value[int64]             `value:"${int:=3}" expr:"$<6"`
	Float  dync.Value[float64]           `value:"${float:=1.2}"`
	Map    dync.Value[map[string]string] `value:"${map:=}"`
	Slice  dync.Value[[]string]          `value:"${slice:=}"`
	Object dync.Value[StructValue]       `value:"${object:=}"`
}

func newTest() (*dync.Properties, *Config, error) {
	mgr := dync.New()
	cfg := new(Config)
	err := mgr.AddField(reflect.ValueOf(cfg), conf.BindParam{})
	if err != nil {
		return nil, nil, err
	}
	return mgr, cfg, nil
}

func TestDynamic(t *testing.T) {

	t.Run("default", func(t *testing.T) {
		_, cfg, err := newTest()
		if err != nil {
			return
		}
		b, _ := json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":3,"Float":1.2,"Map":{},"Slice":[],"Object":{"Str":"abc","Int":3}}`)
	})

	t.Run("init", func(t *testing.T) {
		_, cfg, err := newTest()
		if err != nil {
			return
		}
		// cfg.Slice.Init(make([]string, 0))
		// cfg.Map.Init(make(map[string]string))
		// cfg.Event.OnEvent(func(prop conf.ReadOnlyProperties, param conf.BindParam) error {
		// 	fmt.Println("event fired.")
		// 	return nil
		// })
		b, _ := json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":3,"Float":1.2,"Map":{},"Slice":[],"Object":{"Str":"abc","Int":3}}`)
	})

	// t.Run("default validate error", func(t *testing.T) {
	// 	mgr := dync.New()
	// 	cfg := new(Config)
	// 	// cfg.Int.OnValidate(func(v int64) error {
	// 	// 	if v < 6 {
	// 	// 		return errors.New("should greeter than 6")
	// 	// 	}
	// 	// 	return nil
	// 	// })
	// 	err := mgr.BindValue(reflect.ValueOf(cfg), conf.BindParam{})
	// 	assert.Error(t, err, "should greeter than 6")
	// })

	t.Run("init validate error", func(t *testing.T) {

		mgr := dync.New()
		cfg := new(Config)
		// cfg.Int.OnValidate(func(v int64) error {
		// 	if v < 3 {
		// 		return errors.New("should greeter than 3")
		// 	}
		// 	return nil
		// })
		// cfg.Slice.Init(make([]string, 0))
		// cfg.Map.Init(make(map[string]string))
		// cfg.Event.OnEvent(func(prop conf.ReadOnlyProperties, param conf.BindParam) error {
		// 	fmt.Println("event fired.")
		// 	return nil
		// })
		err := mgr.AddField(reflect.ValueOf(cfg), conf.BindParam{})
		if err != nil {
			t.Fatal(err)
		}

		b, _ := json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":3,"Float":1.2,"Map":{},"Slice":[],"Object":{"Str":"abc","Int":3}}`)

		p := conf.New()
		p.Set("int", 1)
		p.Set("float", 5.4)
		p.Set("map.a", 3)
		p.Set("map.b", 7)
		p.Set("slice[0]", 2)
		p.Set("slice[1]", 9)
		err = mgr.Refresh(p)
		// assert.Error(t, err, "should greeter than 3")

		b, _ = json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":1,"Float":5.4,"Map":{"a":"3","b":"7"},"Slice":["2","9"],"Object":{"Str":"abc","Int":3}}`)
	})

	t.Run("success", func(t *testing.T) {

		mgr := dync.New()
		cfg := new(Config)
		// cfg.Int.OnValidate(func(v int64) error {
		// 	if v < 3 {
		// 		return errors.New("should greeter than 3")
		// 	}
		// 	return nil
		// })
		// cfg.Slice.Init(make([]string, 0))
		// cfg.Map.Init(make(map[string]string))
		// cfg.Event.OnEvent(func(prop conf.ReadOnlyProperties, param conf.BindParam) error {
		// 	fmt.Println("event fired.")
		// 	return nil
		// })
		err := mgr.AddField(reflect.ValueOf(cfg), conf.BindParam{})
		if err != nil {
			t.Fatal(err)
		}

		b, _ := json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":3,"Float":1.2,"Map":{},"Slice":[],"Object":{"Str":"abc","Int":3}}`)

		p := conf.New()
		p.Set("int", 1)
		p.Set("float", 5.4)
		p.Set("map.a", 3)
		p.Set("map.b", 7)
		p.Set("slice[0]", 2)
		p.Set("slice[1]", 9)
		err = mgr.Refresh(p)
		// assert.Error(t, err, "should greeter than 3")

		b, _ = json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":1,"Float":5.4,"Map":{"a":"3","b":"7"},"Slice":["2","9"],"Object":{"Str":"abc","Int":3}}`)

		p = conf.New()
		p.Set("int", 6)
		p.Set("float", 2.3)
		p.Set("map.a", 1)
		p.Set("map.b", 2)
		p.Set("slice[0]", 3)
		p.Set("slice[1]", 4)
		err = mgr.Refresh(p)
		assert.Error(t, err, "validate failed on \"\\$<6\" for value 6")

		b, _ = json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":1,"Float":2.3,"Map":{"a":"1","b":"2"},"Slice":["3","4"],"Object":{"Str":"abc","Int":3}}`)

		p = conf.New()
		p.Set("int", 4)
		p.Set("float", 2.3)
		p.Set("map.a", 1)
		p.Set("map.b", 2)
		p.Set("slice[0]", 3)
		p.Set("slice[1]", 4)
		mgr.Refresh(p)

		assert.Equal(t, cfg.Int.Value(), int64(4))

		b, _ = json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":4,"Float":2.3,"Map":{"a":"1","b":"2"},"Slice":["3","4"],"Object":{"Str":"abc","Int":3}}`)
	})
}
