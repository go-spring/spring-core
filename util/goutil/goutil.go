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

package goutil

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/go-spring/spring-core/util/syslog"
)

// OnPanic is a callback function triggered when a panic occurs.
var OnPanic = func(r interface{}) {
	syslog.Errorf("panic: %v\n%s", r, debug.Stack())
}

// Status provides a mechanism to wait for a goroutine to finish.
type Status struct {
	wg sync.WaitGroup
}

// newStatus creates a new Status object.
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

// Go provides a framework for running a goroutine with built-in recover,
// preventing wild goroutines from crashing the process.
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
