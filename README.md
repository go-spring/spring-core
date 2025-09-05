# Go-Spring

<div>
   <img src="https://img.shields.io/github/license/go-spring/spring-core" alt="license"/>
   <img src="https://img.shields.io/github/go-mod/go-version/go-spring/spring-core" alt="go-version"/>
   <img src="https://img.shields.io/github/v/release/go-spring/spring-core?include_prereleases" alt="release"/>
   <a href="https://codecov.io/gh/go-spring/spring-core" > 
      <img src="https://codecov.io/gh/go-spring/spring-core/graph/badge.svg?token=SX7CV1T0O8" alt="test-coverage"/>
   </a>
   <a href="https://deepwiki.com/go-spring/spring-core"><img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki"></a>
</div>

[English](README.md) | [中文](README_CN.md)

**Go-Spring is a high-performance framework for modern Go application development, inspired by the Spring / Spring Boot
ecosystem in the Java community.**
Its design philosophy deeply integrates the characteristics of the Go language, retaining mature development paradigms
from the Spring world, such as Dependency Injection (DI), auto-configuration, and lifecycle management,
while avoiding the complexity and performance overhead that traditional frameworks might bring.
Go-Spring allows developers to enjoy higher levels of abstraction and automation while maintaining Go's native style and
execution efficiency.

**Whether you are developing a monolithic application or building a microservices-based distributed system, Go-Spring
provides a unified and flexible development experience.**
It simplifies the project setup process in an "out-of-the-box" manner, reduces the need for boilerplate code, and does
not impose an intrusive framework structure, allowing developers to focus more on business logic implementation.
Go-Spring is committed to improving development efficiency, maintainability, and system consistency, making it a
milestone framework in the Go language ecosystem.

## 🚀 Feature Overview

Go-Spring offers a rich set of practical features to help developers efficiently build modern Go applications:

1. ⚡ **Extreme Startup Performance**
    - Bean registration based on Go's `init()` mechanism, with no runtime scanning, ensuring rapid startup;
    - Injection relies only on reflection during the initialization phase, with zero reflection at runtime, maximizing
      performance.

2. 🧩 **Out-of-the-Box, Non-Intrusive Design**
    - Supports struct tag injection and chained configuration, making it easy to use without mastering complex concepts;
    - Does not strongly depend on interfaces or inheritance structures, keeping business logic in native Go style.

3. 🔄 **Hot Configuration Updates, Real-Time Application**
    - Supports loading configurations from multiple formats and sources, with environment isolation and dynamic refresh
      capabilities;
    - Configuration changes can be applied immediately, facilitating gray releases and online parameter tuning.

4. ⚙️ **Flexible Dependency Injection Mechanism**
    - Supports constructor injection, struct field injection, and constructor parameter injection in various ways;
    - Injection behavior can be flexibly adjusted based on configuration items or runtime environments.

5. 🔌 **Multi-Model Service Startup Support**
    - Built-in HTTP Server launcher for quick deployment of web services;
    - Supports three running models: `Runner`, `Job`, and `Server`, adapting to different service forms;
    - Comprehensive lifecycle hooks support graceful startup and shutdown.

6. 🧪 **Built-In Testing Capabilities**
    - Seamlessly integrates with `go test`, supports Bean Mock and dependency injection, making it easy to write unit
      tests.

## 📦 Installation

Go-Spring uses Go Modules for dependency management, making installation straightforward:

```bash
go get github.com/go-spring/spring-core
```

## 🚀 Quick Start

Go-Spring emphasizes "out-of-the-box" usage. Below are two examples to quickly experience its powerful capabilities.

> More examples can be found at: [gs/examples](gs/examples)

### Example 1: Minimal API Service

```go
func main() {
   http.HandleFunc("/echo", func (w http.ResponseWriter, r *http.Request) {
      w.Write([]byte("hello world!"))
   })
   gs.Run()
}
```

Access method:

```bash
curl http://127.0.0.1:9090/echo
# Output: hello world!
```

✅ No complex configuration required; the Go standard library `http` can be used directly;  
✅ `gs.Run()` manages the lifecycle, supporting graceful exit, signal listening, and other capabilities.

### Example 2: Basic Feature Demonstration

