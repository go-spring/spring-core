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

// Package goutil provides safe goroutine utilities with built-in panic recovery.
//
// Goroutines may panic due to programming errors such as nil pointer dereference
// or out-of-bounds access. Panics can be recovered without crashing the whole
// process. This package offers wrappers to run goroutines safely and recover
// from such panics.
//
// A global `OnPanic` handler is triggered whenever a panic is recovered. It allows
// developers to log the panic, report metrics, or perform other custom recovery
// logic, making it easier to monitor and debug failures in concurrent code.
package goutil

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/go-spring/spring-base/util"
)

// OnPanic is a global callback function triggered whenever a panic is recovered
// inside a goroutine launched by this package.
//
// By default it prints the panic value and stack trace to stdout.
// Applications may override it during initialization to provide custom logging,
// metrics, or alerting.
//
// Note: being global means it is shared across all usages. In testing
// scenarios, remember to restore it after modification if necessary.
var OnPanic = func(ctx context.Context, r any, stack []byte) {
	fmt.Printf("[PANIC] %v\n%s\n", r, stack)
}

/********************************** go ***************************************/

// Status provides a handle to wait for a goroutine to finish.
type Status struct {
	wg sync.WaitGroup
}

// newStatus creates and initializes a new Status.
func newStatus() *Status {
	s := &Status{}
	s.wg.Add(1)
	return s
}

// done marks the goroutine as finished.
func (s *Status) done() {
	s.wg.Done()
}

// Wait blocks until the goroutine completes.
func (s *Status) Wait() {
	s.wg.Wait()
}

// Go launches a goroutine that recovers from panics and invokes the global
// OnPanic handler when a panic occurs.
//
// The provided context is passed to the goroutine function `f` and to OnPanic.
// The goroutine does not stop automatically when the context is cancelled;
// `f` should check `ctx.Done()` and return when appropriate.
func Go(ctx context.Context, f func(ctx context.Context)) *Status {
	s := newStatus()
	go func() {
		defer s.done()
		defer func() {
			if r := recover(); r != nil {
				if OnPanic != nil {
					OnPanic(ctx, r, debug.Stack())
				}
			}
		}()
		f(ctx)
	}()
	return s
}

/******************************* go with value *******************************/

// ValueStatus represents a goroutine that returns a value and an error.
// It allows the caller to wait for the result.
type ValueStatus[T any] struct {
	wg  sync.WaitGroup
	val T
	err error
}

// newValueStatus creates and initializes a new ValueStatus.
func newValueStatus[T any]() *ValueStatus[T] {
	s := &ValueStatus[T]{}
	s.wg.Add(1)
	return s
}

// done marks the goroutine as finished.
func (s *ValueStatus[T]) done() {
	s.wg.Done()
}

// Wait blocks until the goroutine completes and returns its value and error.
func (s *ValueStatus[T]) Wait() (T, error) {
	s.wg.Wait()
	return s.val, s.err
}

// GoValue launches a goroutine that executes the provided function `f`,
// recovers from any panic, and invokes the global OnPanic handler.
//
// The context is passed to both `f` and OnPanic. The caller must ensure
// that `f` observes `ctx.Done()` if early cancellation is desired.
//
// If a panic occurs, the recovered panic and stack trace are also reported
// via OnPanic and wrapped into the returned error.
func GoValue[T any](ctx context.Context, f func(ctx context.Context) (T, error)) *ValueStatus[T] {
	s := newValueStatus[T]()
	go func() {
		defer s.done()
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				if OnPanic != nil {
					OnPanic(ctx, r, stack)
				}
				s.err = util.FormatError(nil, "panic recovered: %v\n%s", r, stack)
			}
		}()
		s.val, s.err = f(ctx)
	}()
	return s
}
