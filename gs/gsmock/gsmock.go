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
	"testing"
)

type Mode int

const (
	ModeHandle = Mode(iota)
	ModeWhenReturn
)

// Invoker is the interface that all mockers must implement.
type Invoker interface {
	Mode() Mode
	When(ctx context.Context, params []interface{}) bool
	Return(ctx context.Context, params []interface{}) []interface{}
	Handle(ctx context.Context, params []interface{}) ([]interface{}, bool)
}

// Rooter is the mock manager that stores mock implementations.
type Rooter struct {
	mockers map[string][]Invoker
}

// GetMockers retrieves the list of mockers for a given method.
func (r *Rooter) GetMockers(method string) []Invoker {
	return r.mockers[method]
}

// AddMocker adds a mock implementation for a given method.
func (r *Rooter) AddMocker(method string, i Invoker) {
	r.mockers[method] = append(r.mockers[method], i)
}

var rooterKey int

// getRooter retrieves the mock manager from the given context.
func getRooter(ctx context.Context) *Rooter {
	if r, ok := ctx.Value(&rooterKey).(*Rooter); ok {
		return r
	}
	return nil
}

// Init initializes the mock manager and attaches it to the given context.
func Init(ctx context.Context) (*Rooter, context.Context) {
	r := &Rooter{
		mockers: make(map[string][]Invoker),
	}
	return r, context.WithValue(ctx, &rooterKey, r)
}

// Invoke attempts to call a mock implementation of a given method.
func Invoke(method string, ctx context.Context, params ...interface{}) ([]interface{}, bool) {
	if !testing.Testing() {
		return nil, false
	}
	if r := getRooter(ctx); r != nil {
		mockers := r.GetMockers(method)
		for _, f := range mockers {
			switch f.Mode() {
			case ModeHandle:
				ret, ok := f.Handle(ctx, params)
				if ok {
					return ret, true
				}
			case ModeWhenReturn:
				if f.When(ctx, params) {
					ret := f.Return(ctx, params)
					return ret, true
				}
			default:
				log.Printf("Warning: unknown mode: %d", f.Mode())
			}
		}
	}
	return nil, false
}
