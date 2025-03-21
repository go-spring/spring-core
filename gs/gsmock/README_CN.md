### **Go Mock 框架使用指南**

[英文](README.md)

本 Mock 框架旨在提供 **灵活、可扩展** 的模拟工具，适用于 Go 语言的单元测试。它能够 Mock：

- **基于 `context.Context` 的方法**
- **基于接口的调用**

Mock 使我们能够**隔离被测代码**，避免依赖外部 API、数据库或其他不稳定因素，从而编写**高效、
可靠的单元测试**。

---

## **1️⃣ 两种 Mock 方式**

本框架支持两种 Mock 方式：

1. **侵入式封装**（适用于 `context.Context` 方法）
2. **基于接口的封装**（适用于接口方法）

---

### **🛠 方式 1：侵入式封装（适用于 `context.Context` 方法）**

这种方式允许在不影响原有业务逻辑的前提下，为 `context.Context` 方法提供 Mock 支持。

#### **🔹 Mock 之前的代码**

在未使用 Mock 之前，`Client` 的 `Get` 方法如下：

```go
type Client struct{}

func (c *Client) Get(ctx context.Context, req *Request, trace *Trace) (*Response, error) {
    // 实际业务逻辑，例如调用外部 API
    return &Response{Message: "Real Response"}, nil
}
```

这里 `Get` 方法的行为是直接返回真实的 `Response`，但在测试时，如果它依赖外部服务（如 HTTP 请求
、数据库查询），会导致测试不稳定。因此，我们需要**Mock 机制** 来替换 `Get` 方法，以模拟不同的返回值。

---

#### **🔹 封装后的 `Client`**

```go
var clientType = reflect.TypeFor[Client]()

type Client struct{}

func (c *Client) Get(ctx context.Context, req *Request, trace *Trace) (*Response, error) {
    // 尝试从 Mock 管理器中获取 Mock 结果
    if ret, ok := gsmock.InvokeContext(ctx, clientType, "Get", ctx, req, trace); ok {
        r0, _ := ret[0].(*Response)
        r1, _ := ret[1].(error)
        return r0, r1
    }
    // 没有 Mock 时，执行真实逻辑
    return &Response{Message: "Real Response"}, nil
}
```

Mock 逻辑由 `InvokeContext` 代理，如果 `Get` 方法已注册 Mock，则返回 Mock 结果，否则执行真实逻辑。

---

#### **🔹 注册 Mock**

```go
type GetMocker = gsmock.Mocker32[context.Context, *Request, *Trace, *Response, error]
type GetInvoker = gsmock.Invoker32[context.Context, *Request, *Trace, *Response, error]

// 注册 Mock 逻辑
func MockGet(r *gsmock.Manager) *GetMocker {
    m := &GetMocker{}
    i := &GetInvoker{Mocker32: m}
    r.AddMocker(clientType, "Get", i)
    return m
}
```

- `MockGet` 方法会创建一个 `GetMocker` 并注册到 Mock 管理器中，使 `Get` 方法可以被 Mock。

---

#### **🔹 测试时使用 Mock**

```go
func TestClientGet(t *testing.T) {
    ctx := context.Background()
    mockManager, ctx := gsmock.Init(ctx) // 初始化 Mock 管理器
    mock := MockGet(mockManager) // 注册 Mock 方法
    
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

**✅ 侵入式封装的优势：**

- **适用于 `context.Context` 方法**
- **适用于已有代码，不改变原始 API**
- **Mock 逻辑集中管理，方便测试**

**⚠️ 适用场景：**

- 适用于**已有结构体方法**
- 适用于**包含 `context.Context` 的方法**
- 适用于**无法改成接口的老代码**

---

### **🛠 方式 2：基于接口的封装**

对于基于 **接口** 设计的代码，我们可以直接 Mock 接口，而无需修改原始代码。

#### **🔹 原始接口**

```go
type ClientInterface interface {
    Get(req *Request, trace *Trace) (*Response, error)
}
```

此方法不依赖 `context.Context`，可以通过 **Mock 结构体** 进行模拟。

---

#### **🔹 Mock 结构体**

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

- `MockClient` 作为 `ClientInterface` 的 Mock 实现
- `MockGet()` 方法用于注册 Mock 逻辑
- `Invoke` 方法在 Mock 管理器中查找 Mock 逻辑

---

#### **🔹 测试时使用 Mock**

```go
func TestClientGet_Interface(t *testing.T) {
    mockManager, _ := gsmock.Init(context.Background()) // 初始化 Mock 管理器
    mockClient := NewMockClient(mockManager) // 创建 MockClient
    mock := mockClient.MockGet()             // 注册 Mock 方法
    
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

**✅ 基于接口封装的优势：**

- **完全不侵入原始代码**
- **适用于基于接口的设计**
- **更灵活，支持依赖注入**

**⚠️ 适用场景：**

- 适用于**面向接口编程**
- 适用于**新代码**

---

这样，你可以根据项目需求选择最合适的 Mock 方式，高效完成单元测试！🚀