```go
func init() {
   gs.Object(&Service{})
   
   gs.Provide(func (s *Service) *http.ServeMux {
      http.HandleFunc("/echo", s.Echo)
      http.HandleFunc("/refresh", s.Refresh)
      return http.DefaultServeMux
   })
   
   sysconf.Set("start-time", time.Now().Format(timeLayout))
   sysconf.Set("refresh-time", time.Now().Format(timeLayout))
}
```

```go
const timeLayout = "2006-01-02 15:04:05.999 -0700 MST"

type Service struct {
   StartTime   time.Time          `value:"${start-time}"`
   RefreshTime gs.Dync[time.Time] `value:"${refresh-time}"`
}

func (s *Service) Echo(w http.ResponseWriter, r *http.Request) {
   str := fmt.Sprintf("start-time: %s refresh-time: %s",
      s.StartTime.Format(timeLayout),
      s.RefreshTime.Value().Format(timeLayout))
   w.Write([]byte(str))
}

func (s *Service) Refresh(w http.ResponseWriter, r *http.Request) {
   sysconf.Set("refresh-time", time.Now().Format(timeLayout))
   gs.RefreshProperties()
   w.Write([]byte("OK!"))
}
```

Access method:

```bash
curl http://127.0.0.1:9090/echo     # View current time
curl http://127.0.0.1:9090/refresh  # Trigger hot refresh
```

✅ `value` tag automatically binds configuration;  
✅ `gs.Dync[T]` implements field hot updates;  
✅ `gs.Object` and `gs.Provide()` register Beans.

## 🔧 Configuration Management

Go-Spring provides a flexible and powerful configuration loading mechanism, supporting configuration items from multiple
sources, easily meeting the needs of multi-environment and multi-deployment scenarios.
Whether for local development, containerized deployment, or cloud-native architectures, Go-Spring can provide consistent
and flexible configuration support.

To address the complex requirements of diverse configuration sources and coverage relationships, Go-Spring has built a
hierarchical configuration loading system.
The system automatically merges configuration items from different sources at startup and resolves and overwrites them
according to priority rules.

### 📌 Configuration Priority

1. **Command Line Arguments**  
   Use the `-Dkey=value` format to inject, with the highest priority, suitable for quickly overriding runtime
   configurations.

2. **Environment Variables**  
   Directly read from the operating system environment variables, convenient for injecting configurations in containers
   or CI/CD pipelines.

3. **Remote Files**  
   Supports pulling configurations from a configuration center, with scheduled pull and hot update capabilities,
   suitable for centralized configuration management.

4. **Local Files**  
   Supports common formats such as `.yaml`, `.properties`, and `.toml`, suitable for most development and deployment
   scenarios.

5. **In-Memory Configuration (`sysconf`)**  
   Suitable for testing scenarios or runtime temporary configuration injection, offering high flexibility.

6. **Struct Default Values**  
   Set default values through struct tags, serving as the final fallback mechanism in the configuration system.

Example: Property Binding

```go
type AppConfig struct {
   Name    string `value:"${app.name}"`
   Version string `value:"${app.version}"`
}
```

## 🔧 Bean Management

In Go-Spring, **Beans are the core building units of an application**, similar to the component concept in other
dependency injection frameworks.
The entire system is organized around the registration, initialization, dependency injection, and lifecycle management
of Beans.
Go-Spring does not rely on runtime reflection but achieves type-safe and high-performance Bean management through
compile-time metadata generation and explicit calls.
This design is particularly suitable for building **high-performance, maintainable large-scale systems**.

The framework adopts a combination of "**explicit registration + tag declaration + conditional assembly**," giving
developers clear control over Bean registration and dependency relationships.
Since it does not rely on runtime container scanning and has no magic configurations, this approach ensures a good
development experience while further enhancing debugging and operational controllability, achieving the goal of **zero
intrusion and (runtime) zero reflection**.

### 1️⃣ Registration Methods

Go-Spring provides multiple ways to register Beans:

- **`gs.Object(obj)`** - Registers an existing object as a Bean
- **`gs.Provide(ctor, args...)`** - Uses a constructor to generate and register a Bean
- **`gs.Register(bd)`** - Registers a complete Bean definition (suitable for low-level encapsulation or advanced usage)

