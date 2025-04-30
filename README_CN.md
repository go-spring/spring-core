<div>
    <img src="https://raw.githubusercontent.com/go-spring/go-spring/master/logo@h.png" width="140" height="*" alt="logo"/>
    <br/>
    <img src="https://img.shields.io/github/license/go-spring/spring-core" alt="license"/>
    <img src="https://img.shields.io/github/go-mod/go-version/go-spring/spring-core" alt="go-version"/>
    <img src="https://img.shields.io/github/v/release/go-spring/spring-core?include_prereleases" alt="release"/>
    <img src="https://codecov.io/gh/go-spring/spring-core/branch/main/graph/badge.svg" alt="test-coverage"/>
</div>

**Go-Spring** 是一个功能强大、使用方便的 Go 应用开发框架，其灵感来源于 Java 生态中的 Spring 和 Spring Boot，
它的设计目标是将 Java 世界中的优秀开发理念无缝迁移到 Go 语言中，从而提升开发效率、增强模块可复用性、提高代码可维护性。

它为 Go 应用带来了类似 Spring Boot 的体验，提供自动配置、依赖注入、配置热更新、条件注入、生命周期管理、微服务支持等功能，
力求“一站式”解决实际开发问题。同时，它又高度兼容 Go 标准库，延续了 Go 一贯的简洁与高性能，特别适合构建现代 Go 微服务系统。

### 🌟 框架亮点

- ⚡ **秒级启动**  
  利用 Go 的 `init()` 机制实现 Bean 主动注册，省去运行时扫描，提升应用启动速度。

- 🧩 **极致易用**  
  支持结构体标签注入和链式 API 配置，开发者无需编写复杂的模板代码，快速上手开发。

- 🔄 **配置热更新**  
  支持动态属性绑定与运行时刷新，无需重启应用即可实时生效，适用于灰度发布、动态调整等场景。

- 📦 **微服务原生支持**  
  内置标准 HTTP Server 启动器与注册机制，具备丰富的生命周期钩子，构建微服务更高效。

- 🧪 **完善的测试能力**  
  提供 Mock 与单元测试工具，便于开发者编写高质量、可验证的测试用例。

- 🔍 **运行时零反射**  
  框架仅在启动时使用反射完成 Bean 构造与注入，运行时不依赖反射，保障性能表现。

- 💡 **零侵入式设计**  
  框架对业务代码无强依赖，使用者无需实现特定接口即可被管理，保持代码干净、易迁移。

## ✨ 核心功能总览

| 功能           | 描述                                       |
|--------------|------------------------------------------|
| 🚀 自动配置      | 自动加载配置文件、构建 Bean，支持环境隔离与多文件合并            |
| ⚙️ 依赖注入      | 结构体字段注入、构造函数注入、接口注入等多种形式                 |
| 🌀 配置热更新     | 支持运行时刷新配置，动态响应配置变更                       |
| 🔄 生命周期管理    | 支持自定义初始化与销毁函数，并提供优雅的退出机制                 |
| 🔌 服务注册      | 原生兼容 HTTP，支持自定义 Server 模型                |
| 🧪 条件注入      | 支持按属性、环境、Bean 存在与否等灵活注入控制                |
| 🔧 Bean 注册管理 | 提供灵活的 Bean 注册与构建 API                     |
| 📡 微服务支持     | 内建 Job、Runner、Server 三种运行模型，助力构建多形态微服务架构 |
| 🧪 单元测试支持    | 内置 Mock、自动注入等机制，支持高质量测试开发                |

## 📦 安装

Go-Spring 使用 Go Modules 管理依赖，安装非常简单：

```bash
go get github.com/go-spring/spring-core
```

## 🚀 快速开始

Go-Spring 的核心理念之一就是**开箱即用**。下面通过两个简单示例快速体验它的能力。

### 示例一：最小 API 使用

```go
func main() {
    http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
        _, _ = w.Write([]byte("hello world!"))
    })
    gs.Run()
}
```

在这个例子中你可以看到：

- 无需繁杂配置，Go 标准库 `http` 可以直接使用
- `gs.Run()` 会托管应用生命周期，包括信号监听、优雅退出等

运行后即可通过如下命令访问服务：

```bash
curl http://127.0.0.1:9090/echo
# 输出: hello world!
```

### 示例二：Startup 基础用法

该示例展示了 Go-Spring 的核心能力：**属性绑定**、**依赖注入**、**配置动态刷新**、**标准库兼容**等。

```go
func init() {
    // Register the Service struct as a bean.
    gs.Object(&Service{})

    // Provide a [*http.ServeMux] as a bean.
    gs.Provide(func(s *Service) *http.ServeMux {
        http.HandleFunc("/echo", s.Echo)
        http.HandleFunc("/refresh", s.Refresh)
        return http.DefaultServeMux
    })

    sysconf.Set("start-time", time.Now().Format(timeLayout))
    sysconf.Set("refresh-time", time.Now().Format(timeLayout))
}
```

