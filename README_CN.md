<div>
    <img src="https://raw.githubusercontent.com/go-spring/go-spring/master/logo@h.png" width="140" height="*" alt="logo"/>
    <br/>
    <img src="https://img.shields.io/github/license/go-spring/spring-core" alt="license"/>
    <img src="https://img.shields.io/github/go-mod/go-version/go-spring/spring-core" alt="go-version"/>
    <img src="https://img.shields.io/github/v/release/go-spring/spring-core?include_prereleases" alt="release"/>
    <img src="https://codecov.io/gh/go-spring/spring-core/branch/main/graph/badge.svg" alt="test-coverage"/>
    <br/>
</div>

[English](README.md)

Go-Spring 是一个面向现代 Go 应用开发的高性能框架，灵感源于 Java 社区的 Spring / Spring Boot，但设计理念完全贴合 Go 语言本身。
它致力于将 Spring 世界成熟的开发范式（如依赖注入、自动配置、生命周期管理等）引入 Go，同时保持原生库的极简风格与执行效率。
你可以像使用 Spring Boot 那样轻松构建 Go 应用，几乎无需模板代码，也不受侵入式约束。
无论是构建单体系统，还是分布式服务网格，Go-Spring 都提供了“一站式”开发体验，帮助你显著提升开发效率与可维护性。

## 🌟 框架亮点

1. ⚡ **极致启动性能**
   - 基于 Go 的 `init()` 机制实现 Bean 注册，跳过运行时扫描，启动迅速；
   - 注入只依赖初始化阶段的反射，运行时零反射，保障极致性能。

2. 🧩 **开箱即用、无侵入式设计**
   - 支持结构体标签注入与链式配置，无需掌握复杂概念即可使用；
   - 不强依赖接口或父类，业务逻辑保持原生 Go 风格。

3. 🔄 **配置热更新，实时生效**
   - 多格式、多源配置加载，支持环境隔离与动态刷新；
   - 配置变更即时生效，适用于灰度发布与动态调参。

4. ⚙️ **灵活依赖注入机制**
   - 支持构造函数、结构体字段、参数注入等方式；
   - 注入行为可按配置、环境等条件灵活控制。

5. 🔌 **多模型服务启动支持**
   - 内建 HTTP Server 启动器；
   - 支持 `Runner`、`Job`、`Server` 三类运行模型，便于构建多形态微服务架构；
   - 生命周期钩子完善，支持优雅退出。

6. 🧪 **内建测试能力**
   - 原生集成 Mock、自动注入，轻松实现高可测性的单元测试。

## 📦 安装方式

Go-Spring 使用 Go Modules 管理依赖，安装非常简单：

```bash
go get github.com/go-spring/spring-core
```

## 🚀 快速开始

Go-Spring 主打“开箱即用”，下面通过两个示例，快速感受其强大能力。

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

服务结构体：

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

运行：

```bash
curl http://127.0.0.1:9090/echo     # 查看当前时间
curl http://127.0.0.1:9090/refresh  # 触发热刷新
```

✅ `value` 标签自动绑定配置；  
✅ `gs.Dync[T]` 实现字段热更新；  
✅ `gs.Provide()` 注入依赖，保持标准库 API 完整性。

### 更多示例

📎 更多示例请见：[gs/examples](gs/examples)

## 🔧 配置管理

Go-Spring 提供了灵活强大的配置加载机制，支持从多种来源获取配置项，轻松满足多环境、多部署场景的需求。

### 📌 配置优先级（从高到低）

1. **命令行参数**  
   使用 `-Dkey=value` 格式注入，优先级最高。
2. **环境变量**  
   直接读取系统环境变量。
3. **远程文件**  
   支持定时拉取与热更新。
4. **本地文件**  
   支持格式：`.yaml` `.yml` `.properties` `.toml`
5. **内存配置 (`sysconf`)**  
   适用于测试或运行时动态注入。
6. **结构体默认值**  
   通过标签设置，优先级最低。

#### 📁 示例：YAML 配置文件

**文件：`config/app.yml`**

```yaml
server:
   port: 8080
app:
   name: demo-app
   version: 1.0.0
```

#### 🔗 示例：结构体绑定配置

```go
type AppConfig struct {
   Name    string `value:"${app.name}"`
   Version string `value:"${app.version}"`
}
```

## 🔧 Bean 管理

在 Go-Spring 中，**Bean 是应用的核心构建单元**。框架采用显式注册 + 标签声明的模式，结合灵活的条件装配，
做到了 **零侵入、零反射（运行时）**，非常适合构建大型可维护系统。

### 注册方式

Go-Spring 提供多种方式注册 Bean：

- **`gs.Object(obj)`** - 将已有对象注册为 Bean
- **`gs.Provide(ctor, args...)`** - 使用构造函数生成并注册 Bean
- **`gs.Register(bd)`** - 注册完整 Bean 定义（适合底层封装或高级用法）
- **`gs.GroupRegister(fn)`** - 批量注册多个 Bean（常用于模块初始化等场景）

#### 示例

```go
gs.Object(&Service{})  // 注册结构体实例
gs.Provide(NewService) // 使用构造函数注册
gs.Provide(NewRepo, ValueArg("db")) // 构造函数带参数
gs.Register(gs.NewBean(NewService)) // 完整定义注册

// 批量注册多个 Bean
gs.GroupRegister(func (p Properties) []*BeanDefinition {
    return []*BeanDefinition{
        gs.NewBean(NewUserService),
        gs.NewBean(NewOrderService),
    }
})
```