Example:

```go
gs.Object(&Service{})  // Register a struct instance
gs.Provide(NewService) // Register using a constructor
gs.Provide(NewRepo, gs.ValueArg("db")) // Constructor with parameters
gs.Register(gs.NewBean(NewService)) // Complete definition registration
```

### 2️⃣ Injection Methods

Go-Spring offers multiple flexible dependency injection methods.

#### 1. Struct Field Injection

Inject configuration items or Beans into struct fields through tags, suitable for most scenarios.

```go
type App struct {
   Logger    *log.Logger  `autowire:""`
   Filters   []*Filter    `autowire:"access,*?"`
   StartTime time.Time    `value:"${start-time}"`
}
```

- `value:"${...}"` indicates binding configuration values;
- `autowire:""` indicates automatic injection by type and name;
- `autowire:"access,*?"` indicates injecting multiple Beans by type and name.

#### 2. Constructor Injection

Complete automatic injection through function parameters; Go-Spring automatically infers and matches dependent Beans.

```go
func NewService(logger *log.Logger) *Service {
   return &Service{Logger: logger}
}

gs.Provide(NewService)
```

#### 3. Constructor Parameter Injection

Explicitly define injection behavior through parameter wrappers, more suitable for complex construction logic:

```go
gs.Provide(NewService,
   TagArg("${log.level}"), // Inject from configuration
   ValueArg("value"),      // Direct value injection
   BindArg(parseFunc), // Option function injection
)
```

Available parameter types:

- **`TagArg(tag)`**: Extract values from configuration
- **`ValueArg(value)`**: Use fixed values
- **`IndexArg(i, arg)`**: Inject by parameter position
- **`BindArg(fn, args...)`**: Inject through option functions

### 3️⃣ Lifecycle

Developers can explicitly declare initialization, destruction, dependencies, conditional registration, and other
behaviors for each Bean.

```go
gs.Provide(NewService).
   Name("myService").
   Init(func (s *Service) { ... }).
   Destroy(func (s *Service) { ... }).
   Condition(OnProperty("feature.enabled")).
   DependsOn(gs.BeanSelectorFor[*Repo]()).
   Export(gs.As[ServiceInterface]()).
   AsRunner()
```

Configuration item descriptions:

- **`Name(string)`**: Specifies the Bean name
- **`Init(fn)`**: Initialization function (supports method name strings)
- **`Destroy(fn)`**: Destruction function (supports method name strings)
- **`DependsOn(...)`**: Specifies dependencies on other Beans
- **`Condition(...)`**: Conditional registration control
- **`Export(...)`**: Exports the Bean as an interface, supporting multiple interface exports

## ⚙️ Conditional Injection

Inspired by Spring's `@Conditional` concept, Go-Spring implements a flexible and powerful conditional injection system.
It dynamically decides whether to register a Bean based on configuration, environment, context, and other conditions,
achieving "on-demand assembly." This is particularly crucial in multi-environment deployment, plugin architectures,
feature toggles, and gray release scenarios.

### 🎯 Common Condition Types

- **`OnProperty("key")`**: Activates when the specified configuration key exists
- **`OnBean[Type]("name")`**: Activates when a Bean of the specified type/name exists
- **`OnMissingBean[Type]("name")`**: Activates when a Bean of the specified type/name does not exist
- **`OnSingleBean[Type]("name")`**: Activates when a Bean of the specified type/name is the only instance
- **`OnFunc(func(ctx ConditionContext) bool)`**: Uses custom condition logic to determine activation

Example:

```go
gs.Provide(NewService).
   Condition(OnProperty("service.enabled"))
```

The `NewService` will only be registered if `service.enabled=true` exists in the configuration file.

### 🔁 Supports Combined Conditions

Go-Spring supports combining multiple conditions to build more complex judgment logic:

- **`Not(...)`** - Negates a condition
- **`And(...)`** - All conditions must be satisfied
- **`Or(...)`** - Any condition being satisfied is sufficient
- **`None(...)`** - All conditions must not be satisfied

Example:

```go
gs.Provide(NewService).
   Condition(
      And(
         OnProperty("feature.enabled"),
         Not(OnBean[*DeprecatedService]()),
      ),
   )
```

