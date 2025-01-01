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

package dync

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/go-spring/spring-core/conf"
	"go.uber.org/atomic"
)

// ValueInterface 可动态刷新的对象
type ValueInterface interface {
	OnRefresh(prop conf.ReadOnlyProperties, param conf.BindParam) error
}

// Value 可动态刷新的对象
type Value[T interface{}] struct {
	v atomic.Value
}

func (r *Value[T]) OnRefresh(prop conf.ReadOnlyProperties, param conf.BindParam) error {
	var o T
	v := reflect.ValueOf(&o).Elem()
	err := conf.BindValue(prop, v, v.Type(), param, nil)
	if err != nil {
		return err
	}
	r.v.Store(o)
	return nil
}

func (r *Value[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.v.Load())
}

type Field struct {
	value ValueInterface
	param conf.BindParam
}

// Properties 动态属性
type Properties struct {
	value  atomic.Value
	fields []*Field
}

func New() *Properties {
	p := &Properties{}
	p.value.Store(conf.New())
	return p
}

func (p *Properties) load() *conf.Properties {
	return p.value.Load().(*conf.Properties)
}

func (p *Properties) Data() conf.ReadOnlyProperties {
	return p.load()
}

func (p *Properties) Refresh(prop conf.ReadOnlyProperties) (err error) {

	old := p.load()
	oldKeys := old.Keys()
	newKeys := prop.Keys()

	changes := make(map[string]struct{})
	{
		for _, k := range newKeys {
			if !old.Has(k) || old.Get(k) != prop.Get(k) {
				changes[k] = struct{}{}
			}
		}
		for _, k := range oldKeys {
			if _, ok := changes[k]; !ok {
				changes[k] = struct{}{}
			}
		}
	}

	keys := make([]string, 0, len(changes))
	for k := range changes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return p.refreshKeys(prop, keys)
}

func (p *Properties) refreshKeys(prop conf.ReadOnlyProperties, keys []string) (err error) {

	updateIndexes := make(map[int]*Field)
	for _, key := range keys {
		for index, field := range p.fields {
			s := strings.TrimPrefix(key, field.param.Key)
			if len(s) == len(key) {
				continue
			}
			if len(s) == 0 || s[0] == '.' || s[0] == '[' {
				if _, ok := updateIndexes[index]; !ok {
					updateIndexes[index] = field
				}
			}
		}
	}

	updateFields := make([]*Field, 0, len(updateIndexes))
	{
		ints := make([]int, 0, len(updateIndexes))
		for k := range updateIndexes {
			ints = append(ints, k)
		}
		sort.Ints(ints)
		for _, k := range ints {
			updateFields = append(updateFields, updateIndexes[k])
		}
	}

	return p.refreshFields(prop, updateFields)
}

func (p *Properties) refreshFields(prop conf.ReadOnlyProperties, fields []*Field) (err error) {

	old := p.load()
	defer func() {
		if r := recover(); err != nil || r != nil {
			if err == nil {
				err = fmt.Errorf("%v", r)
			}
			p.value.Store(old)
			_ = refreshFields(old, fields)
		}
	}()

	p.value.Store(prop)
	return refreshFields(p.load(), fields)
}

func refreshFields(prop conf.ReadOnlyProperties, fields []*Field) error {
	for _, f := range fields {
		err := f.value.OnRefresh(prop, f.param)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Properties) BindValue(v reflect.Value, param conf.BindParam) error {
	if v.Kind() == reflect.Ptr {
		ok, err := p.bindValue(v.Interface(), param)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return conf.BindValue(p.load(), v.Elem(), v.Elem().Type(), param, p.bindValue)
}

func (p *Properties) bindValue(i interface{}, param conf.BindParam) (bool, error) {

	v, ok := i.(ValueInterface)
	if !ok {
		return false, nil
	}

	prop := p.load()
	err := v.OnRefresh(prop, param)
	if err != nil {
		return false, err
	}

	p.fields = append(p.fields, &Field{
		value: v,
		param: param,
	})
	return true, nil
}
