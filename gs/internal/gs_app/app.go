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
	"github.com/go-spring/spring-core/gs/internal/gs_core"
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
	c gs.Container
	p *gs_conf.AppConfig

	exitChan chan struct{}

	Runners []AppRunner `autowire:"${spring.app.runners:=*?}"`
	Servers []AppServer `autowire:"${spring.app.servers:=*?}"`
}

// NewApp application 的构造函数
func NewApp() *App {
	app := &App{
		c:        gs_core.New(),
		p:        gs_conf.NewAppConfig(),
		exitChan: make(chan struct{}),
	}
	app.Object(app)
	app.Object(app.p)
	return app
}

func (app *App) Config() *gs_conf.AppConfig {
	return app.p
}

// Object 参考 Container.Object 的解释。
func (app *App) Object(i interface{}) *gs.BeanRegistration {
	b := gs_core.NewBean(reflect.ValueOf(i))
	app.c.Accept(b)
	return &gs.BeanRegistration{B: b}
}

// Provide 参考 Container.Provide 的解释。
func (app *App) Provide(ctor interface{}, args ...gs.Arg) *gs.BeanRegistration {
	b := gs_core.NewBean(ctor, args...)
	app.c.Accept(b)
	return &gs.BeanRegistration{B: b}
}

// Accept 参考 Container.Accept 的解释。
func (app *App) Accept(b *gs.BeanDefinition) {
	app.c.Accept(b)
}

func (app *App) Group(fn func(p gs.Properties) ([]*gs.BeanDefinition, error)) {
	app.c.Group(fn)
}

func (app *App) Run() error {
	if err := app.Start(); err != nil {
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

func (app *App) Start() error {

	p, err := app.p.Refresh()
	if err != nil {
		return err
	}

	err = app.c.RefreshProperties(p)
	if err != nil {
		return err
	}

	err = app.c.Refresh()
	if err != nil {
		return err
	}

	ctx := app.c.(gs.Context)

	// 执行命令行启动器
	for _, r := range app.Runners {
		r.Run(ctx)
	}

	// 通知应用启动事件
	for _, svr := range app.Servers {
		svr.OnAppStart(ctx)
	}

	app.c.SimplifyMemory()
	return nil
}

func (app *App) Stop() {
	for _, svr := range app.Servers {
		svr.OnAppStop(context.Background())
	}
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