This Bean will be enabled when `feature.enabled` is turned on and `*DeprecatedService` is not registered.

## 🔁 Dynamic Configuration

Go-Spring supports a lightweight hot configuration update mechanism. Through the generic type `gs.Dync[T]` and
`gs.RefreshProperties()`,
applications can perceive configuration changes in real-time during runtime without restarting. This is particularly
crucial for gray releases, dynamic parameter tuning, and configuration center integration in microservices
architectures.

### 🌡 Usage

1. Use `gs.Dync[T]` to declare dynamic fields

Wrap fields with the generic type `gs.Dync[T]` to listen for configuration changes and automatically update:

```go
type Config struct {
   Version gs.Dync[string] `value:"${app.version}"`
}
```

> Use `.Value()` to get the current value; the framework automatically updates this value when the configuration
> changes.

2. Call `gs.RefreshProperties()` to trigger a refresh

After the configuration changes, call this method to immediately update all dynamic fields:

```go
gs.RefreshProperties()
```

### Example: Version Update

```go
const versionKey = "app.version"

type App struct {
   Version gs.Dync[string] `value:"${app.version:=v0.0.1}"`
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   fmt.Fprintln(w, "Version:", a.Version.Value())
}

func RefreshVersion(w http.ResponseWriter, r *http.Request) {
   sysconf.Set(versionKey, "v2.0.1")
   gs.RefreshProperties()
   fmt.Fprintln(w, "Version updated!")
}
```

```go
func main() {
   gs.Object(&App{})
   gs.Provide(func (app *App) *http.ServeMux {
      http.Handle("/", app)
      http.HandleFunc("/refresh", RefreshVersion)
      return http.DefaultServeMux
   })
   gs.Run()
}
```

After running the program, accessing `/` will output the current version. After accessing `/refresh`, accessing `/`
again will show the updated version number.

## 🖥️ Custom Server

Go-Spring provides a generic `Server` interface for registering various service components (such as HTTP, gRPC,
WebSocket, etc.). All registered Servers
are automatically integrated into the application's lifecycle management, supporting concurrent startup, unified
shutdown, and other capabilities, helping developers build cleanly structured and consistently managed systems.

### 📌 Server Interface Definition

```go
type Server interface {
   ListenAndServe(sig ReadySignal) error
   Shutdown(ctx context.Context) error
}
```

- `ListenAndServe(sig ReadySignal)`: Starts the service and provides services externally after receiving the `sig`
  signal.
- `Shutdown(ctx)`: Gracefully shuts down the service and releases resources.

### 📶 ReadySignal Interface

```go
type ReadySignal interface {
   TriggerAndWait() <-chan struct{}
}
```

You can wait in `ListenAndServe` for the main process to trigger the startup completion signal before officially
providing services externally.

### Example: HTTP Server Integration

```go
func init() {
   gs.Object(NewServer()).AsServer()
}

type MyServer struct {
   svr *http.Server
}

func NewServer() *MyServer {
   return &MyServer{
      svr: &http.Server{Addr: ":8080"},
   }
}

func (s *MyServer) ListenAndServe(sig gs.ReadySignal) error {
   ln, err := net.Listen("tcp", s.svr.Addr)
   if err != nil {
      return err
   }
   <-sig.TriggerAndWait() // Wait for the startup signal
   return s.svr.Serve(ln)
}

func (s *MyServer) Shutdown(ctx context.Context) error {
   return s.svr.Shutdown(ctx)
}
```

### Example: gRPC Server Integration

```go
type GRPCServer struct {
   svr *grpc.Server
}

// ...

func (s *GRPCServer) ListenAndServe(sig gs.ReadySignal) error {
   lis, err := net.Listen("tcp", ":9595")
   if err != nil {
      return err
   }
   <-sig.TriggerAndWait()
   return s.svr.Serve(lis)
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
   s.svr.GracefulStop()
   return nil
}
```

### 💡 Multiple Servers Running Concurrently

All services registered through `.AsServer()` will start concurrently when `gs.Run()` is called and listen for exit
signals uniformly:

