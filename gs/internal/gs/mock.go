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

package gs

import (
	"reflect"

	"github.com/go-spring/spring-core/gs/gsmock"
)

var mockArgContextType = reflect.TypeFor[MockArgContext]()

type MockArgContext struct {
	r *gsmock.Manager
}

func NewMockArgContext(r *gsmock.Manager) *MockArgContext {
	return &MockArgContext{r: r}
}

func (c *MockArgContext) Check(cond Condition) (bool, error) {
	if r1, r2, ok := gsmock.Invoke2[bool, error](c.r, mockArgContextType, "Check", cond); ok {
		return r1, r2
	}
	panic("mock error")
}

func (c *MockArgContext) MockCheck() *gsmock.Mocker12[Condition, bool, error] {
	m := &gsmock.Mocker12[Condition, bool, error]{}
	i := &gsmock.Invoker12[Condition, bool, error]{Mocker12: m}
	c.r.AddMocker(mockArgContextType, "Check", i)
	return m
}

func (c *MockArgContext) Bind(v reflect.Value, tag string) error {
	if r1, ok := gsmock.Invoke1[error](c.r, mockArgContextType, "Bind", v, tag); ok {
		return r1
	}
	panic("mock error")
}

func (c *MockArgContext) MockBind() *gsmock.Mocker12[reflect.Value, string, error] {
	m := &gsmock.Mocker12[reflect.Value, string, error]{}
	i := &gsmock.Invoker12[reflect.Value, string, error]{Mocker12: m}
	c.r.AddMocker(mockArgContextType, "Bind", i)
	return m
}

func (c *MockArgContext) Wire(v reflect.Value, tag string) error {
	if r1, ok := gsmock.Invoke1[error](c.r, mockArgContextType, "Wire", v, tag); ok {
		return r1
	}
	panic("mock error")
}

func (c *MockArgContext) MockWire() *gsmock.Mocker12[reflect.Value, string, error] {
	m := &gsmock.Mocker12[reflect.Value, string, error]{}
	i := &gsmock.Invoker12[reflect.Value, string, error]{Mocker12: m}
	c.r.AddMocker(mockArgContextType, "Wire", i)
	return m
}
