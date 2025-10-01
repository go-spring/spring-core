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

package gs_app

import (
	"sync"
	"sync/atomic"
)

// ReadySignal is a synchronization helper used to indicate
// when an application is ready to serve requests.
type ReadySignal struct {
	wg sync.WaitGroup
	ch chan struct{}
	b  atomic.Bool
}

// NewReadySignal creates and returns a new ReadySignal instance.
func NewReadySignal() *ReadySignal {
	return &ReadySignal{
		ch: make(chan struct{}),
	}
}

// Add increments the WaitGroup counter.
func (s *ReadySignal) Add() {
	s.wg.Add(1)
}

// TriggerAndWait marks an operation as done by decrementing the WaitGroup
// counter, and then returns the readiness signal channel for waiting.
func (s *ReadySignal) TriggerAndWait() <-chan struct{} {
	s.wg.Done()
	return s.ch
}

// Intercepted returns true if the signal has been intercepted.
func (s *ReadySignal) Intercepted() bool {
	return s.b.Load()
}

// Intercept marks the signal as intercepted.
func (s *ReadySignal) Intercept() {
	s.b.Store(true)
	s.wg.Done()
}

// Wait blocks until all WaitGroup counters reach zero.
func (s *ReadySignal) Wait() {
	s.wg.Wait()
}

// Close closes the signal channel, notifying all goroutines waiting on it.
func (s *ReadySignal) Close() {
	close(s.ch)
}
