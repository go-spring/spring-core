# spring-core

# 介绍

Go-Spring 是为 Go 开发者打造的轻量级微服务框架，灵感来源于 Java 的 Spring 和 Spring Boot。
它旨在降低开发门槛、提高项目结构的清晰度和可维护性。主要特点包括：

- 易用性：通过注解标签和链式调用注册 Bean 与配置，降低手写样板代码。
- 扩展性：支持动态刷新属性和 Bean，在运行时调整配置而无需重启应用。
- 微服务支持：内置启动框架及丰富的扩展接口，可快速构建多种微服务应用。
- 测试友好：提供丰富的单元测试工具和 mock 能力，保证代码质量。

# 快速开始

### 安装

通过 Go Modules 获取最新版本：

```
go get github.com/go-spring/spring-core@develop
```

### 最小示例

下面是一个最简单的示例，展示了如何注册一个 Bean、绑定属性、动态属性以及启动应用：

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/util/sysconf"
	"github.com/go-spring/spring-core/util/syslog"
)

func init() {
	// Register the Service struct as a bean.
	gs.Object(&Service{})

	// Provide a [*http.ServeMux] as a bean.
	gs.Provide(func(s *Service) *http.ServeMux {
		http.HandleFunc("/echo", s.Echo)
		http.HandleFunc("/refresh", s.Refresh)
		return http.DefaultServeMux
	})
}

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
	_ = sysconf.Set("refresh-time", time.Now().Format(timeLayout))
	_ = gs.RefreshProperties()
	_, _ = w.Write([]byte("OK!"))
}

func main() {
	_ = sysconf.Set("start-time", time.Now().Format(timeLayout))
	_ = sysconf.Set("refresh-time", time.Now().Format(timeLayout))

	// Start the Go-Spring application. If it fails, log the error.
	if err := gs.Run(); err != nil {
		syslog.Errorf("app run failed: %s", err.Error())
	}
}
```

当你运行这个程序时，它将启动一个 HTTP 服务器，并注册两个处理器：一个处理 "/echo" 请求，
返回当前时间和刷新时间；另一个处理 "/refresh" 请求，用于刷新配置并返回 "OK!"。

运行这个程序，你可以访问 "/echo" 和 "/refresh"，并观察到它们返回的当前时间和刷新时间。

```shell
➜ ~ curl http://127.0.0.1:9090/echo
start-time: 2025-03-14 13:32:51.608 +0800 CST refresh-time: 2025-03-14 13:32:51.608 +0800 CST%
➜ ~ curl http://127.0.0.1:9090/refresh
OK!%
➜ ~ curl http://127.0.0.1:9090/echo
start-time: 2025-03-14 13:32:51.608 +0800 CST refresh-time: 2025-03-14 13:33:02.936 +0800 CST%
➜ ~ curl http://127.0.0.1:9090/refresh
OK!%
➜ ~ curl http://127.0.0.1:9090/echo
start-time: 2025-03-14 13:32:51.608 +0800 CST refresh-time: 2025-03-14 13:33:08.88 +0800 CST%
```
