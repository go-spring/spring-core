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

/******************************** mocker22 ***********************************/

// Mocker22 is a mock implementation for a function with two parameters and two return values.
type Mocker22[T1, T2 any, R1, R2 any] struct {
	fnHandle func(T1, T2) (R1, R2, bool)
	fnWhen   func(T1, T2) bool
	fnReturn func() (R1, R2)
	retR1    R1
	retR2    R2
}

// Handle sets a custom function to handle requests.
func (m *Mocker22[T1, T2, R1, R2]) Handle(fn func(T1, T2) (R1, R2, bool)) {
	m.fnHandle = fn
}

// When sets a condition function that determines if the mock should apply.
func (m *Mocker22[T1, T2, R1, R2]) When(fn func(T1, T2) bool) *Mocker22[T1, T2, R1, R2] {
	m.fnWhen = fn
	return m
}

// Return sets a function that returns predefined values.
func (m *Mocker22[T1, T2, R1, R2]) Return(fn func() (R1, R2)) {
	m.fnReturn = fn
}

// ReturnValue sets predefined response and error values.
func (m *Mocker22[T1, T2, R1, R2]) ReturnValue(r1 R1, r2 R2) {
	m.retR1 = r1
	m.retR2 = r2
}

// Invoker22 is an Invoker implementation for Mocker22.
type Invoker22[T1, T2 any, R1, R2 any] struct {
	*Mocker22[T1, T2, R1, R2]
}

// Mode determines whether the mock operates in Handle mode or WhenReturn mode.
func (m *Invoker22[T1, T2, R1, R2]) Mode() Mode {
	if m.fnHandle != nil {
		return ModeHandle
	}
	return ModeWhenReturn
}

// Handle executes the custom function if set.
func (m *Invoker22[T1, T2, R1, R2]) Handle(params []interface{}) ([]interface{}, bool) {
	r0, r1, ok := m.fnHandle(params[0].(T1), params[1].(T2))
	return []interface{}{r0, r1}, ok
}

// When checks if the condition function evaluates to true.
func (m *Invoker22[T1, T2, R1, R2]) When(params []interface{}) bool {
	if m.fnWhen == nil {
		return false
	}
	return m.fnWhen(params[0].(T1), params[1].(T2))
}

// Return provides predefined response and error values.
func (m *Invoker22[T1, T2, R1, R2]) Return(params []interface{}) []interface{} {
	if m.fnReturn == nil {
		return []interface{}{m.retR1, m.retR2}
	}
	r0, r1 := m.fnReturn()
	return []interface{}{r0, r1}
}

/******************************** mocker23 ***********************************/

// Mocker23 is a mock implementation for a function with two parameters and three return values.
type Mocker23[T1, T2 any, R1, R2, R3 any] struct {
	fnHandle func(T1, T2) (R1, R2, R3, bool)
	fnWhen   func(T1, T2) bool
	fnReturn func() (R1, R2, R3)
	retR1    R1
	retR2    R2
	retR3    R3
}

// Handle sets a custom function to handle requests.
func (m *Mocker23[T1, T2, R1, R2, R3]) Handle(fn func(T1, T2) (R1, R2, R3, bool)) {
	m.fnHandle = fn
}

// When sets a condition function that determines if the mock should apply.
func (m *Mocker23[T1, T2, R1, R2, R3]) When(fn func(T1, T2) bool) *Mocker23[T1, T2, R1, R2, R3] {
	m.fnWhen = fn
	return m
}

// Return sets a function that returns predefined values.
func (m *Mocker23[T1, T2, R1, R2, R3]) Return(fn func() (R1, R2, R3)) {
	m.fnReturn = fn
}

// ReturnValue sets predefined response and error values.
func (m *Mocker23[T1, T2, R1, R2, R3]) ReturnValue(r1 R1, r2 R2, r3 R3) {
	m.retR1 = r1
	m.retR2 = r2
	m.retR3 = r3
}

// Invoker23 is an Invoker implementation for Mocker23.
type Invoker23[T1, T2 any, R1, R2, R3 any] struct {
	*Mocker23[T1, T2, R1, R2, R3]
}

// Mode determines whether the mock operates in Handle mode or WhenReturn mode.
func (m *Invoker23[T1, T2, R1, R2, R3]) Mode() Mode {
	if m.fnHandle != nil {
		return ModeHandle
	}
	return ModeWhenReturn
}

// Handle executes the custom function if set.
func (m *Invoker23[T1, T2, R1, R2, R3]) Handle(params []interface{}) ([]interface{}, bool) {
	r0, r1, r2, ok := m.fnHandle(params[0].(T1), params[1].(T2))
	return []interface{}{r0, r1, r2}, ok
}

// When checks if the condition function evaluates to true.
func (m *Invoker23[T1, T2, R1, R2, R3]) When(params []interface{}) bool {
	if m.fnWhen == nil {
		return false
	}
	return m.fnWhen(params[0].(T1), params[1].(T2))
}

// Return provides predefined response and error values.
func (m *Invoker23[T1, T2, R1, R2, R3]) Return(params []interface{}) []interface{} {
	if m.fnReturn == nil {
		return []interface{}{m.retR1, m.retR2, m.retR3}
	}
	r0, r1, r2 := m.fnReturn()
	return []interface{}{r0, r1, r2}
}

/******************************** mocker32 ***********************************/

