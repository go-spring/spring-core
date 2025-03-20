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

package gsmock_test

import (
	"context"
	"testing"

	"github.com/go-spring/spring-core/gs/gsmock"
	"github.com/go-spring/spring-core/util/assert"
)

/*********************************** mock ************************************/

// Trace represents a tracing context for tracking requests.
type Trace struct {
	TraceId string
}

// Request represents the input for a mocked function.
type Request struct {
	Token string
}

// Response represents the output for a mocked function.
type Response struct {
	Message string
}

// GetMocker is a mock struct for handling and returning predefined responses.
type GetMocker struct {
	fnHandle func(ctx context.Context, req *Request, trace *Trace) (resp *Response, err error, ok bool)
	fnWhen   func(ctx context.Context, req *Request, trace *Trace) bool
	fnReturn func() (resp *Response, err error)
	retResp  *Response
	retError error
}

// Handle sets a custom function to handle requests.
func (m *GetMocker) Handle(fn func(ctx context.Context, req *Request, trace *Trace) (resp *Response, err error, ok bool)) {
	m.fnHandle = fn
}

// When sets a condition function that determines if the mock should apply.
func (m *GetMocker) When(fn func(ctx context.Context, req *Request, trace *Trace) bool) *GetMocker {
	m.fnWhen = fn
	return m
}

// Return sets a function that returns predefined values.
func (m *GetMocker) Return(fn func() (resp *Response, err error)) {
	m.fnReturn = fn
}

// ReturnValue sets predefined response and error values.
func (m *GetMocker) ReturnValue(resp *Response, err error) {
	m.retResp = resp
	m.retError = err
}

// GetInvoker wraps GetMocker and implements the [gsmock.Invoker] interface.
type GetInvoker struct {
	*GetMocker
}

// Mode determines whether the mock operates in Handle mode or WhenReturn mode.
func (m *GetInvoker) Mode() gsmock.Mode {
	if m.fnHandle != nil {
		return gsmock.ModeHandle
	}
	return gsmock.ModeWhenReturn
}

// Handle executes the custom function if set.
func (m *GetInvoker) Handle(ctx context.Context, params []interface{}) ([]interface{}, bool) {
	r0, r1, ok := m.fnHandle(ctx, params[0].(*Request), params[1].(*Trace))
	return []interface{}{r0, r1}, ok
}

// When checks if the condition function evaluates to true.
func (m *GetInvoker) When(ctx context.Context, params []interface{}) bool {
	if m.fnWhen == nil {
		return false
	}
	return m.fnWhen(ctx, params[0].(*Request), params[1].(*Trace))
}

// Return provides predefined response and error values.
func (m *GetInvoker) Return(ctx context.Context, params []interface{}) []interface{} {
	if m.fnReturn == nil {
		return []interface{}{m.retResp, m.retError}
	}
	r0, r1 := m.fnReturn()
	return []interface{}{r0, r1}
}

// MockGet registers a mocker for a given method name in the mock manager.
func MockGet(r *gsmock.Rooter) *GetMocker {
	m := &GetMocker{}
	i := &GetInvoker{m}
	r.AddMocker("API:Get", i)
	return m
}

// GetWithHeaderMocker extends GetMocker to support additional response headers.
type GetWithHeaderMocker struct {
	fnHandle   func(ctx context.Context, req *Request, trace *Trace) (resp *Response, respHeader map[string]string, err error, ok bool)
	fnWhen     func(ctx context.Context, req *Request, trace *Trace) bool
	fnReturn   func() (resp *Response, respHeader map[string]string, err error)
	retResp    *Response
	respHeader map[string]string
	retError   error
}

// Handle sets a custom function to handle requests.
func (m *GetWithHeaderMocker) Handle(fn func(ctx context.Context, req *Request, trace *Trace) (resp *Response, respHeader map[string]string, err error, ok bool)) {
	m.fnHandle = fn
}

// When sets a condition function that determines if the mock should apply.
func (m *GetWithHeaderMocker) When(fn func(ctx context.Context, req *Request, trace *Trace) bool) *GetWithHeaderMocker {
	m.fnWhen = fn
	return m
}

// Return sets a function that returns predefined values.
func (m *GetWithHeaderMocker) Return(fn func() (resp *Response, respHeader map[string]string, err error)) {
	m.fnReturn = fn
}

// ReturnValue sets predefined response, response headers, and error values.
func (m *GetWithHeaderMocker) ReturnValue(resp *Response, respHeader map[string]string, err error) {
	m.retResp = resp
	m.respHeader = respHeader
	m.retError = err
}

// GetWithHeaderInvoker wraps GetWithHeaderMocker and implements the [gsmock.Invoker] interface.
type GetWithHeaderInvoker struct {
	*GetWithHeaderMocker
}

// Mode determines whether the mock operates in Handle mode or WhenReturn mode.
func (m *GetWithHeaderInvoker) Mode() gsmock.Mode {
	if m.fnHandle != nil {
		return gsmock.ModeHandle
	}
	return gsmock.ModeWhenReturn
}

// Handle executes the custom function if set.
func (m *GetWithHeaderInvoker) Handle(ctx context.Context, params []interface{}) ([]interface{}, bool) {
	r0, r1, r2, ok := m.fnHandle(ctx, params[0].(*Request), params[1].(*Trace))
	return []interface{}{r0, r1, r2}, ok
}

// When checks if the condition function evaluates to true.
func (m *GetWithHeaderInvoker) When(ctx context.Context, params []interface{}) bool {
	if m.fnWhen == nil {
		return false
	}
	return m.fnWhen(ctx, params[0].(*Request), params[1].(*Trace))
}

