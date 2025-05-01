# Go-Spring

<div>
   <img src="https://img.shields.io/github/license/go-spring/spring-core" alt="license"/>
   <img src="https://img.shields.io/github/go-mod/go-version/go-spring/spring-core" alt="go-version"/>
   <img src="https://img.shields.io/github/v/release/go-spring/spring-core?include_prereleases" alt="release"/>
   <img src="https://codecov.io/gh/go-spring/spring-core/branch/main/graph/badge.svg" alt="test-coverage"/>
</div>

[English](README.md)

**Go-Spring 是一个面向现代 Go 应用开发的高性能框架，灵感源自 Java 社区的 Spring / Spring Boot。**
它的设计理念深度融合 Go 语言的特性，既保留了 Spring 世界中成熟的开发范式，如依赖注入（DI）、自动配置和生命周期管理，
又避免了传统框架可能带来的繁复和性能开销。
Go-Spring 让开发者能够在保持 Go 原生风格与执行效率的前提下，享受更高层次的抽象与自动化能力。

**无论你是在开发单体应用，还是构建基于微服务的分布式系统，Go-Spring 都提供了统一且灵活的开发体验。**
它以“开箱即用”的方式简化了项目搭建流程，减少模板代码的编写需求，并且不强加侵入式的框架结构，让开发者可以更专注于业务逻辑的实现。
Go-Spring 致力于提升开发效率、可维护性和系统的一致性，是 Go 语言生态中一个具有里程碑意义的框架。

## 🚀 特性一览

Go-Spring 提供了丰富而实用的特性，帮助开发者高效构建现代 Go 应用：

1. ⚡ **极致启动性能**
   - 基于 Go 的 `init()` 机制进行 Bean 注册，无运行时扫描，启动迅速；
   - 注入仅依赖初始化阶段的反射，运行时零反射，保障性能最大化。

2. 🧩 **开箱即用、无侵入式设计**
   - 支持结构体标签注入与链式配置，无需掌握复杂概念即可使用；
   - 不强依赖接口或继承结构，业务逻辑保持原生 Go 风格。

3. 🔄 **配置热更新，实时生效**
   - 多格式、多来源配置加载，支持环境隔离与动态刷新；
   - 配置变更可即时应用，便于灰度发布与在线调参。

4. ⚙️ **灵活依赖注入机制**
   - 支持构造函数注入、结构体字段注入、构造函数参数注入多种方式；
   - 注入行为可按配置项或运行环境灵活调整。

5. 🔌 **多模型服务启动支持**
   - 内建 HTTP Server 启动器，快速部署 Web 服务；
   - 支持 `Runner`、`Job`、`Server` 三类运行模型，适配不同服务形态；
   - 生命周期钩子完备，支持优雅启停。

6. 🧪 **内建测试能力**
   - 与 `go test` 无缝集成，支持 Bean Mock 和依赖注入，轻松编写单元测试。

## 📦 安装方式

Go-Spring 使用 Go Modules 管理依赖，安装非常简单：

```bash
go get github.com/go-spring/spring-core
```

## 🚀 快速开始

Go-Spring 主打“开箱即用”，下面通过两个示例，快速感受其强大能力。

> 更多示例请见：[gs/examples](gs/examples)

### 示例一：最小 API 服务

```go
func main() {
    http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("hello world!"))
    })
    gs.Run()
}
```

访问方式：

```bash
curl http://127.0.0.1:9090/echo
# 输出: hello world!
```

✅ 无需繁杂配置，Go 标准库 `http` 可以直接使用;  
✅ `gs.Run()` 接管生命周期，支持优雅退出、信号监听等能力。

### 示例二：基础特性展示

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

访问方式：

```bash
curl http://127.0.0.1:9090/echo     # 查看当前时间
curl http://127.0.0.1:9090/refresh  # 触发热刷新
```

✅ `value` 标签自动绑定配置；  
✅ `gs.Dync[T]` 实现字段热更新；  
✅ `gs.Object` `gs.Provide()` 注册 Bean。

## 🔧 配置管理

Go-Spring 提供了灵活强大的配置加载机制，支持从多种来源获取配置项，轻松满足多环境、多部署场景的需求。
无论是本地开发、容器化部署，还是云原生架构，Go-Spring 都能够提供一致而灵活的配置支持。