服务结构体定义如下：

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
    _, _ = w.Write([]byte(str))
}

func (s *Service) Refresh(w http.ResponseWriter, r *http.Request) {
    sysconf.Set("refresh-time", time.Now().Format(timeLayout))
    _ = gs.RefreshProperties()
    _, _ = w.Write([]byte("OK!"))
}
```

主函数入口：

```go
func main() {
    gs.Run()
}
```

请求示例：

```bash
curl http://127.0.0.1:9090/echo
# 输出: start-time: ... refresh-time: ...

curl http://127.0.0.1:9090/refresh
# 输出: OK!

curl http://127.0.0.1:9090/echo
# 输出中的 refresh-time 已更新
```

## 🔧 配置管理

Go-Spring 提供了灵活强大的配置加载机制，支持从多种来源获取配置项，轻松满足多环境、多部署场景的需求。

### 🔍 支持的配置来源

| 来源类型      | 描述                                          |
|-----------|---------------------------------------------|
| `sysconf` | 内存配置，适用于测试或临时注入                             |
| 本地文件      | 支持 `.yaml`、`.yml`、`.properties`、`.toml` 等格式 |
| 远程文件      | 通过远程 URL 拉取配置，支持定时轮询更新                      |
| 环境变量      | 读取系统环境变量作为配置项                               |
| 命令行参数     | 以 `--key=value` 形式注入参数，覆盖配置文件与环境变量          |

### 🔗 配置加载优先级（从高到低）

1. 命令行参数
2. 环境变量
3. 远程配置文件
4. 本地配置文件
5. `sysconf` 内存设置
6. 默认值（通过结构体标签设置）

#### 示例配置文件：

```yaml
# config/app.yml
server:
  port: 8080
app:
  name: demo-app
  version: 1.0.0
```

结构体绑定：

```go
type AppConfig struct {
    Name    string `value:"${app.name}"`
    Version string `value:"${app.version}"`
}
```

### 🌡️ 热更新配置

Go-Spring 支持热更新，结合 `gs.Dync[T]` 类型，可以实时响应配置变化，而无需重启服务。

```go
type AppInfo struct {
    Version gs.Dync[string] `value:"${app.version}"`
}
```

运行时触发刷新：

```go
_ = gs.RefreshProperties()
```

刷新后，所有 `gs.Dync[T]` 绑定字段会自动更新。

## 🔧 Bean 管理

在 Go-Spring 中，**Bean 是应用的核心构建单元**。框架采用显式注册 + 标签声明的模式，结合灵活的条件装配，
做到了 **零侵入、零反射（运行时）**，非常适合构建大型可维护系统。

### ✅ Bean 注册方式

Go-Spring 提供多种方式注册 Bean：

| 方法                          | 描述                       |
|-----------------------------|--------------------------|
| `gs.Object(obj)`            | 将已有对象注册为 Bean            |
| `gs.Provide(ctor, args...)` | 使用构造函数生成并注册 Bean         |
| `gs.Register(bd)`           | 注册完整 Bean 定义，适合底层封装或高级用法 |
| `gs.GroupRegister(fn)`      | 批量注册多个 Bean，常用于模块初始化等场景  |

#### 示例

```go
gs.Object(&Service{})  // 注册结构体实例
gs.Provide(NewService) // 使用构造函数注册
gs.Provide(NewRepo, ValueArg("db")) // 构造函数带参数
gs.Register(gs.NewBean(NewService)) // 完整定义注册
```

批量注册：

```go
gs.GroupRegister(func(p Properties) []*BeanDefinition {
    return []*BeanDefinition{
        gs.NewBean(NewUserService),
        gs.NewBean(NewOrderService),
    }
})
```

### 💉 注入方式

Go-Spring 支持字段注入、构造函数注入以及构造参数注入。

#### 1. 字段注入

通过标签绑定依赖 Bean 或配置项：

```go
type App struct {
    Logger    *log.Logger  `autowire:""`
    StartTime time.Time    `value:"${start-time}"`
}
```

- `autowire:""`：表示自动注入依赖 Bean（根据类型或名称）
- `value:"${...}"`：表示绑定配置属性值

#### 2. 构造函数注入

依赖通过构造函数参数自动注入：

```go
func NewService(logger *log.Logger) *Service {
    return &Service{Logger: logger}
}

