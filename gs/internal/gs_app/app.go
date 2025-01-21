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

// AppContext wraps the app's [gs.Context].
type AppContext struct {
	c gs.Context
}

// Unsafe returns the underlying [gs.Context]. It's an unsafe operation using the
// app's [gs.Context] in new goroutines, because the [gs.Context] may release all
// the bean definitions, so binding and injection is forbidden.
func (p *AppContext) Unsafe() gs.Context {
	return p.c
}

// Go runs the function in a goroutine. The function will receive the canceled
// signal when the application is shutting down.
func (p *AppContext) Go(fn func(ctx context.Context)) {
	p.c.Go(fn)
}

// AppRunner defines an interface using after all bean injections and before runs
// the servers, you can use it to run some background tasks.
type AppRunner interface {
	Run(ctx *AppContext)
}

// AppServer defines an interface using to start and stop the servers, such as http
// server, grpc server, thrift server, mq server, etc.
type AppServer interface {
	OnAppStart(ctx *AppContext)
	OnAppStop(ctx context.Context)
}

// App is the application.
type App struct {
	C gs.Container
	P *gs_conf.AppConfig

	exitChan chan struct{}

	Runners []AppRunner `autowire:"${spring.app.runners:=*?}"`
	Servers []AppServer `autowire:"${spring.app.servers:=*?}"`
}

// NewApp creates a new application.
func NewApp() *App {
	app := &App{
		C:        gs_core.New(),
		P:        gs_conf.NewAppConfig(),
		exitChan: make(chan struct{}),
	}
	app.C.Object(app)
	return app
}

// Run runs the application. It will listen the kill signal and the
// CTRL+C signal, so you can shut down the application gracefully.
// You can't use Stop but ShutDown to shut down the application.
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

// Start starts the application, init the properties, the ioc container,
// all beans injection and runs the runners and servers.
func (app *App) Start() error {

	// loads the app layered properties
	p, err := app.P.Refresh()
	if err != nil {
		return err
	}

	// refreshes the container's properties
	err = app.C.RefreshProperties(p)
	if err != nil {
		return err
	}

	// refreshes the container
	err = app.C.Refresh()
	if err != nil {
		return err
	}

	ctx := app.C.(gs.Context)

	// executes the app runners
	for _, r := range app.Runners {
		r.Run(&AppContext{ctx})
	}

	// executes the app servers
	for _, svr := range app.Servers {
		svr.OnAppStart(&AppContext{ctx})
	}

	app.C.ReleaseUnusedMemory()
	return nil
}

// Stop stops the application gracefully. It's used to stop the app
// started by Start.
func (app *App) Stop() {
	ctx := context.Background()
	for _, svr := range app.Servers {
		svr.OnAppStop(ctx)
	}
	app.C.Close()
}

// ShutDown shuts down the application gracefully. It's used to shut down
// the app started by Run.
func (app *App) ShutDown(msg ...string) {
	select {
	case <-app.exitChan:
		// do nothing, the exit chan has closed.
	default:
		close(app.exitChan)
	}
}