// Return provides predefined response, response headers, and error values.
func (m *GetWithHeaderInvoker) Return(ctx context.Context, params []interface{}) []interface{} {
	if m.fnReturn == nil {
		return []interface{}{m.retResp, m.respHeader, m.retError}
	}
	r0, r1, r2 := m.fnReturn()
	return []interface{}{r0, r1, r2}
}

// MockGetWithHeader registers a mocker with headers for a given method name in the mock manager.
func MockGetWithHeader(r *gsmock.Rooter) *GetWithHeaderMocker {
	m := &GetWithHeaderMocker{}
	i := &GetWithHeaderInvoker{m}
	r.AddMocker("API:GetWithHeader", i)
	return m
}

/*********************************** test ************************************/

func TestMock(t *testing.T) {
	// Test case: Init function has not been called
	_, ok := gsmock.Invoke("no", context.Background(), nil)
	assert.False(t, ok)

	r, ctx := gsmock.Init(context.Background())

	// Test case: When && ReturnValue
	MockGet(r).
		When(func(ctx context.Context, req *Request, trace *Trace) bool {
			return req.Token == "0:abc"
		}).
		ReturnValue(&Response{Message: "0:abc"}, nil)
	ret, ok := gsmock.Invoke("API:Get", ctx, &Request{Token: "0:abc"}, &Trace{})
	assert.True(t, ok)
	assert.Equal(t, len(r.GetMockers("API:Get")), 1)
	assert.Equal(t, ret[0].(*Response).Message, "0:abc")

	// Test case: When && Return
	MockGet(r).
		When(func(ctx context.Context, req *Request, trace *Trace) bool {
			return req.Token == "1:abc"
		}).
		Return(func() (resp *Response, err error) {
			return &Response{Message: "1:abc"}, nil
		})
	ret, ok = gsmock.Invoke("API:Get", ctx, &Request{Token: "1:abc"}, &Trace{})
	assert.True(t, ok)
	assert.Equal(t, len(r.GetMockers("API:Get")), 2)
	assert.Equal(t, ret[0].(*Response).Message, "1:abc")

	// Test case: When && ReturnValue && WithHeader
	MockGetWithHeader(r).
		When(func(ctx context.Context, req *Request, trace *Trace) bool {
			return req.Token == "2:123"
		}).
		ReturnValue(&Response{Message: "2:123"}, nil, nil)
	ret, ok = gsmock.Invoke("API:GetWithHeader", ctx, &Request{Token: "2:123"}, &Trace{})
	assert.True(t, ok)
	assert.Equal(t, len(r.GetMockers("API:GetWithHeader")), 1)
	assert.Equal(t, ret[0].(*Response).Message, "2:123")

	// Test case: When && Return && WithHeader
	MockGetWithHeader(r).
		When(func(ctx context.Context, req *Request, trace *Trace) bool {
			return req.Token == "3:123"
		}).
		Return(func() (resp *Response, _ map[string]string, err error) {
			return &Response{Message: "3:123"}, nil, nil
		})
	ret, ok = gsmock.Invoke("API:GetWithHeader", ctx, &Request{Token: "3:123"}, &Trace{})
	assert.True(t, ok)
	assert.Equal(t, len(r.GetMockers("API:GetWithHeader")), 2)
	assert.Equal(t, ret[0].(*Response).Message, "3:123")

	// Test case: Handle
	MockGet(r).
		Handle(func(ctx context.Context, req *Request, trace *Trace) (resp *Response, err error, ok bool) {
			return &Response{Message: "4:xyz"}, nil, req.Token == "4:xyz"
		})
	ret, ok = gsmock.Invoke("API:Get", ctx, &Request{Token: "4:xyz"}, &Trace{})
	assert.True(t, ok)
	assert.Equal(t, len(r.GetMockers("API:Get")), 3)
	assert.Equal(t, ret[0].(*Response).Message, "4:xyz")

	// Test case: Handle && WithHeader
	MockGetWithHeader(r).
		Handle(func(ctx context.Context, req *Request, trace *Trace) (resp *Response, respHeader map[string]string, err error, ok bool) {
			return &Response{Message: "5:890"}, nil, nil, req.Token == "5:890"
		})
	ret, ok = gsmock.Invoke("API:GetWithHeader", ctx, &Request{Token: "5:890"}, &Trace{})
	assert.True(t, ok)
	assert.Equal(t, len(r.GetMockers("API:GetWithHeader")), 3)
	assert.Equal(t, ret[0].(*Response).Message, "5:890")

	// Test invalid case: When && ReturnValue
	MockGet(r).
		When(nil).
		ReturnValue(nil, nil)
	_, ok = gsmock.Invoke("API:Get", ctx, &Request{}, &Trace{})
	assert.False(t, ok)
	assert.Equal(t, len(r.GetMockers("API:Get")), 4)

	// Test invalid case: When && ReturnValue && WithHeader
	MockGetWithHeader(r).
		When(nil).
		ReturnValue(nil, nil, nil)
	_, ok = gsmock.Invoke("API:GetWithHeader", ctx, &Request{}, &Trace{})
	assert.False(t, ok)
	assert.Equal(t, len(r.GetMockers("API:GetWithHeader")), 4)

	// Test invalid case: Handle
	MockGet(r).Handle(nil)
	_, ok = gsmock.Invoke("API:Get", ctx, &Request{}, &Trace{})
	assert.False(t, ok)
	assert.Equal(t, len(r.GetMockers("API:Get")), 5)

	// Test invalid case: Handle && WithHeader
	MockGetWithHeader(r).Handle(nil)
	_, ok = gsmock.Invoke("API:GetWithHeader", ctx, &Request{}, &Trace{})
	assert.False(t, ok)
	assert.Equal(t, len(r.GetMockers("API:GetWithHeader")), 5)
}