为了应对配置项来源多样、覆盖关系复杂的实际需求，Go-Spring 构建了一套分层配置加载体系。
系统会在启动时自动合并不同来源的配置项，并按照优先级规则进行解析和覆盖。

### 📌 配置优先级

1. **命令行参数**  
   使用 `-Dkey=value` 格式注入，优先级最高，适合快速覆盖运行时配置。

2. **环境变量**  
   直接读取操作系统环境变量，方便在容器或 CI/CD 流水线中注入配置。

3. **远程文件**  
   支持从配置中心拉取配置，具备定时拉取与热更新能力，适用于集中式配置管理。

4. **本地文件**  
   支持常见格式，如 `.yaml`、`.properties`、`.toml`，适合大多数开发与部署场景。

5. **内存配置 (`sysconf`)**  
   适用于测试场景或运行时临时注入配置，具备较高的灵活性。

6. **结构体默认值**  
   通过结构体标签设定默认值，是配置体系中的最后兜底机制。

示例：属性绑定

```go
type AppConfig struct {
   Name    string `value:"${app.name}"`
   Version string `value:"${app.version}"`
}
```

## 🔧 Bean 管理

在 Go-Spring 中，**Bean 是应用的核心构建单元**，类似于其他依赖注入框架中的组件概念。
整个系统围绕 Bean 的注册、初始化、依赖注入与生命周期管理进行组织。
Go-Spring 不依赖运行时反射，而是通过编译期生成元数据和显式调用方式，实现了类型安全、性能优越的 Bean 管理机制。
这样设计特别适合构建 **高性能、可维护性强的大型系统**。

框架采用“**显式注册 + 标签声明 + 条件装配**”的组合方式，让开发者对 Bean 的注册与依赖关系有清晰的控制。
由于不依赖运行时容器扫描，也没有魔法配置，这种做法在保证开发体验的同时，
进一步提升了调试和运维的可控性，实现了**零侵入、（运行时）零反射**的目标。

### 1️⃣ 注册方式

Go-Spring 提供多种方式注册 Bean：

- **`gs.Object(obj)`** - 将已有对象注册为 Bean
- **`gs.Provide(ctor, args...)`** - 使用构造函数生成并注册 Bean
- **`gs.Register(bd)`** - 注册完整 Bean 定义（适合底层封装或高级用法）
- **`gs.GroupRegister(fn)`** - 批量注册多个 Bean（常用于模块初始化等场景）

示例:

```go
gs.Object(&Service{})  // 注册结构体实例
gs.Provide(NewService) // 使用构造函数注册
gs.Provide(NewRepo, gs.ValueArg("db")) // 构造函数带参数
gs.Register(gs.NewBean(NewService))    // 完整定义注册

// 批量注册多个 Bean
gs.GroupRegister(func (p conf.Properties) []*gs.BeanDefinition {
    return []*gs.BeanDefinition{
        gs.NewBean(NewUserService),
        gs.NewBean(NewOrderService),
    }
})
```

### 2️⃣ 注入方式

Go-Spring 提供了多种灵活的依赖注入方式。

#### 1. 结构体字段注入

通过标签将配置项或 Bean 注入结构体字段，适合绝大多数场景。

```go
type App struct {
   Logger    *log.Logger  `autowire:""`
   Filters   []*Filter    `autowire:"access,*?"`
   StartTime time.Time    `value:"${start-time}"`
}
```

- `value:"${...}"` 表示绑定配置值；
- `autowire:""`  表示按类型和名称自动注入；  
- `autowire:"access,*?"` 表示按类型和名称注入多个 Bean。

#### 2. 构造函数注入

通过函数参数完成自动注入，Go-Spring 会自动推断并匹配依赖 Bean。

```go
func NewService(logger *log.Logger) *Service {
   return &Service{Logger: logger}
}

gs.Provide(NewService)
```

#### 3. 构造函数参数注入

可通过参数包装器明确注入行为，更适用于复杂构造逻辑：

```go
gs.Provide(NewService,
    TagArg("${log.level}"), // 从配置注入
    ValueArg("value"),      // 直接值注入
    BindArg(parseFunc),     // option 函数注入
)
```

