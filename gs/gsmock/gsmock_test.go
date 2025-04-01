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
	"reflect"
	"testing"

	"github.com/go-spring/spring-core/gs/gsmock"
	"github.com/go-spring/spring-core/util/assert"
)

/*********************************** mock ************************************/

var clientType = reflect.TypeFor[Client]()

type Client struct{}

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

// Get performs a request and retrieves a response, potentially using a mock implementation.
func (c *Client) Get(ctx context.Context, req *Request, trace *Trace) (*Response, error) {
	if ret, ok := gsmock.InvokeContext(ctx, clientType, "Get", ctx, req, trace); ok {
		return gsmock.Unbox2[*Response, error](ret)
	}
	return &Response{Message: "9:xxx"}, nil
}

// MockGet registers a mock implementation for the Get method.
func MockGet(r *gsmock.Manager) *gsmock.Mocker32[context.Context, *Request, *Trace, *Response, error] {
	return gsmock.NewMocker32[context.Context, *Request, *Trace, *Response, error](r, clientType, "Get")
}

// GetWithHeader performs a request and retrieves a response with additional headers, potentially using a mock implementation.
func (c *Client) GetWithHeader(ctx context.Context, req *Request, trace *Trace) (*Response, map[string]string, error) {
	if ret, ok := gsmock.InvokeContext(ctx, clientType, "GetWithHeader", ctx, req, trace); ok {
		return gsmock.Unbox3[*Response, map[string]string, error](ret)
	}
	return &Response{Message: "9:yyy"}, nil, nil
}

// MockGetWithHeader registers a mock implementation for the GetWithHeader method.
func MockGetWithHeader(r *gsmock.Manager) *gsmock.Mocker33[context.Context, *Request, *Trace, *Response, map[string]string, error] {
	return gsmock.NewMocker33[context.Context, *Request, *Trace, *Response, map[string]string, error](r, clientType, "GetWithHeader")
}

/*********************************** test ************************************/

func TestMockWithContext(t *testing.T) {
	var c Client

	// Test case: Init function has not been called
	{
		ctx := context.Background()
		resp, err := c.Get(ctx, &Request{}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "9:xxx")
	}

	r, ctx := gsmock.Init(context.Background())

	// Test case: When && ReturnValue
	{
		MockGet(r).
			When(func(ctx context.Context, req *Request, trace *Trace) bool {
				return req.Token == "0:abc"
			}).
			ReturnValue(&Response{Message: "0:abc"}, nil)

		resp, err := c.Get(ctx, &Request{Token: "0:abc"}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "0:abc")
		assert.Equal(t, len(r.GetMockers(clientType, "Get")), 1)
	}

	// Test case: When && Return
	{
		MockGet(r).
			When(func(ctx context.Context, req *Request, trace *Trace) bool {
				return req.Token == "1:abc"
			}).
			Return(func() (resp *Response, err error) {
				return &Response{Message: "1:abc"}, nil
			})

		resp, err := c.Get(ctx, &Request{Token: "1:abc"}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "1:abc")
		assert.Equal(t, len(r.GetMockers(clientType, "Get")), 2)
	}

	// Test case: When && ReturnValue && WithHeader
	{
		MockGetWithHeader(r).
			When(func(ctx context.Context, req *Request, trace *Trace) bool {
				return req.Token == "2:123"
			}).
			ReturnValue(&Response{Message: "2:123"}, nil, nil)

		resp, _, err := c.GetWithHeader(ctx, &Request{Token: "2:123"}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "2:123")
		assert.Equal(t, len(r.GetMockers(clientType, "GetWithHeader")), 1)
	}

	// Test case: When && Return && WithHeader
	{
		MockGetWithHeader(r).
			When(func(ctx context.Context, req *Request, trace *Trace) bool {
				return req.Token == "3:123"
			}).
			Return(func() (resp *Response, _ map[string]string, err error) {
				return &Response{Message: "3:123"}, nil, nil
			})

		resp, _, err := c.GetWithHeader(ctx, &Request{Token: "3:123"}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "3:123")
		assert.Equal(t, len(r.GetMockers(clientType, "GetWithHeader")), 2)
	}

	// Test case: Handle
	{
		MockGet(r).
			Handle(func(ctx context.Context, req *Request, trace *Trace) (resp *Response, err error, ok bool) {
				return &Response{Message: "4:xyz"}, nil, req.Token == "4:xyz"
			})

		resp, err := c.Get(ctx, &Request{Token: "4:xyz"}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "4:xyz")
		assert.Equal(t, len(r.GetMockers(clientType, "Get")), 3)
	}

	// Test case: Handle && WithHeader
	{
		MockGetWithHeader(r).
			Handle(func(ctx context.Context, req *Request, trace *Trace) (resp *Response, respHeader map[string]string, err error, ok bool) {
				return &Response{Message: "5:890"}, nil, nil, req.Token == "5:890"
			})

		resp, _, err := c.GetWithHeader(ctx, &Request{Token: "5:890"}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "5:890")
		assert.Equal(t, len(r.GetMockers(clientType, "GetWithHeader")), 3)
	}

	// Test invalid case: When && ReturnValue
	{
		MockGet(r).
			When(nil).
			ReturnValue(nil, nil)

		resp, err := c.Get(ctx, &Request{}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "9:xxx")
		assert.Equal(t, len(r.GetMockers(clientType, "Get")), 4)
	}

	// Test invalid case: When && ReturnValue && WithHeader
	{
		MockGetWithHeader(r).
			When(nil).
			ReturnValue(nil, nil, nil)

		resp, _, err := c.GetWithHeader(ctx, &Request{}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "9:yyy")
		assert.Equal(t, len(r.GetMockers(clientType, "GetWithHeader")), 4)
	}

	// Test invalid case: Handle
	{
		MockGet(r).Handle(nil)

		resp, err := c.Get(ctx, &Request{}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "9:xxx")
		assert.Equal(t, len(r.GetMockers(clientType, "Get")), 5)
	}

	// Test invalid case: Handle && WithHeader
	{
		MockGetWithHeader(r).Handle(nil)

		resp, _, err := c.GetWithHeader(ctx, &Request{}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "9:yyy")
		assert.Equal(t, len(r.GetMockers(clientType, "GetWithHeader")), 5)
	}
}

