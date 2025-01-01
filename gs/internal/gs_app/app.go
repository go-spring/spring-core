/*
 * Copyright 2012-2024 the original author or authors.
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

package gs_app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_ctx"
)

const (
	Version = "go-spring@v1.1.3"
	Website = "https://go-spring.com/"
)

type AppRunner interface {
	Run(ctx gs.Context)
}

type AppServer interface {
	OnAppStart(ctx gs.Context)     // 应用启动的事件
	OnAppStop(ctx context.Context) // 应用停止的事件
}

// App 应用
type App struct {
	banner string

	b *Boot
	c *gs_ctx.Container
	p *gs_conf.Configuration

	exitChan chan struct{}

	Servers []AppServer `autowire:"${spring.app.servers:=*?}"`
}

// NewApp application 的构造函数
func NewApp() *App {
	app := &App{
		banner:   DefaultBanner,
		c:        gs_ctx.New(),
		p:        gs_conf.NewConfiguration(),
		exitChan: make(chan struct{}),
	}
	app.Object(app)
	return app
}

func (app *App) Run() error {
	if _, err := app.Start(); err != nil {
		return err
	}
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		sig := <-ch
		app.ShutDown(fmt.Sprintf("signal %v", sig))
	}()
	<-app.exitChan
	app.Stop()
	return nil
}

func (app *App) Start() (gs.Context, error) {

	app.showBanner()

	if app.b != nil {
		err := app.b.run()
		if err != nil {
			return nil, err
		}
	}

	p, err := app.p.Refresh()
	if err != nil {
		return nil, err
	}

	err = app.c.RefreshProperties(p)
	if err != nil {
		return nil, err
	}

	err = app.c.Refresh()
	if err != nil {
		return nil, err
	}

	var runners []AppRunner
	err = app.c.Get(&runners, "${spring.app.runners:=*?}")
	if err != nil {
		return nil, err
	}

	// 执行命令行启动器
	for _, r := range runners {
		r.Run(app.c)
	}

	// 通知应用启动事件
	for _, event := range app.Servers {
		event.OnAppStart(app.c)
	}

	// 通知应用停止事件
	app.c.Go(func(ctx context.Context) {
		<-ctx.Done()
		for _, event := range app.Servers {
			event.OnAppStop(context.Background())
		}
	})

	return app.c, nil
}

func (app *App) Stop() {
	app.c.Close()
}

// ShutDown 关闭执行器
func (app *App) ShutDown(msg ...string) {
	select {
	case <-app.exitChan:
		// chan 已关闭，无需再次关闭。
	default:
		close(app.exitChan)
	}
}

// Boot 返回 *bootstrap 对象。
func (app *App) Boot() *Boot {
	if app.b == nil {
		app.b = newBoot()
	}
	return app.b
}

// Object 参考 Container.Object 的解释。
func (app *App) Object(i interface{}) *gs.BeanDefinition {
	return app.c.Accept(gs_ctx.NewBean(reflect.ValueOf(i)))
}

// Provide 参考 Container.Provide 的解释。
func (app *App) Provide(ctor interface{}, args ...gs.Arg) *gs.BeanDefinition {
	return app.c.Accept(gs_ctx.NewBean(ctor, args...))
}

// Accept 参考 Container.Accept 的解释。
func (app *App) Accept(b *gs.BeanDefinition) *gs.BeanDefinition {
	return app.c.Accept(b)
}

func (app *App) Group(fn gs.GroupFunc) {
	app.c.Group(fn)
}