```go
gs.Object(&HTTPServer{}).AsServer()
gs.Object(&GRPCServer{}).AsServer()
```

## ⏳ Application Lifecycle Management

Go-Spring abstracts the application runtime cycle into three roles: `Runner`, `Job`, and `Server`, with the following
meanings:

1. **Runner**: One-time tasks executed immediately after application startup (e.g., initialization)
2. **Job**: Background tasks that run continuously during application runtime (e.g., daemon threads, polling)
3. **Server**: Long-term service processes that provide external services (e.g., HTTP/gRPC)

These roles can be registered through `.AsRunner() / .AsJob() / .AsServer()`.

Example: Runner

```go
type Bootstrap struct{}

func (b *Bootstrap) Run() error {
   fmt.Println("Bootstrap init...")
   return nil
}

func init() {
   gs.Object(&Bootstrap{}).AsRunner()
}
```

- If a Runner returns an error during execution, the application startup process will be terminated.

Example: Job

```go
type Job struct{}

func (j *Job) Run(ctx context.Context) error {
   for {
      select {
      case <-ctx.Done():
         fmt.Println("job exit")
         return nil
      default:
         if gs.Exiting() {
            return nil
         }
         time.Sleep(300 * time.Millisecond)
         fmt.Println("job tick")
      }
   }
}

func init() {
   gs.Object(&Job{}).AsJob()
}
```

- Jobs start after `gs.Run()` and continue until the exit signal arrives;
- Supports graceful shutdown, promptly responding to `ctx.Done()` or `gs.Exiting()` states.

## ⏳ Mock and Unit Testing

Go-Spring provides a unit testing framework that seamlessly integrates with the standard `go test`, making dependency
injection and mock testing simple and efficient.

### 1. Mock Object Injection

Use `gstest.MockFor[T]().With(obj)` to easily replace any bean at runtime:

```go
gstest.MockFor[*book_dao.BookDao]().With(&book_dao.BookDao{
   Store: map[string]book_dao.Book{
      "978-0132350884": {
         Title:     "Clean Code",
         Author:    "Robert C. Martin",
         ISBN:      "978-0132350884",
         Publisher: "Prentice Hall",
      },
   },
})
```

### 2. Obtain Test Objects

There are two ways to obtain the object under test:

**Directly Get Instance**:

```go
o := gstest.Get[*BookDao](t)
assert.NotNil(t, o)
```

**Structured Injection**:

```go
s := gstest.Wire(t, new(struct {
   SvrAddr string            `value:"${server.addr}"`
   Service *BookService      `autowire:""`
   BookDao *book_dao.BookDao `autowire:""`
}))
assert.That(t, s.SvrAddr).Equal("0.0.0.0:9090")
```

## 📚 Comparison with Other Frameworks

Go-Spring differentiates itself with these key features:

| Feature                  | Go-Spring | Wire | fx | dig |
|--------------------------|-----------|------|----|-----|
| Runtime IoC Container    | ✓         | ✗    | ✓  | ✓   |
| Compile-time Validation  | Partial   | ✓    | ✗  | ✗   |
| Conditional Beans        | ✓         | ✗    | ✗  | ✗   |
| Dynamic Configuration    | ✓         | ✗    | ✗  | ✗   |
| Lifecycle Management     | ✓         | ✗    | ✓  | ✗   |
| Property Binding         | ✓         | ✗    | ✗  | ✗   |
| Zero-struct Modification | ✓         | ✓    | ✗  | ✓   |

## 🏢 Who's using Go-Spring?

- ...

> Using Go-Spring and want to be featured here? Welcome to submit a PR!

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) to get started.

### 💬 QQ Group

<img src="https://raw.githubusercontent.com/go-spring/go-spring-website/master/qq(1).jpeg" width="140" height="*"  alt="qq-group"/>

### 📱 WeChat Official Account

<img src="https://raw.githubusercontent.com/go-spring/go-spring-website/master/go-spring-action.jpg" width="140" height="*"  alt="wechat-public"/>

### 🎉 Thanks!

Thanks to **JetBrains** for providing the **IntelliJ IDEA** product, which offers a convenient and fast code editing and
testing environment.

### 🛡️ License

The **Go-Spring** is released under version **2.0 of the Apache License**.
