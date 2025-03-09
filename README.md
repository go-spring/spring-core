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
下面是一个最简单的示例，展示了如何注册一个 Bean、绑定属性以及启动应用：
```
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/util/sysconf"
	"github.com/go-spring/spring-core/util/syslog"
)

func init() {
	gs.Object(&Service{})     // Register the Service object.
	gs.Object(&Job{}).AsJob() // Register the Job object and mark it as a scheduled job.
}

type Service struct{}

// Echo prints a formatted log message using the syslog package.
func (s *Service) Echo(format string, a ...any) {
	syslog.Infof(fmt.Sprintf(format, a...))
}

// Job struct represents a scheduled task that depends on the Service.
type Job struct {
	Service *Service `autowire:""`                // Automatically inject the Service dependency.
	AppName string   `value:"${spring.app.name}"` // Read the application name from the configuration.
}

// Run method is executed when the Job is triggered.
func (j *Job) Run(ctx context.Context) error {
	time.Sleep(time.Second * 2)
	j.Service.Echo("app '%s' will exit", j.AppName)
	gs.ShutDown() // Shut down the application.
	return nil
}

func main() {
	// Set the application name in the configuration.
	_ = sysconf.Set("spring.app.name", "test")

	// Start the Go-Spring application. If it fails, log the error.
	if err := gs.Run(); err != nil {
		syslog.Errorf("app run failed: %s", err.Error())
	}
}
```