gs.Provide(NewService)
```

#### 3. 构造参数注入

通过包装器指定注入方式：

```go
gs.Provide(NewService,
    TagArg("${log.level}"), // 从配置注入
    ValueArg("some static value"), // 直接值注入
    BindArg(parseFunc, "arg"), // option 函数注入
)
```

可用的参数类型：

| 参数类型                | 描述          |
|---------------------|-------------|
| `TagArg(tag)`       | 从配置中提取值     |
| `ValueArg(value)`   | 使用固定值       |
| `IndexArg(i, v)`    | 按参数位置注入     |
| `BindArg(fn, args)` | option 函数注入 |

### 🔄 Bean 生命周期配置

每个 Bean 支持自定义生命周期行为，包括初始化、销毁、条件注册等：

```go
gs.Provide(NewService).
    Name("myService").
    Init(func(s *Service) { ... }).
    Destroy(func(s *Service) { ... }).
    Condition(OnProperty("feature.enabled")).
    DependsOn("logger").
    Export((*MyInterface)(nil)).
    AsRunner()
```

配置项说明：

| 方法               | 说明                    |
|------------------|-----------------------|
| `Name(string)`   | 指定 Bean 名称            |
| `Init(fn)`       | 初始化函数（支持方法名字符串）       |
| `Destroy(fn)`    | 销毁函数（支持方法名字符串）        |
| `DependsOn(...)` | 指定依赖的其他 Bean 名称       |
| `Condition(...)` | 条件装配控制（见下一节）          |
| `Export(...)`    | 将 Bean 作为接口导出，支持多接口导出 |
| `AsRunner()`     | 注册为 `Runner`，运行在主线程   |
| `AsJob()`        | 注册为后台任务 Job           |
| `AsServer()`     | 注册为服务 Server（需实现接口）   |

## ⚙️ 条件注入（Condition）

Go-Spring 支持基于条件的 Bean 注入机制，这使得组件可以根据运行时环境、配置状态或其他上下文信息进行“按需装配”，
类似于 Java Spring 的 `@Conditional`。

这种机制特别适合复杂应用场景，比如：多环境部署、插件系统、功能开关、灰度发布等。

### 🎯 支持的条件类型

| 条件方法                                 | 描述                      |
|--------------------------------------|-------------------------|
| `OnProperty("key")`                  | 指定配置 key 存在并有值时激活       |
| `OnMissingProperty("key")`           | 指定配置 key 不存在时激活         |
| `OnBean[Type]("name")`               | 当指定类型/名称的 Bean 存在时激活    |
| `OnMissingBean[Type]("name")`        | 当指定类型/名称的 Bean 不存在时激活   |
| `OnSingleBean[Type]("name")`         | 当指定类型/名称的 Bean 是唯一实例时激活 |
| `OnFunc(func(ctx CondContext) bool)` | 自定义条件逻辑                 |

### 🔍 示例：按属性控制注册

```go
gs.Provide(NewService).
    Condition(OnProperty("service.enabled"))
```

只有当配置文件中存在 `service.enabled=true` 时，`NewService` 才会注册。

### 🔁 组合条件

Go-Spring 支持组合多个条件，构建更复杂的判断逻辑：

| 方法          | 描述          |
|-------------|-------------|
| `Not(...)`  | 条件取反        |
| `And(...)`  | 所有条件都满足时成立  |
| `Or(...)`   | 任一条件满足即成立   |
| `None(...)` | 所有条件都不满足时成立 |

#### 示例：组合条件控制 Bean 激活

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

## 🔁 配置动态刷新（热更新）

Go-Spring 内置了轻量的配置热更新能力。通过 `gs.Dync[T]` 类型与 `gs.RefreshProperties()` 方法的组合，
可以实现应用在运行中动态响应配置变化，无需重启。

这非常适合微服务、配置中心、灰度发布场景，能够 **显著提升系统的可运维性与弹性**。

### 🌡 使用方式

#### 1. 使用 `gs.Dync[T]` 声明动态字段

通过泛型类型 `gs.Dync[T]` 包装字段，即可监听配置变化并自动更新：

```go
type Config struct {
    Version gs.Dync[string] `value:"${app.version}"`
}
```

> 调用时通过 `.Value()` 获取当前值，框架在配置变更时会自动更新该值。

#### 2. 调用 `gs.RefreshProperties()` 手动触发刷新

在配置发生变化后，调用此方法可以让所有动态字段立即更新：

```go
_ = gs.RefreshProperties()
```

### 💡 示例：实时版本号更新

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
    _ = gs.RefreshProperties()
    fmt.Fprintln(w, "Version updated!")
}
```

注册路由并启动应用：

```go
gs.Object(&App{})
gs.Provide(func(app *App) *http.ServeMux {
    http.Handle("/", app)
    http.HandleFunc("/refresh", RefreshVersion)
    return http.DefaultServeMux
})
gs.Run()
```

访问 `/` 会输出当前版本，访问 `/refresh` 后，再次访问 `/` 即可看到更新后的版本号。

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

用于协调服务“何时准备好”。你可以在 `ListenAndServe` 中等到主流程触发启动完成信号，然后正式对外提供服务。

### 🛠 示例：标准库 HTTP Server 接入

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
    go func() {
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