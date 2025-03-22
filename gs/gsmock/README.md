# **Go Mock Framework Usage Guide**

[中文](README_CN.md)[]()

This mock framework is designed to provide a **flexible and extensible** mocking tool
for unit testing in Go. It supports mocking:

- **Methods that use `context.Context`**
- **Interface-based calls**

Mocking helps us **isolate the code under test**, avoiding dependencies on external APIs,
databases, or other unstable factors, thereby enabling **efficient and reliable unit testing**.

---

## **1️⃣ Two Mocking Approaches**

This framework supports two mocking approaches:

1. **Intrusive Wrapping** (for methods using `context.Context`)
2. **Interface-Based Wrapping** (for interface methods)

---

### **🛠 Approach 1: Intrusive Wrapping (for `context.Context` methods)**

This approach allows us to introduce mocking without modifying existing business logic.

#### **🔹 Before Mocking**

Before using mocks, the `Get` method of `Client` looks like this:

```go
type Client struct{}

func (c *Client) Get(ctx context.Context, req *Request, trace *Trace) (*Response, error) {
    // Real business logic, such as calling an external API
    return &Response{Message: "Real Response"}, nil
}
```

Here, `Get` directly returns a real `Response`. However, in testing, if it relies on
external services (such as HTTP requests or database queries), it can make tests unstable.
Therefore, we need **mocking** to replace `Get` with controlled behavior.

---

#### **🔹 Wrapped `Client` for Mocking**

```go
var clientType = reflect.TypeFor[Client]()

type Client struct{}

func (c *Client) Get(ctx context.Context, req *Request, trace *Trace) (*Response, error) {
    // Attempt to retrieve a mock result from the mock manager
    if ret, ok := gsmock.InvokeContext(ctx, clientType, "Get", ctx, req, trace); ok {
        r0, _ := ret[0].(*Response)
        r1, _ := ret[1].(error)
        return r0, r1
    }
    // If no mock is available, execute the real logic
    return &Response{Message: "Real Response"}, nil
}
```

The `InvokeContext` function handles mock logic. If a mock is registered, it returns the
mocked result; otherwise, it executes the real logic.

---

#### **🔹 Registering a Mock**

```go
type GetMocker = gsmock.Mocker32[context.Context, *Request, *Trace, *Response, error]
type GetInvoker = gsmock.Invoker32[context.Context, *Request, *Trace, *Response, error]

// Register a mock implementation
func MockGet(r *gsmock.Manager) *GetMocker {
    m := &GetMocker{}
    i := &GetInvoker{Mocker32: m}
    r.AddMocker(clientType, "Get", i)
    return m
}
```

- The `MockGet` function creates a `GetMocker` and registers it in the mock manager,
  allowing `Get` to be mocked.

---

#### **🔹 Using the Mock in Testing**

```go
func TestClientGet(t *testing.T) {
    ctx := context.Background()
    mockManager, ctx := gsmock.Init(ctx) // Initialize the mock manager
    mock := MockGet(mockManager) // Register the mock method
    
    mock.When(func (ctx context.Context, req *Request, trace *Trace) bool {
        return req.Token == "test-token"
    }).Return(func (ctx context.Context, req *Request, trace *Trace) (*Response, error) {
        return &Response{Message: "Mocked Response"}, nil
    })
    
    client := &Client{}
    resp, err := client.Get(ctx, &Request{Token: "test-token"}, &Trace{})
    
    if err != nil || resp.Message != "Mocked Response" {
        t.Errorf("Unexpected response: %+v, error: %v", resp, err)
    }
}
```

**✅ Advantages of Intrusive Wrapping:**

- **Works with `context.Context` methods**
- **Can be applied to existing code without altering APIs**
- **Centralized mock logic for easier testing**

**⚠️ Suitable for:**

- **Existing struct methods**
- **Methods using `context.Context`**
- **Legacy code that cannot be changed into interfaces**

---

### **🛠 Approach 2: Interface-Based Wrapping**

For code designed with **interfaces**, we can mock the interface directly without
modifying the original implementation.

#### **🔹 Original Interface**

```go
type ClientInterface interface {
    Get(req *Request, trace *Trace) (*Response, error)
}
```

Since this method does not use `context.Context`, we can mock it with a **mock struct**.

---

#### **🔹 Mocking the Interface**

```go
var mockClientType = reflect.TypeFor[MockClient]()

type MockClient struct {
    r *gsmock.Manager
}

func NewMockClient(r *gsmock.Manager) *MockClient {
    return &MockClient{r}
}

func (c *MockClient) Get(req *Request, trace *Trace) (*Response, error) {
    if ret, ok := gsmock.Invoke(c.r, mockClientType, "Get", req, trace); ok {
        r0, _ := ret[0].(*Response)
        r1, _ := ret[1].(error)
        return r0, r1
    }
    panic("mock error")
}

type GetMocker = gsmock.Mocker22[*Request, *Trace, *Response, error]
type GetInvoker = gsmock.Invoker22[*Request, *Trace, *Response, error]

func (c *MockClient) MockGet() *GetMocker {
    m := &GetMocker{}
    i := &GetInvoker{Mocker22: m}
    c.r.AddMocker(mockClientType, "Get", i)
    return m
}
```

- `MockClient` implements `ClientInterface` as a mock.
- `MockGet()` registers the mock logic.
- The `Invoke` method looks up mock behavior in the mock manager.

---

#### **🔹 Using the Mock in Testing**

```go
func TestClientGet_Interface(t *testing.T) {
    mockManager, _ := gsmock.Init(context.Background()) // Initialize the mock manager
    mockClient := NewMockClient(mockManager) // Create a MockClient
    mock := mockClient.MockGet()             // Register the mock method
    
    mock.When(func (req *Request, trace *Trace) bool {
        return req.Token == "test-token"
    }).Return(func (req *Request, trace *Trace) (*Response, error) {
		return &Response{Message: "Mocked Response"}, nil
    })
    
    client := ClientInterface(mockClient)
    resp, err := client.Get(&Request{Token: "test-token"}, &Trace{})
    
    if err != nil || resp.Message != "Mocked Response" {
        t.Errorf("Unexpected response: %+v, error: %v", resp, err)
    }
}
```

**✅ Advantages of Interface-Based Wrapping:**

- **Does not modify the original code**
- **Works well with interface-driven design**
- **More flexible and supports dependency injection**

**⚠️ Suitable for:**

- **Code using interfaces**
- **New projects with clean architecture**

---

By choosing the most appropriate mocking approach based on project needs, you can
effectively perform unit testing! 🚀