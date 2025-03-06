/*
 * Copyright 2024 The Go-Spring Authors.
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

// Package gs_dync provides dynamic properties and refreshable objects.
package gs_dync

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
)

// Value represents a thread-safe object that can dynamically refresh its value.
type Value[T any] struct {
	v atomic.Value
}

// Value retrieves the current value stored in the object.
// If no value is set, it returns the zero value for the type T.
func (r *Value[T]) Value() T {
	v, ok := r.v.Load().(T)
	if !ok {
		var zero T
		return zero
	}
	return v
}

// OnRefresh refreshes the value of the object by binding new properties.
func (r *Value[T]) OnRefresh(prop gs.Properties, param conf.BindParam) error {
	var o T
	v := reflect.ValueOf(&o).Elem()
	err := conf.BindValue(prop, v, v.Type(), param, nil)
	if err != nil {
		return err
	}
	r.v.Store(o)
	return nil
}

// MarshalJSON serializes the stored value as JSON.
func (r *Value[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.v.Load())
}

// refreshObject represents an object bound to dynamic properties that can be refreshed.
type refreshObject struct {
	target gs.Refreshable // The refreshable object.
	param  conf.BindParam // Parameters used for refreshing.
}

// Properties manages dynamic properties and refreshable objects.
type Properties struct {
	prop    gs.Properties    // The current properties.
	lock    sync.RWMutex     // A read-write lock for thread-safe access.
	objects []*refreshObject // List of refreshable objects bound to the properties.
}

// New creates and returns a new Properties instance.
func New() *Properties {
	return &Properties{
		prop: conf.New(),
	}
}

// Data returns the current properties.
func (p *Properties) Data() gs.Properties {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.prop
}

// ObjectsCount returns the number of registered refreshable objects.
func (p *Properties) ObjectsCount() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return len(p.objects)
}

// Refresh updates the properties and refreshes all bound objects as necessary.
func (p *Properties) Refresh(prop gs.Properties) (err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(p.objects) == 0 {
		p.prop = prop
		return nil
	}

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

	// Refresh objects based on changed keys.
	keys := make([]string, 0, len(changes))
	for k := range changes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return p.refreshKeys(keys)
}

// refreshKeys refreshes objects bound by the specified keys.
func (p *Properties) refreshKeys(keys []string) (err error) {
	updateIndexes := make(map[int]*refreshObject)
	for _, key := range keys {
		for index, o := range p.objects {
			s := strings.TrimPrefix(key, o.param.Key)
			if len(s) == len(key) { // Check if the key starts with the parameter key.
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

// Errors represents a collection of errors.
type Errors struct {
	arr []error
}

// Len returns the number of errors.
func (e *Errors) Len() int {
	return len(e.arr)
}

// Append adds an error to the collection if it is non-nil.
func (e *Errors) Append(err error) {
	if err != nil {
		e.arr = append(e.arr, err)
	}
}

// Error concatenates all errors into a single string.
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

// refreshObjects refreshes all provided objects and aggregates errors.
func (p *Properties) refreshObjects(objects []*refreshObject) error {
	ret := &Errors{}
	for _, obj := range objects {
		err := p.safeRefreshObject(obj)
		ret.Append(err)
	}
	if ret.Len() == 0 {
		return nil
	}
	return ret
}

// safeRefreshObject refreshes an object and recovers from panics.
func (p *Properties) safeRefreshObject(obj *refreshObject) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return obj.target.OnRefresh(p.prop, obj.param)
}

// RefreshBean refreshes a single refreshable object.
// Optionally registers the object as refreshable if watch is true.
func (p *Properties) RefreshBean(v gs.Refreshable, param conf.BindParam, watch bool) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.doRefresh(v, param, watch)
}

// doRefresh performs the refresh operation and registers the object if watch is enabled.
func (p *Properties) doRefresh(v gs.Refreshable, param conf.BindParam, watch bool) error {
	if watch {
		p.objects = append(p.objects, &refreshObject{
			target: v,
			param:  param,
		})
	}
	return v.OnRefresh(p.prop, param)
}

// filter is used to selectively refresh objects and fields.
type filter struct {
	*Properties
	watch bool // Whether to register objects as refreshable.
}

// Do attempts to refresh a single object if it implements the [gs.Refreshable] interface.
func (f *filter) Do(i interface{}, param conf.BindParam) (bool, error) {
	v, ok := i.(gs.Refreshable)
	if !ok {
		return false, nil
	}
	return true, f.doRefresh(v, param, f.watch)
}

// RefreshField refreshes a field of a bean, optionally registering it as refreshable.
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
