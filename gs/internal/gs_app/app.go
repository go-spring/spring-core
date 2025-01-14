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
	"syscall"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_core"
)

type AppRunner interface {
	Run(ctx *AppContext)
}

type AppServer interface {
	OnAppStart(ctx *AppContext)    // 应用启动的事件
	OnAppStop(ctx context.Context) // 应用停止的事件
}

type AppContext struct {
	c gs.Context
}

func (p *AppContext) Unsafe() gs.Context {
	return p.c
}

func (p *AppContext) Go(fn func(ctx context.Context)) {
	p.c.Go(fn)
}

// App 应用
type App struct {
	C gs.Container
	P *gs_conf.AppConfig

	exitChan chan struct{}

	Runners []AppRunner `autowire:"${spring.app.runners:=*?}"`
	Servers []AppServer `autowire:"${spring.app.servers:=*?}"`
}

// NewApp application 的构造函数
func NewApp() *App {
	app := &App{
		C:        gs_core.New(),
		P:        gs_conf.NewAppConfig(),
		exitChan: make(chan struct{}),
	}
	app.C.Object(app)
	return app
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

	p, err := app.P.Refresh()
	if err != nil {
		return err
	}

	err = app.C.RefreshProperties(p)
	if err != nil {
		return err
	}

	err = app.C.Refresh()
	if err != nil {
		return err
	}

	ctx := app.C.(gs.Context)

	// 执行命令行启动器
	for _, r := range app.Runners {
		r.Run(&AppContext{ctx})
	}

	// 通知应用启动事件
	for _, svr := range app.Servers {
		svr.OnAppStart(&AppContext{ctx})
	}

	app.C.ReleaseUnusedMemory()
	return nil
}

func (app *App) Stop() {
	for _, svr := range app.Servers {
		svr.OnAppStop(context.Background())
	}
	app.C.Close()
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
