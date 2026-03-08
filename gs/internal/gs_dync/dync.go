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

// Package gs_dync provides dynamic configuration binding and refresh
// capabilities for Go-Spring applications.
//
// It allows application components to register themselves as refreshable
// objects that automatically update their internal state whenever the
// underlying configuration changes.
//
// Key components:
//   - Properties: holds the current configuration and manages all
//     registered `refreshable` objects.
//   - Value[T]: a type-safe container for dynamic configuration values.
//   - Listener: allows components to receive change notifications.
//   - `refreshable`: interface that application components can implement
//     to react to configuration updates.
//
// This package is designed to be thread-safe and suitable for hot-reload
// scenarios in long-running applications.
package gs_dync

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/stdlib/flatten"
)

// refreshable represents an object that can be dynamically refreshed.
type refreshable interface {
	onRefresh(prop flatten.Storage, param conf.BindParam, commit bool) error
}

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

// onRefresh updates the stored value with new properties and notifies listeners.
func (r *Value[T]) onRefresh(prop flatten.Storage, param conf.BindParam, commit bool) error {
	t := reflect.TypeFor[T]()
	v := reflect.New(t).Elem()
	if err := conf.BindValue(prop, v, t, param, nil); err != nil {
		return err
	}
	if commit {
		r.v.Store(v.Interface())
	}
	return nil
}

// MarshalJSON serializes the stored value as JSON.
func (r *Value[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.v.Load())
}

// refreshObject represents an object bound to dynamic properties that can be refreshed.
type refreshObject struct {
	target refreshable    // The refreshable object.
	param  conf.BindParam // Parameters used for refreshing.
}

// Properties manages dynamic properties and refreshable objects.
type Properties struct {
	prop    flatten.Storage  // The current properties.
	lock    sync.RWMutex     // A read-write lock for thread-safe access.
	objects []*refreshObject // List of refreshable objects bound to the properties.
}

// New creates and returns a new Properties instance.
func New(p flatten.Storage) *Properties {
	return &Properties{
		prop: p,
	}
}

// Data returns the current properties.
func (p *Properties) Data() flatten.Storage {
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
func (p *Properties) Refresh(prop flatten.Storage) (err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.prop = prop
	if len(p.objects) == 0 {
		return nil
	}

	// 首先预刷新所有动态值，校验通过之后进行提交
	if err = p.refreshObjects(p.objects, false); err != nil {
		return err
	}
	return p.refreshObjects(p.objects, true)
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
			sb.WriteString("; ")
		}
	}
	return sb.String()
}

// refreshObjects refreshes all provided objects and aggregates errors.
func (p *Properties) refreshObjects(objects []*refreshObject, commit bool) error {
	ret := &Errors{}
	for _, obj := range objects {
		err := obj.target.onRefresh(p.prop, obj.param, commit)
		ret.Append(err)
	}
	if ret.Len() == 0 {
		return nil
	}
	return ret
}

// filter is used to selectively refresh objects and fields.
type filter struct {
	*Properties
}

// Do attempts to refresh a single object if it implements the [refreshable] interface.
func (f *filter) Do(i any, param conf.BindParam) (bool, error) {
	v, ok := i.(refreshable)
	if !ok || v == nil {
		return false, nil
	}
	f.objects = append(f.objects, &refreshObject{
		target: v,
		param:  param,
	})
	return true, v.onRefresh(f.prop, param, true)
}

// RefreshField refreshes a field of a bean, optionally registering it as refreshable.
func (p *Properties) RefreshField(v reflect.Value, param conf.BindParam) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	f := &filter{Properties: p}
	if v.Kind() == reflect.Pointer {
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