// Mocker32 is a mock implementation for a function with three parameters and two return values.
type Mocker32[T1, T2, T3 any, R1, R2 any] struct {
	fnHandle func(T1, T2, T3) (R1, R2, bool)
	fnWhen   func(T1, T2, T3) bool
	fnReturn func() (R1, R2)
	retR1    R1
	retR2    R2
}

// Handle sets a custom function to handle requests.
func (m *Mocker32[T1, T2, T3, R1, R2]) Handle(fn func(T1, T2, T3) (R1, R2, bool)) {
	m.fnHandle = fn
}

// When sets a condition function that determines if the mock should apply.
func (m *Mocker32[T1, T2, T3, R1, R2]) When(fn func(T1, T2, T3) bool) *Mocker32[T1, T2, T3, R1, R2] {
	m.fnWhen = fn
	return m
}

// Return sets a function that returns predefined values.
func (m *Mocker32[T1, T2, T3, R1, R2]) Return(fn func() (R1, R2)) {
	m.fnReturn = fn
}

// ReturnValue sets predefined response and error values.
func (m *Mocker32[T1, T2, T3, R1, R2]) ReturnValue(r1 R1, r2 R2) {
	m.retR1 = r1
	m.retR2 = r2
}

// Invoker32 is an Invoker implementation for Mocker32.
type Invoker32[T1, T2, T3 any, R1, R2 any] struct {
	*Mocker32[T1, T2, T3, R1, R2]
}

// Mode determines whether the mock operates in Handle mode or WhenReturn mode.
func (m *Invoker32[T1, T2, T3, R1, R2]) Mode() Mode {
	if m.fnHandle != nil {
		return ModeHandle
	}
	return ModeWhenReturn
}

// Handle executes the custom function if set.
func (m *Invoker32[T1, T2, T3, R1, R2]) Handle(params []interface{}) ([]interface{}, bool) {
	r0, r1, ok := m.fnHandle(params[0].(T1), params[1].(T2), params[2].(T3))
	return []interface{}{r0, r1}, ok
}

// When checks if the condition function evaluates to true.
func (m *Invoker32[T1, T2, T3, R1, R2]) When(params []interface{}) bool {
	if m.fnWhen == nil {
		return false
	}
	return m.fnWhen(params[0].(T1), params[1].(T2), params[2].(T3))
}

// Return provides predefined response and error values.
func (m *Invoker32[T1, T2, T3, R1, R2]) Return(params []interface{}) []interface{} {
	if m.fnReturn == nil {
		return []interface{}{m.retR1, m.retR2}
	}
	r0, r1 := m.fnReturn()
	return []interface{}{r0, r1}
}

/******************************** mocker33 ***********************************/

// Mocker33 is a mock implementation for a function with three parameters and three return values.
type Mocker33[T1, T2, T3 any, R1, R2, R3 any] struct {
	fnHandle func(T1, T2, T3) (R1, R2, R3, bool)
	fnWhen   func(T1, T2, T3) bool
	fnReturn func() (R1, R2, R3)
	retR1    R1
	retR2    R2
	retR3    R3
}

// Handle sets a custom function to handle requests.
func (m *Mocker33[T1, T2, T3, R1, R2, R3]) Handle(fn func(T1, T2, T3) (R1, R2, R3, bool)) {
	m.fnHandle = fn
}

// When sets a condition function that determines if the mock should apply.
func (m *Mocker33[T1, T2, T3, R1, R2, R3]) When(fn func(T1, T2, T3) bool) *Mocker33[T1, T2, T3, R1, R2, R3] {
	m.fnWhen = fn
	return m
}

// Return sets a function that returns predefined values.
func (m *Mocker33[T1, T2, T3, R1, R2, R3]) Return(fn func() (R1, R2, R3)) {
	m.fnReturn = fn
}

// ReturnValue sets predefined response and error values.
func (m *Mocker33[T1, T2, T3, R1, R2, R3]) ReturnValue(r1 R1, r2 R2, r3 R3) {
	m.retR1 = r1
	m.retR2 = r2
	m.retR3 = r3
}

// Invoker33 is an Invoker implementation for Mocker33.
type Invoker33[T1, T2, T3 any, R1, R2, R3 any] struct {
	*Mocker33[T1, T2, T3, R1, R2, R3]
}

// Mode determines whether the mock operates in Handle mode or WhenReturn mode.
func (m *Invoker33[T1, T2, T3, R1, R2, R3]) Mode() Mode {
	if m.fnHandle != nil {
		return ModeHandle
	}
	return ModeWhenReturn
}

// Handle executes the custom function if set.
func (m *Invoker33[T1, T2, T3, R1, R2, R3]) Handle(params []interface{}) ([]interface{}, bool) {
	r0, r1, r2, ok := m.fnHandle(params[0].(T1), params[1].(T2), params[2].(T3))
	return []interface{}{r0, r1, r2}, ok
}

// When checks if the condition function evaluates to true.
func (m *Invoker33[T1, T2, T3, R1, R2, R3]) When(params []interface{}) bool {
	if m.fnWhen == nil {
		return false
	}
	return m.fnWhen(params[0].(T1), params[1].(T2), params[2].(T3))
}

// Return provides predefined response and error values.
func (m *Invoker33[T1, T2, T3, R1, R2, R3]) Return(params []interface{}) []interface{} {
	if m.fnReturn == nil {
		return []interface{}{m.retR1, m.retR2, m.retR3}
	}
	r0, r1, r2 := m.fnReturn()
	return []interface{}{r0, r1, r2}
}
