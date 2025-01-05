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

// ObjectsCount 绑定的可刷新对象数量
func (p *Properties) ObjectsCount() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return len(p.objects)
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

	// 找出需要更新的对象，一个对象可能对应多个 key，因此需要去重
	updateIndexes := make(map[int]*refreshObject)
	for _, key := range keys {
		for index, o := range p.objects {
			s := strings.TrimPrefix(key, o.param.Key)
			if len(s) == len(key) { // 是否去除了前缀
				continue
			}
			if len(s) == 0 || s[0] == '.' || s[0] == '[' {
				if _, ok := updateIndexes[index]; !ok {
					updateIndexes[index] = o
				}
			}
		}
	}

	updateObjects := make([]*refreshObject, 0, len(updateIndexes))
	{
		ints := make([]int, 0, len(updateIndexes))
		for k := range updateIndexes {
			ints = append(ints, k)
		}
		sort.Ints(ints)
		for _, k := range ints {
			updateObjects = append(updateObjects, updateIndexes[k])
		}
	}

	if len(updateObjects) == 0 {
		return nil
	}
	return p.refreshObjects(updateObjects)
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
	for i, err := range e.arr {
		sb.WriteString(err.Error())
		if i < len(e.arr)-1 {
			sb.WriteString("\n")
		}
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

// RefreshBean 刷新一个 bean 对象，根据 watch 的值决定是否添加监听
func (p *Properties) RefreshBean(v Refreshable, param conf.BindParam, watch bool) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.doRefresh(v, param, watch)
}

func (p *Properties) doRefresh(v Refreshable, param conf.BindParam, watch bool) error {
	if watch {
		p.objects = append(p.objects, &refreshObject{
			target: v,
			param:  param,
		})
	}
	return v.OnRefresh(p.prop, param)
}

type filter struct {
	*Properties
	watch bool
}

func (f *filter) Do(i interface{}, param conf.BindParam) (bool, error) {
	v, ok := i.(Refreshable)
	if !ok {
		return false, nil
	}
	return true, f.doRefresh(v, param, f.watch)
}

// RefreshField 刷新一个 bean 的 field，根据 watch 的值决定是否添加监听
func (p *Properties) RefreshField(v reflect.Value, param conf.BindParam, watch bool) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	f := &filter{Properties: p, watch: watch}
	if v.Kind() == reflect.Ptr {
		ok, err := f.Do(v.Interface(), param)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return conf.BindValue(p.prop, v.Elem(), v.Elem().Type(), param, f)
}