可用的参数类型：

- **`TagArg(tag)`**：从配置中提取值
- **`ValueArg(value)`**：使用固定值
- **`IndexArg(i, arg)`**：按参数位置注入
- **`BindArg(fn, args...)`**：通过 option 函数注入

### 3️⃣ 生命周期

开发者可以为每个 Bean 显式声明初始化、销毁、依赖、条件注册等行为。

```go
gs.Provide(NewService).
    Name("myService").
    Init(func(s *Service) { ... }).
    Destroy(func(s *Service) { ... }).
    Condition(OnProperty("feature.enabled")).
    DependsOn(gs.BeanSelectorFor[*Repo]()).
    Export(gs.As[ServiceInterface]()).
    AsRunner()
```

配置项说明：

- **`Name(string)`**：指定 Bean 名称
- **`Init(fn)`**：初始化函数（支持方法名字符串）
- **`Destroy(fn)`**：销毁函数（支持方法名字符串）
- **`DependsOn(...)`**：指定依赖的其他 Bean
- **`Condition(...)`**：条件注册控制
- **`Export(...)`**：将 Bean 作为接口导出，支持多接口导出

## ⚙️ 条件注入

Go-Spring 借鉴 Spring 的 `@Conditional` 思想，实现了灵活强大的条件注入系统。通过配置、环境、上下文等条件动态决定 Bean
是否注册，实现“按需装配”。 这在多环境部署、插件化架构、功能开关、灰度发布等场景中尤为关键。

### 🎯 常用条件类型

- **`OnProperty("key")`**：当指定配置 key 存在时激活
- **`OnMissingProperty("key")`**：当指定配置 key 不存在时激活
- **`OnBean[Type]("name")`**：当指定类型/名称的 Bean 存在时激活
- **`OnMissingBean[Type]("name")`**：当指定类型/名称的 Bean 不存在时激活
- **`OnSingleBean[Type]("name")`**：当指定类型/名称的 Bean 是唯一实例时激活
- **`OnFunc(func(ctx CondContext) bool)`**：使用自定义条件逻辑判断是否激活

示例：

```go
gs.Provide(NewService).
    Condition(OnProperty("service.enabled"))
```

只有当配置文件中存在 `service.enabled=true` 时，`NewService` 才会注册。

### 🔁 支持组合条件

Go-Spring 支持组合多个条件，构建更复杂的判断逻辑：

- **`Not(...)`** - 对条件取反
- **`And(...)`** - 所有条件都满足时成立
- **`Or(...)`** - 任一条件满足即成立
- **`None(...)`** - 所有条件都不满足时成立

示例：

```go
gs.Provide(NewService).
    Condition(
        And(
            OnProperty("feature.enabled"),
            Not(OnBean[*DeprecatedService]()),
        ),
    )
```

该 Bean 会在 `feature.enabled` 开启且未注册 `*DeprecatedService` 时启用。

## 🔁 动态配置

Go-Spring 支持轻量级的配置热更新机制。通过泛型类型 `gs.Dync[T]` 和 `gs.RefreshProperties()`，
应用可以在运行中实时感知配置变更，而无需重启。这对于微服务架构中的灰度发布、动态调参、配置中心集成等场景尤为关键。

### 🌡 使用方式

1. 使用 `gs.Dync[T]` 声明动态字段

通过泛型类型 `gs.Dync[T]` 包装字段，即可监听配置变化并自动更新：

```go
type Config struct {
    Version gs.Dync[string] `value:"${app.version}"`
}
```

> 调用时通过 `.Value()` 获取当前值，框架在配置变更时会自动更新该值。

2. 调用 `gs.RefreshProperties()` 触发刷新

在配置发生变化后，调用此方法可以让所有动态字段立即更新：

```go
gs.RefreshProperties()
```