/*********************************** test ************************************/

var mockClientType = reflect.TypeFor[MockClient]()

// ClientInterface defines the expected behavior for a client.
type ClientInterface interface {
	Query(req *Request, trace *Trace) (*Response, error)
	QueryWithHeader(req *Request, trace *Trace) (*Response, map[string]string, error)
}

// MockClient is a mock implementation of ClientInterface.
type MockClient struct {
	r *gsmock.Manager
}

// NewMockClient creates a new instance of MockClient.
func NewMockClient(r *gsmock.Manager) *MockClient {
	return &MockClient{r}
}

// Query mocks the Query method by invoking a registered mock implementation.
func (c *MockClient) Query(req *Request, trace *Trace) (*Response, error) {
	if ret, ok := gsmock.Invoke(c.r, mockClientType, "Query", req, trace); ok {
		return gsmock.Unbox2[*Response, error](ret)
	}
	panic("mock error")
}

// MockQuery registers a mock implementation for the Query method.
func (c *MockClient) MockQuery() *gsmock.Mocker22[*Request, *Trace, *Response, error] {
	return gsmock.NewMocker22[*Request, *Trace, *Response, error](c.r, mockClientType, "Query")
}

// QueryWithHeader mocks the QueryWithHeader method by invoking a registered mock implementation.
func (c *MockClient) QueryWithHeader(req *Request, trace *Trace) (*Response, map[string]string, error) {
	if ret, ok := gsmock.Invoke(c.r, mockClientType, "QueryWithHeader", req, trace); ok {
		return gsmock.Unbox3[*Response, map[string]string, error](ret)
	}
	panic("mock error")
}

// MockQueryWithHeader registers a mock implementation for the QueryWithHeader method.
func (c *MockClient) MockQueryWithHeader() *gsmock.Mocker23[*Request, *Trace, *Response, map[string]string, error] {
	return gsmock.NewMocker23[*Request, *Trace, *Response, map[string]string, error](c.r, mockClientType, "QueryWithHeader")
}

