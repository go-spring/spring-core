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

package goutil

import (
	"context"
	"errors"
	"runtime/debug"
	"sync"

	"github.com/go-spring/spring-core/util/syslog"
)

// OnPanic is a global callback function triggered when a panic occurs.
var OnPanic = func(r interface{}) {
	syslog.Errorf("panic: %v\n%s", r, debug.Stack())
}

/********************************** go ***************************************/

// Status provides a mechanism to wait for a goroutine to finish.
type Status struct {
	wg sync.WaitGroup
}

// newStatus creates and initializes a new Status object.
func newStatus() *Status {
	s := &Status{}
	s.wg.Add(1)
	return s
}

// done marks the goroutine as finished.
func (s *Status) done() {
	s.wg.Done()
}

// Wait waits for the goroutine to finish.
func (s *Status) Wait() {
	s.wg.Wait()
}

// Go runs a goroutine safely with context support and panic recovery.
// It ensures the process does not crash due to an uncaught panic in the goroutine.
func Go(ctx context.Context, f func(ctx context.Context)) *Status {
	s := newStatus()
	go func() {
		defer s.done()
		defer func() {
			if r := recover(); r != nil {
				if OnPanic != nil {
					OnPanic(r)
				}
			}
		}()
		f(ctx)
	}()
	return s
}

// GoFunc runs a goroutine safely with panic recovery.
// It ensures the process does not crash due to an uncaught panic in the goroutine.
func GoFunc(f func()) *Status {
	s := newStatus()
	go func() {
		defer s.done()
		defer func() {
			if r := recover(); r != nil {
				if OnPanic != nil {
					OnPanic(r)
				}
			}
		}()
		f()
	}()
	return s
}

/******************************* go with value *******************************/

// ValueStatus provides a mechanism to wait for a goroutine that returns a value and an error.
type ValueStatus[T any] struct {
	wg  sync.WaitGroup
	val T
	err error
}

// newValueStatus creates and initializes a new ValueStatus object.
func newValueStatus[T any]() *ValueStatus[T] {
	s := &ValueStatus[T]{}
	s.wg.Add(1)
	return s
}

// done marks the goroutine as finished.
func (s *ValueStatus[T]) done() {
	s.wg.Done()
}

// Wait blocks until the goroutine finishes and returns its result and error.
func (s *ValueStatus[T]) Wait() (T, error) {
	s.wg.Wait()
	return s.val, s.err
}

// GoValue runs a goroutine safely with context support and panic recovery and
// returns its result and error.
// It ensures the process does not crash due to an uncaught panic in the goroutine.
func GoValue[T any](ctx context.Context, f func(ctx context.Context) (T, error)) *ValueStatus[T] {
	s := newValueStatus[T]()
	go func() {
		defer s.done()
		defer func() {
			if r := recover(); r != nil {
				if OnPanic != nil {
					OnPanic(r)
				}
				s.err = errors.New("panic occurred")
			}
		}()
		s.val, s.err = f(ctx)
	}()
	return s
}
