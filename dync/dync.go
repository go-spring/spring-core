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
	"sync"
	"sync/atomic"

	"github.com/go-spring/spring-core/conf"
)

// Refreshable 可动态刷新的对象
type Refreshable interface {
	OnRefresh(prop conf.ReadOnlyProperties, param conf.BindParam) error
}

// Value 可动态刷新的对象
type Value[T interface{}] struct {
	v atomic.Value
}

// Value 获取值
func (r *Value[T]) Value() T {
	return r.v.Load().(T)
}

// OnRefresh 实现 Refreshable 接口
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

// MarshalJSON 实现 json.Marshaler 接口
func (r *Value[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.v.Load())
}

// refreshObject 绑定的可刷新对象
type refreshObject struct {
	target Refreshable
	param  conf.BindParam
}

// Properties 动态属性
type Properties struct {
	prop    conf.ReadOnlyProperties
	lock    sync.RWMutex
	objects []*refreshObject
}

// New 创建一个 Properties 对象
func New() *Properties {
	return &Properties{
		prop: conf.New(),
	}
}

// Data 获取属性列表
func (p *Properties) Data() conf.ReadOnlyProperties {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.prop
}

// Refresh 更新属性列表以及绑定的可刷新对象
func (p *Properties) Refresh(prop conf.ReadOnlyProperties) (err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	old := p.prop
	p.prop = prop

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
	return p.refreshKeys(keys)
}

func (p *Properties) refreshKeys(keys []string) (err error) {
	var objects []*refreshObject
	for _, key := range keys {
		for _, obj := range p.objects {
			if !strings.HasPrefix(key, obj.param.Key) {
				continue
			}
			s := strings.TrimPrefix(key, obj.param.Key)
			if len(s) == 0 || s[0] == '.' || s[0] == '[' {
				objects = append(objects, obj)
			}
		}
	}
	if len(objects) == 0 {
		return nil
	}
	return p.refreshObjects(objects)
}

// Errors 错误列表
type Errors struct {
	arr []error
}

// Len 错误数量
func (e *Errors) Len() int {
	return len(e.arr)
}

// Append 添加一个错误
func (e *Errors) Append(err error) {
	if err != nil {
		e.arr = append(e.arr, err)
	}
}

// Error 实现 error 接口
func (e *Errors) Error() string {
	var sb strings.Builder
	for _, err := range e.arr {
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}

func (p *Properties) refreshObjects(objects []*refreshObject) error {
	ret := &Errors{}
	for _, f := range objects {
		err := p.safeRefreshObject(f)
		ret.Append(err)
	}
	if ret.Len() == 0 {
		return nil
	}
	return ret
}

func (p *Properties) safeRefreshObject(f *refreshObject) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return f.target.OnRefresh(p.prop, f.param)
}

// AddBean 添加一个可刷新对象
func (p *Properties) AddBean(v Refreshable, param conf.BindParam) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.addObjectNoLock(v, param)
}

func (p *Properties) addObjectNoLock(v Refreshable, param conf.BindParam) error {
	p.objects = append(p.objects, &refreshObject{
		target: v,
		param:  param,
	})
	return v.OnRefresh(p.prop, param)
}

// AddField 添加一个 bean 的 field
func (p *Properties) AddField(v reflect.Value, param conf.BindParam) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if v.Kind() == reflect.Ptr {
		ok, err := p.bindValue(v.Interface(), param)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return conf.BindValue(p.prop, v.Elem(), v.Elem().Type(), param, p.bindValue)
}

func (p *Properties) bindValue(i interface{}, param conf.BindParam) (bool, error) {
	v, ok := i.(Refreshable)
	if !ok {
		return false, nil
	}
	return true, p.addObjectNoLock(v, param)
}
