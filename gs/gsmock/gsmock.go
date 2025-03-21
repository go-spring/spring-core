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

// Invoke attempts to call a mock implementation of a given method.
func Invoke(r *Manager, typ reflect.Type, method string, params ...interface{}) ([]interface{}, bool) {
	if r == nil || !testing.Testing() {
		return nil, false
	}
	return invoke0(r, typ, method, params...)
}

// InvokeContext attempts to call a mock implementation using context.
func InvokeContext(ctx context.Context, typ reflect.Type, method string, params ...interface{}) ([]interface{}, bool) {
	if !testing.Testing() {
		return nil, false
	}
	r := getManager(ctx)
	if r == nil {
		return nil, false
	}
	return invoke0(r, typ, method, params...)
}

// invoke0 is a helper function that calls a mock implementation of a given method.
func invoke0(r *Manager, typ reflect.Type, method string, params ...interface{}) ([]interface{}, bool) {
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