func TestMockNoContext(t *testing.T) {

	r, _ := gsmock.Init(context.Background())

	var c ClientInterface
	mc := NewMockClient(r)
	c = mc

	// Test case: When && ReturnValue
	{
		mc.MockQuery().
			When(func(req *Request, trace *Trace) bool {
				return req.Token == "0:abc"
			}).
			ReturnValue(&Response{Message: "0:abc"}, nil)

		resp, err := c.Query(&Request{Token: "0:abc"}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "0:abc")
		assert.Equal(t, len(r.GetMockers(mockClientType, "Query")), 1)
	}

	// Test case: When && Return
	{
		mc.MockQuery().
			When(func(req *Request, trace *Trace) bool {
				return req.Token == "1:abc"
			}).
			Return(func() (resp *Response, err error) {
				return &Response{Message: "1:abc"}, nil
			})

		resp, err := c.Query(&Request{Token: "1:abc"}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "1:abc")
		assert.Equal(t, len(r.GetMockers(mockClientType, "Query")), 2)
	}

	// Test case: When && ReturnValue && WithHeader
	{
		mc.MockQueryWithHeader().
			When(func(req *Request, trace *Trace) bool {
				return req.Token == "2:123"
			}).
			ReturnValue(&Response{Message: "2:123"}, nil, nil)

		resp, _, err := c.QueryWithHeader(&Request{Token: "2:123"}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "2:123")
		assert.Equal(t, len(r.GetMockers(mockClientType, "QueryWithHeader")), 1)
	}

	// Test case: When && Return && WithHeader
	{
		mc.MockQueryWithHeader().
			When(func(req *Request, trace *Trace) bool {
				return req.Token == "3:123"
			}).
			Return(func() (resp *Response, _ map[string]string, err error) {
				return &Response{Message: "3:123"}, nil, nil
			})

		resp, _, err := c.QueryWithHeader(&Request{Token: "3:123"}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "3:123")
		assert.Equal(t, len(r.GetMockers(mockClientType, "QueryWithHeader")), 2)
	}

	// Test case: Handle
	{
		mc.MockQuery().
			Handle(func(req *Request, trace *Trace) (resp *Response, err error, ok bool) {
				return &Response{Message: "4:xyz"}, nil, req.Token == "4:xyz"
			})

		resp, err := c.Query(&Request{Token: "4:xyz"}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "4:xyz")
		assert.Equal(t, len(r.GetMockers(mockClientType, "Query")), 3)
	}

	// Test case: Handle && WithHeader
	{
		mc.MockQueryWithHeader().
			Handle(func(req *Request, trace *Trace) (resp *Response, respHeader map[string]string, err error, ok bool) {
				return &Response{Message: "5:890"}, nil, nil, req.Token == "5:890"
			})

		resp, _, err := c.QueryWithHeader(&Request{Token: "5:890"}, &Trace{})
		assert.Nil(t, err)
		assert.Equal(t, resp.Message, "5:890")
		assert.Equal(t, len(r.GetMockers(mockClientType, "QueryWithHeader")), 3)
	}

	// Test invalid case: When && ReturnValue
	{
		mc.MockQuery().
			When(nil).
			ReturnValue(nil, nil)

		assert.Panic(t, func() {
			_, _ = c.Query(&Request{}, &Trace{})
		}, "mock error")
	}

	// Test invalid case: When && ReturnValue && WithHeader
	{
		mc.MockQueryWithHeader().
			When(nil).
			ReturnValue(nil, nil, nil)

		assert.Panic(t, func() {
			_, _, _ = c.QueryWithHeader(&Request{}, &Trace{})
		}, "mock error")
	}

	// Test invalid case: Handle
	{
		mc.MockQuery().Handle(nil)

		assert.Panic(t, func() {
			_, _ = c.Query(&Request{}, &Trace{})
		}, "mock error")
	}

	// Test invalid case: Handle && WithHeader
	{
		mc.MockQueryWithHeader().Handle(nil)

		assert.Panic(t, func() {
			_, _, _ = c.QueryWithHeader(&Request{}, &Trace{})
		}, "mock error")
	}
}