### 示例：版本号更新

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
   gs.Provide(func(app *App) *http.ServeMux {
      http.Handle("/", app)
      http.HandleFunc("/refresh", RefreshVersion)
      return http.DefaultServeMux
   })
   gs.Run()
}
```

运行程序后，访问 `/` 会输出当前版本，访问 `/refresh` 后，再次访问 `/` 即可看到更新后的版本号。

## 🖥️ 自定义 Server

Go-Spring 提供了通用的 `Server` 接口，用于注册各种服务组件（如 HTTP、gRPC、WebSocket 等）。所有注册的 Server
都会自动接入应用的生命周期管理，支持并发启动、统一关闭等能力，帮助开发者构建结构整洁、管理一致的系统。

### 📌 Server 接口定义

```go
type Server interface {
    ListenAndServe(sig ReadySignal) error
    Shutdown(ctx context.Context) error
}
```

- `ListenAndServe(sig ReadySignal)`: 启动服务，并在收到 `sig` 信号后对外提供服务。
- `Shutdown(ctx)`: 优雅关闭服务，释放资源。

### 📶 ReadySignal 接口

```go
type ReadySignal interface {
    TriggerAndWait() <-chan struct{}
}
```

你可以在 `ListenAndServe` 中等到主流程触发启动完成信号，然后正式对外提供服务。

### 示例：HTTP Server 接入

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
    <-sig.TriggerAndWait() // 等待启动信号
    return s.svr.Serve(ln)
}

func (s *MyServer) Shutdown(ctx context.Context) error {
    return s.svr.Shutdown(ctx)
}
```

### 示例：gRPC Server 接入

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

### 💡 多 Server 并发运行

所有通过 `.AsServer()` 注册的服务，会在 `gs.Run()` 时并发启动，并统一监听退出信号：

```go
gs.Object(&HTTPServer{}).AsServer()
gs.Object(&GRPCServer{}).AsServer()
```

## ⏳ 应用生命周期管理

Go-Spring 将应用运行周期抽象为三个角色：`Runner`、`Job`、`Server`，含义分别如下：

1. **Runner**：应用启动后立即执行的一次性任务（初始化等）
2. **Job**：应用运行期间持续运行的后台任务（守护线程、轮询等）
3. **Server**：对外提供服务的长期服务进程（如 HTTP/gRPC 等）

这些角色可以通过 `.AsRunner() / .AsJob() / .AsServer()` 进行注册。

示例：Runner

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

- Runner 执行过程中如果返回错误，将会终止应用启动流程。

示例：Job

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

- Job 会在 `gs.Run()` 后启动，直到退出信号到来；
- 支持优雅停止，及时响应 `ctx.Done()` 或 `gs.Exiting()` 状态。

## ⏳ Mock 与单元测试

Go-Spring 提供了与标准 `go test` 无缝集成的单元测试框架，让依赖注入和模拟测试变得简单高效。

### 1. 模拟对象注入

使用 `gstest.MockFor[T]().With(obj)` 可以在运行时轻松替换任何 bean：

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

### 2. 获取测试对象

有两种方式获取被测试对象：

**直接获取实例**：

```go
o := gstest.Get[*BookDao](t)
assert.NotNil(t, o)
```

**结构化注入**：

```go
s := gstest.Wire(t, new(struct {
   SvrAddr string            `value:"${server.addr}"`
   Service *BookService      `autowire:""`
   BookDao *book_dao.BookDao `autowire:""`
}))
assert.That(t, s.SvrAddr).Equal("0.0.0.0:9090")
```

## 📚 与其他框架的对比

Go-Spring 具备以下几个显著优势：

| 功能点              | Go-Spring | Wire | fx | dig |
|------------------|-----------|------|----|-----|
| 运行时 IoC 容器       | ✓         | ✗    | ✓  | ✓   |
| 编译期校验            | 部分支持      | ✓    | ✗  | ✗   |
| 条件 Bean 支持       | ✓         | ✗    | ✗  | ✗   |
| 动态配置能力           | ✓         | ✗    | ✗  | ✗   |
| 生命周期管理           | ✓         | ✗    | ✓  | ✗   |
| 属性绑定             | ✓         | ✗    | ✗  | ✗   |
| 零结构体侵入（无需修改原结构体） | ✓         | ✓    | ✗  | ✓   |

## 🏢 谁在使用 Go-Spring？

- ...

> 在使用 Go-Spring 并希望展示在此处？欢迎提交 PR！

## 🤝 参与贡献

我们欢迎所有形式的贡献！请查阅 [CONTRIBUTING.md](CONTRIBUTING.md) 获取参与方式。