### 注入方式

Go-Spring 提供多种灵活的依赖注入方式，支持结构体字段注入、构造函数注入、参数化注入等，兼容配置绑定与 Bean 引用，几乎适配所有开发需求。

#### 1️⃣ 结构体字段注入

通过标签将配置项或 Bean 注入结构体字段，适合绝大多数场景。

```go
type App struct {
   Logger    *log.Logger  `autowire:""`
   Filters   []*Filter    `autowire:"access,*?"`
   StartTime time.Time    `value:"${start-time}"`
}
```

- `autowire:""`  表示按类型自动注入；  
- `value:"${...}"` 表示绑定配置值。

### 2️⃣ 构造函数注入

通过函数参数完成自动注入，Go-Spring 会自动推断并匹配依赖 Bean。

```go
func NewService(logger *log.Logger) *Service {
   return &Service{Logger: logger}
}

gs.Provide(NewService)
```

### 3️⃣ 构造参数注入（自定义注入方式）

可通过参数包装器明确注入行为，更适用于复杂构造逻辑：

```go
gs.Provide(NewService,
    TagArg("${log.level}"),        // 从配置注入
    ValueArg("some static value"), // 直接值注入
    BindArg(parseFunc, "arg"),     // option 函数注入
)
```

可用的参数类型：

- **`TagArg(tag)`**：从配置中提取值
- **`ValueArg(value)`**：使用固定值
- **`IndexArg(i, arg)`**：按参数位置注入
- **`BindArg(fn, args...)`**：通过 option 函数注入

### 生命周期

Go-Spring 提供完整的 Bean 生命周期管理机制，开发者可以为每个 Bean
显式声明初始化、销毁、依赖、条件注册等行为，并将其声明为应用组件（Runner、Job、Server）参与整个应用流程。

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
- **`AsRunner()`**：注册为 `Runner`
- **`AsJob()`**：注册为后台任务 Job
- **`AsServer()`**：注册为服务 Server

## ⚙️ 条件注入

Go-Spring 借鉴 Spring 的 `@Conditional` 思想，实现了灵活强大的条件注入系统。通过配置、环境、上下文等条件动态决定 Bean
是否注册，实现“按需装配”。 这在多环境部署、插件化架构、功能开关、灰度发布等场景中尤为关键。

#### 🎯 常用条件类型

- **`OnProperty("key")`**：当指定配置 key 存在时激活
- **`OnMissingProperty("key")`**：当指定配置 key 不存在时激活
- **`OnBean[Type]("name")`**：当指定类型/名称的 Bean 存在时激活
- **`OnMissingBean[Type]("name")`**：当指定类型/名称的 Bean 不存在时激活
- **`OnSingleBean[Type]("name")`**：当指定类型/名称的 Bean 是唯一实例时激活
- **`OnFunc(func(ctx CondContext) bool)`**：使用自定义条件逻辑判断是否激活

#### 🧪 示例：按配置激活 Bean

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

#### 示例：组合条件注册控制

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

#### 💡 示例：实时版本号更新

```go
const versionKey = "app.version"

type App struct {
    Version gs.Dync[string] `value:"${app.version}"`
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

注册路由并启动应用：

```go
gs.Object(&App{})
gs.Provide(func (app *App) *http.ServeMux {
    http.Handle("/", app)
    http.HandleFunc("/refresh", RefreshVersion)
    return http.DefaultServeMux
})
gs.Run()
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

### 🛠 示例：HTTP Server 接入

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

这样 Server 就会随应用启动自动运行，并在退出时自动关闭。

### 🌐 示例：接入 gRPC Server

```go
type GRPCServer struct {
    svr *grpc.Server
    lis net.Listener
}

func (s *GRPCServer) ListenAndServe(sig gs.ReadySignal) error {
    var err error
    s.lis, err = net.Listen("tcp", ":50051")
    if err != nil {
        return err
    }
    <-sig.TriggerAndWait()
    return s.svr.Serve(s.lis)
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
    stopped := make(chan struct{})
    go func () {
        s.svr.GracefulStop()
        close(stopped)
    }()
    select {
    case <-ctx.Done():
        s.server.Stop()
    case <-stopped:
    }
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

Go-Spring 在设计上对应用启动、运行、退出过程进行了封装和抽象，提供了以下三个核心生命周期角色：

1. **Runner**：应用启动后立即执行的一次性任务（初始化等）
2. **Job**：应用运行期间持续运行的后台任务（守护线程、轮询等）
3. **Server**：对外提供服务的长期服务进程（如 HTTP/gRPC 等）

这些角色可通过 `.AsRunner() / .AsJob() / .AsServer()` 进行注册。

### 🚀 Runner（应用启动后执行一次）

适用于数据预热、系统初始化、打印信息等场景：

```go
type Bootstrap struct{}

func (b *Bootstrap) Run(ctx context.Context) error {
    fmt.Println("Bootstrap init...")
    return nil
}

func init() {
    gs.Object(&Bootstrap{}).AsRunner()
}
```

Runner 执行过程中如果返回错误，将会终止应用启动流程。

### 🔄 Job（后台任务）

适合执行周期任务、健康检查、定时拉取等持续性逻辑：

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






