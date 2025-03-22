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

package gsmock

import (
	"context"
	"log"
	"reflect"
	"testing"
)

// Mode represents the mocking mode.
type Mode int

const (
	ModeHandle = Mode(iota)
	ModeWhenReturn
)

// Invoker is the interface that all mockers must implement.
type Invoker interface {
	Mode() Mode
	When(params []interface{}) bool
	Return(params []interface{}) []interface{}
	Handle(params []interface{}) ([]interface{}, bool)
}

// mockerKey is used as a key to store mockers in the mock manager.
type mockerKey struct {
	typ    reflect.Type
	method string
}

// Manager manages mock implementations.
type Manager struct {
	mockers map[mockerKey][]Invoker
}

// GetMockers retrieves the list of mockers for a given method.
func (r *Manager) GetMockers(typ reflect.Type, method string) []Invoker {
	return r.mockers[mockerKey{typ, method}]
}

// AddMocker registers a mock implementation for a given method.
func (r *Manager) AddMocker(typ reflect.Type, method string, i Invoker) {
	k := mockerKey{typ, method}
	r.mockers[k] = append(r.mockers[k], i)
}

var managerKey int

// getManager retrieves the mock manager from the given context.
func getManager(ctx context.Context) *Manager {
	if r, ok := ctx.Value(&managerKey).(*Manager); ok {
		return r
	}
	return nil
}

// Init initializes a new mock manager and attaches it to the given context.
func Init(ctx context.Context) (*Manager, context.Context) {
	r := &Manager{
		mockers: make(map[mockerKey][]Invoker),
	}
	return r, context.WithValue(ctx, &managerKey, r)
}

// InvokeContext attempts to call a mock implementation using context.
func InvokeContext(ctx context.Context, typ reflect.Type, method string, params ...interface{}) ([]interface{}, bool) {
	return Invoke(getManager(ctx), typ, method, params...)
}

// Invoke attempts to call a mock implementation of a given method.
func Invoke(r *Manager, typ reflect.Type, method string, params ...interface{}) ([]interface{}, bool) {
	if r == nil || !testing.Testing() {
		return nil, false
	}
	mockers := r.GetMockers(typ, method)
	for _, f := range mockers {
		switch f.Mode() {
		case ModeHandle:
			ret, ok := f.Handle(params)
			if ok {
				return ret, true
			}
		case ModeWhenReturn:
			if f.When(params) {
				ret := f.Return(params)
				return ret, true
			}
		default:
			log.Printf("Warning: unknown mode: %d", f.Mode())
		}
	}
	return nil, false
}

// InvokeContext1 attempts to call a mock implementation using context.
func InvokeContext1[R1 any](ctx context.Context, typ reflect.Type, method string, params ...interface{}) (r1 R1, ok bool) {
	return Invoke1[R1](getManager(ctx), typ, method, params...)
}

// Invoke1 attempts to call a mock implementation of a given method.
func Invoke1[R1 any](r *Manager, typ reflect.Type, method string, params ...interface{}) (r1 R1, ok bool) {
	r1, _, _, ok = Invoke3[R1, any, any](r, typ, method, params...)
	return
}

// InvokeContext2 attempts to call a mock implementation using context.
func InvokeContext2[R1, R2 any](ctx context.Context, typ reflect.Type, method string, params ...interface{}) (r1 R1, r2 R2, ok bool) {
	return Invoke2[R1, R2](getManager(ctx), typ, method, params...)
}

// Invoke2 attempts to call a mock implementation of a given method.
func Invoke2[R1, R2 any](r *Manager, typ reflect.Type, method string, params ...interface{}) (r1 R1, r2 R2, ok bool) {
	r1, r2, _, ok = Invoke3[R1, R2, any](r, typ, method, params...)
	return
}

// InvokeContext3 attempts to call a mock implementation using context.
func InvokeContext3[R1, R2, R3 any](ctx context.Context, typ reflect.Type, method string, params ...interface{}) (r1 R1, r2 R2, r3 R3, ok bool) {
	return Invoke3[R1, R2, R3](getManager(ctx), typ, method, params...)
}

// Invoke3 attempts to call a mock implementation of a given method.
func Invoke3[R1, R2, R3 any](r *Manager, typ reflect.Type, method string, params ...interface{}) (r1 R1, r2 R2, r3 R3, ok bool) {
	ret, ok := Invoke(r, typ, method, params...)
	if !ok {
		return
	}
	switch len(ret) {
	case 1:
		r1, _ = ret[0].(R1)
	case 2:
		r1, _ = ret[0].(R1)
		r2, _ = ret[1].(R2)
	case 3:
		r1, _ = ret[0].(R1)
		r2, _ = ret[1].(R2)
		r3, _ = ret[2].(R3)
	default:
		log.Printf("Warning: unexpected number of return values: %d", len(ret))
	}
	ok = true
	return
}
